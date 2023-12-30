package dgotorrent

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/Dizzrt/dgo-torrent/bencode"
	"github.com/Dizzrt/dgo-torrent/config"
	"github.com/Dizzrt/dgo-torrent/dlog"
)

var (
	ErrInvalidTrackerResp = errors.New("invalid tracker resp")
)

type TrackerResp struct {
	Complete    int64  `bencode:"complete"`
	Downloaded  int64  `bencode:"downloaded"`
	Incomplete  int64  `bencode:"incomplete"`
	Interval    int64  `bencode:"interval"`
	MinInterval int64  `bencode:"min interval"`
	Peers       []byte `bencode:"peers"`
}

func parseTrackerResp(resp map[string]any) (TrackerResp, error) {
	ret := TrackerResp{}

	if v, ok := resp["complete"]; ok {
		value, ok := v.(int64)
		if ok {
			ret.Complete = value
		}
	} else {
		ret.Complete = 0
	}

	if v, ok := resp["downloaded"]; ok {
		value, ok := v.(int64)
		if ok {
			ret.Downloaded = value
		}
	} else {
		ret.Downloaded = 0
	}

	if v, ok := resp["incomplete"]; ok {
		value, ok := v.(int64)
		if ok {
			ret.Incomplete = value
		}
	} else {
		ret.Incomplete = 0
	}

	if v, ok := resp["interval"]; ok {
		value, ok := v.(int64)
		if ok {
			ret.Interval = value
		}
	} else {
		ret.Interval = 0
	}

	if v, ok := resp["min interval"]; ok {
		value, ok := v.(int64)
		if ok {
			ret.MinInterval = value
		}
	} else {
		ret.MinInterval = 0
	}

	if v, ok := resp["peers"]; ok {
		value, ok := v.(string)
		if ok {
			ret.Peers = []byte(value)
		}
	} else {
		ret.Peers = nil
	}

	return ret, nil
}

func (tf *TorrentFile) buildHttpTrackerUrl(tracker string) (string, error) {
	base, err := url.Parse(tracker)
	if err != nil {
		return "", err
	}

	peerID := config.Instance().GetPeerID()
	params := url.Values{
		"info_hash":  []string{string(tf.Info.Hash[:])},
		"peer_id":    []string{peerID},
		"port":       []string{strconv.Itoa(6666)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(int(tf.Info.Length))},
	}

	base.RawQuery = params.Encode()
	return base.String(), nil
}

func (tf *TorrentFile) requestHttpTrackers(httpTrackers []string) ([]TrackerResp, error) {
	var wg sync.WaitGroup
	respList := make([]TrackerResp, 0)
	for _, tracker := range httpTrackers {
		wg.Add(1)
		go func(tracker string) {
			defer wg.Done()

			client := &http.Client{Timeout: 15 * time.Second}
			url, err := tf.buildHttpTrackerUrl(tracker)
			if err != nil {
				dlog.Errorf("Failed to build http tracker url with error: %v", err)
				return
			}

			clientResp, err := client.Get(url)
			if err != nil {
				dlog.Errorf("Failed to request http tracker with error: %v", err)
				return
			}
			defer clientResp.Body.Close()

			res, err := bencode.Unmarshal(clientResp.Body)
			if err != nil {
				dlog.Errorf("Failed to unmarshal http tracker res with error: %v", err)
				return
			}

			if v, ok := res.(map[string]any); ok {
				resp, err := parseTrackerResp(v)
				if err != nil {
					dlog.Errorf("Failed to parse http tracker resp with error: %v", err)
					return
				}

				respList = append(respList, resp)
			}
		}(tracker)
	}

	wg.Wait()
	return respList, nil
}

func resolveUDPTracker(tracker string) (*net.UDPAddr, error) {
	parsedUrl, err := url.Parse(tracker)
	if err != nil {
		dlog.Errorf("Failed to parse UDP url, error: %v", err)
		return nil, err
	}

	tracker = parsedUrl.Host
	addr, err := net.ResolveUDPAddr("udp", tracker)
	if err != nil {
		dlog.Errorf("Failed to resolve UDP addr, error: %v", err)
		return nil, err
	}

	return addr, nil
}

