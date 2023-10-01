package dgotorrent

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/Dizzrt/dgo-torrent/bencode"
)

var (
	ErrInvalidTrackerResp = errors.New("invalid tracker resp")
)

type TrackerResp struct {
	Complete    int64  `bencode:"complete"`
	Downloaded  int64  `bencode:"downloaded"`
	Incomplete  int64  `incomplete:"incomplete"`
	Interval    int64  `incomplete:"interval"`
	MinInterval int64  `incomplete:"min interval"`
	Peers       string `incomplete:"peers"`
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
			ret.Peers = value
		}
	} else {
		ret.Peers = ""
	}

	return ret, nil
}

func (tf *TorrentFile) buildHttpTrackerUrl(tracker string) (string, error) {
	base, err := url.Parse(tracker)
	if err != nil {
		return "", err
	}

	var peerId [20]byte
	_, _ = rand.Read(peerId[:])
	params := url.Values{
		"info_hash":  []string{string(tf.Info.Hash[:])},
		"peer_id":    []string{string(peerId[:])},
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
				return
			}

			clientResp, err := client.Get(url)
			if err != nil {
				return
			}
			defer clientResp.Body.Close()

			res, err := bencode.Unmarshal(clientResp.Body)
			if err != nil {
				fmt.Println(err)
			}

			if v, ok := res.(map[string]any); ok {
				resp, err := parseTrackerResp(v)
				if err != nil {
					fmt.Println(err)
					// XXX do something
				}

				respList = append(respList, resp)
			}
		}(tracker)
	}

	wg.Wait()
	return respList, nil
}

// func requestUdpTrackers(udpTrackers []string) ([]TrackerResp, error) {

// 	return nil, nil
// }

func (tf *TorrentFile) RequestTrackers() ([]TrackerResp, error) {
	// udpTrackers := make([]string, 0)
	httpTrackers := make([]string, 0)

	announceList := tf.AnnounceList
	announceList = append(announceList, tf.Announce)

	for _, t := range announceList {
		parsedURL, err := url.Parse(t)
		if err != nil {
			continue
			// XXX do something
		}

		switch parsedURL.Scheme {
		case "http":
			httpTrackers = append(httpTrackers, t)
		case "https":
			// TODO
		case "udp":
			// udpTrackers = append(udpTrackers, t)
		default:
			// XXX do something
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

	fmt.Printf("[XXXXX]: %+v\n", respList)

	// if len(udpTrackers) > 0 {
	// 	udpRespList, err := requestUdpTrackers(udpTrackers)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	respList = append(respList, udpRespList...)
	// }

	return respList, nil
}
