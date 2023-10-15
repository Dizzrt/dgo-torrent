package dgotorrent

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"
)

// ip_len:4 port_len:2
const IP_LEN = 4
const PORT_LEN = 2
const PEER_LEN = IP_LEN + PORT_LEN

const PEER_ID_LEN = 20

type PeerMsgTyep uint8

const (
	// enable upload
	PEER_MSG_TYPE_CHOKE PeerMsgTyep = iota
	// disbale upload
	PEER_MSG_TYPE_UNCHOKE
	// enable download
	PEER_MSG_TYPE_INTERESTED
	// disable download
	PEER_MSG_TYPE_NOT_INTEREST
	// update bitfieled
	PEER_MSG_TYPE_HAVE
	// bitfieled
	PEER_MSG_TYPE_BITFIELED
	// download request
	PEER_MSG_TYPE_REQUEST
	// pirece data
	PEER_MSG_TYPE_PIECE
	PEER_MSG_TYPE_CANCEL
	PEER_MSG_TYPE_KEEP_ALIVE
)

type Peer struct {
	IP   net.IP
	Port uint16
}

type PeerMsg struct {
	Type    PeerMsgTyep
	Payload []byte
}

type PeerConn struct {
	net.Conn
	Choked   bool
	Fieled   Bitfield
	peer     Peer
	peerID   [PEER_ID_LEN]byte
	infoHash [INFO_HASH_LEN]byte
}

func (tf *TorrentFile) FindPeers() ([]Peer, error) {
	trackerRespList, err := tf.RequestTrackers()
	if err != nil {
		return nil, err
	}

	peers := make([]Peer, 0)
	for _, tr := range trackerRespList {
		rawPeers := tr.Peers

		rawPeerLen := len(rawPeers)
		if rawPeerLen%PEER_LEN != 0 {
			fmt.Println("received malformed peers")
			continue
		}

		peerCount := rawPeerLen / PEER_LEN

		peerList := make([]Peer, peerCount)
		for i := 0; i < peerCount; i++ {
			offset := i * PEER_LEN
			peerList[i].IP = net.IP(rawPeers[offset : offset+IP_LEN])
			peerList[i].Port = binary.BigEndian.Uint16(rawPeers[offset+IP_LEN : offset+PEER_LEN])
		}

		peers = append(peers, peerList...)
	}

	return peers, nil
}

func handshake(conn net.Conn, infoHash [INFO_HASH_LEN]byte, peerID [PEER_ID_LEN]byte) error {
	conn.SetDeadline(time.Now().Add(15 * time.Second))
	defer conn.SetDeadline(time.Time{})

	msg := struct {
		PreStr   string
		InfoHash [INFO_HASH_LEN]byte
		PeerID   [PEER_ID_LEN]byte
	}{"BitTorrent protocol", infoHash, peerID}

	buf := make([]byte, len(msg.PreStr)+49)
	buf[0] = byte(len(msg.PreStr))

	cur := 1
	cur += copy(buf[cur:], []byte(msg.PreStr))
	cur += copy(buf[cur:], make([]byte, 8))
	cur += copy(buf[cur:], msg.InfoHash[:])
	cur += copy(buf[cur:], msg.PeerID[:])
	_, err := conn.Write(buf)

	return err
}

func (c *PeerConn) ReadMsg() (*PeerMsg, error) {
	lenBuf := make([]byte, 4)
	_, err := io.ReadFull(c, lenBuf)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lenBuf)
	if length == 0 {
		return nil, nil
	}

	msgBuf := make([]byte, length)
	_, err = io.ReadFull(c, msgBuf)
	if err != nil {
		return nil, err
	}

	return &PeerMsg{
		Type:    PeerMsgTyep(msgBuf[0]),
		Payload: msgBuf[1:],
	}, nil
}

func fillBitfield(c *PeerConn) error {
	c.SetDeadline(time.Now().Add(15 * time.Second))
	defer c.SetDeadline(time.Time{})

	msg, err := c.ReadMsg()
	if err != nil {
		return err
	}

	if msg == nil {
		return fmt.Errorf("expected bitfield")
	}

	if msg.Type != PEER_MSG_TYPE_BITFIELED {
		return fmt.Errorf("type error")
	}

	c.Fieled = msg.Payload
	return nil
}

func NewConn(peer Peer, infoHash [INFO_HASH_LEN]byte, peerID [PEER_ID_LEN]byte) (*PeerConn, error) {
	out, _ := os.OpenFile("test/out/torrent_peer_conn_test_result.json", os.O_WRONLY|os.O_CREATE, 0666)
	defer out.Close()

	addr := net.JoinHostPort(peer.IP.String(), strconv.Itoa(int(peer.Port)))
	conn, err := net.DialTimeout("tcp", addr, 15*time.Second)
	if err != nil {
		return nil, err
	}

	err = handshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		out.Write([]byte(fmt.Sprintf("handshake faild with IP: %v\n", peer.IP)))
		return nil, err
	}

	c := &PeerConn{
		Conn:     conn,
		Choked:   true,
		peer:     peer,
		peerID:   peerID,
		infoHash: infoHash,
	}

	err = fillBitfield(c)
	if err != nil {
		out.Write([]byte(fmt.Sprintf("fill bitfield faild with IP: %v\n", peer.IP)))
		return nil, err
	}

	return c, nil
}