func connectUDPTracker(conn *net.UDPConn, transactionID uint32) (uint64, error) {
	data := make([]byte, 16)
	binary.BigEndian.PutUint64(data[0:8], 0x41727101980) // protocol_id - magic number
	binary.BigEndian.PutUint32(data[8:12], 0)            // action 0: connect
	binary.BigEndian.PutUint32(data[12:16], transactionID)

	_, err := conn.Write(data)
	if err != nil {
		dlog.Errorf("Failed to connect udp server with error: %v", err)
		return 0, err
	}

	conn.SetDeadline(time.Now().Add(15 * time.Second))
	buf := make([]byte, 16)
	_, err = conn.Read(buf)
	if err != nil {
		dlog.Errorf("Faile to read udp message, error: %v", err)
		return 0, err
	}

	// action := binary.BigEndian.Uint32(buf[0:4])
	// transaction_id := binary.BigEndian.Uint32(buf[4:8])
	connectionID := binary.BigEndian.Uint64(buf[8:16])
	return connectionID, nil
}

func (tf *TorrentFile) buildUDPTrackerPackage(connectionID uint64, transactionID uint32) []byte {
	data := make([]byte, 98)
	binary.BigEndian.PutUint64(data[0:8], connectionID)
	binary.BigEndian.PutUint32(data[8:12], 1)
	binary.BigEndian.PutUint32(data[12:16], transactionID)
	copy(data[16:36], tf.Info.Hash[:])
	copy(data[36:56], config.Instance().GetPeerID())
	binary.BigEndian.PutUint64(data[56:64], 0)
	binary.BigEndian.PutUint64(data[64:72], uint64(tf.Info.Length))
	binary.BigEndian.PutUint64(data[72:80], 0)
	binary.BigEndian.PutUint32(data[80:84], 2)
	binary.BigEndian.PutUint32(data[84:88], 0)
	binary.BigEndian.PutUint32(data[88:92], 0)
	binary.BigEndian.PutUint32(data[92:96], 0xffffffff)
	binary.BigEndian.PutUint16(data[96:98], 6666)

	return data
}

func (tf *TorrentFile) requestUdpTrackers(udpTrackers []string) ([]TrackerResp, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var wg sync.WaitGroup
	respList := make([]TrackerResp, 0)
	for _, tracker := range udpTrackers {
		wg.Add(1)
		go func(tracker string) {
			defer wg.Done()

			addr, err := resolveUDPTracker(tracker)
			if err != nil {
				return
			}

			conn, err := net.DialUDP("udp", nil, addr)
			if err != nil {
				dlog.Errorf("Failed to dial udp, error: %v", err)
				return
			}
			defer conn.Close()

			cid, err := connectUDPTracker(conn, r.Uint32())
			if err != nil {
				return
			}

			data := tf.buildUDPTrackerPackage(cid, r.Uint32())
			_, err = conn.Write(data)
			if err != nil {
				dlog.Errorf("Failed to request peers, error: %v", err)
				return
			}
			conn.SetDeadline(time.Now().Add(15 * time.Second))

			buf := make([]byte, 3092)
			oob := make([]byte, 1024)
			n, _, flags, _, err := conn.ReadMsgUDP(buf, oob)
			if err != nil {
				dlog.Errorf("Failed to read peers, error: %v", err)
				return
			}

			dlog.Info(flags)
			// if flags&syscall.MSG_TRUNC != 0 {
			// 	dlog.Warn("Truncated peers")
			// }

			x := (n - 20) / 6 * 6
			if x <= 0 {
				return
			}

			resp := TrackerResp{
				Peers: buf[20 : 20+x],
			}
			respList = append(respList, resp)
		}(tracker)
	}

	wg.Wait()
	return respList, nil
}

func (tf *TorrentFile) RequestTrackers() ([]TrackerResp, error) {
	udpTrackers := make([]string, 0)
	httpTrackers := make([]string, 0)

	announceList := tf.AnnounceList
	announceList = append(announceList, tf.Announce)

	for _, t := range announceList {
		parsedURL, err := url.Parse(t)
		if err != nil {
			dlog.Warnf("Failed to Parse tracker: %s, error: %v", t, err)
			continue
		}

		switch parsedURL.Scheme {
		case "http":
			httpTrackers = append(httpTrackers, t)
		case "https":
			// TODO
		case "udp":
			udpTrackers = append(udpTrackers, t)
		default:
			dlog.Infof("Unrecognized tracker protocal: %s", t)
		}
	}

	respList := make([]TrackerResp, 0)
	if len(httpTrackers) > 0 {
		httpRespList, err := tf.requestHttpTrackers(httpTrackers)
		if err != nil {
			return nil, err
		}

		respList = append(respList, httpRespList...)
	}

	if len(udpTrackers) > 0 {
		udpRespList, err := tf.requestUdpTrackers(udpTrackers)
		if err != nil {
			return nil, err
		}

		respList = append(respList, udpRespList...)
	}

	return respList, nil
}
