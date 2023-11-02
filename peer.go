package dgotorrent

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
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
	PEER_MSG_INVALID
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
	Choked    bool
	PiecesMap Bitfield
	peer      Peer
	peerID    string
	infoHash  [INFO_HASH_LEN]byte
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

func (c *PeerConn) WriteMsg(msg *PeerMsg) (int, error) {
	if msg == nil {
		return 0, nil
	}

	if msg.Type >= PEER_MSG_INVALID {
		return 0, fmt.Errorf("peer msg type out of range")
	}

	PayloadLength := uint32(len(msg.Payload) + 1)
	buf := make([]byte, 4+PayloadLength)

	binary.BigEndian.PutUint32(buf[0:4], PayloadLength)
	buf[4] = byte(msg.Type)

	copy(buf[4+1:], msg.Payload)
	return c.Write(buf)
}

// region new peer conn

func NewConn(peer Peer, infoHash [INFO_HASH_LEN]byte, peerID string) (*PeerConn, error) {
	addr := net.JoinHostPort(peer.IP.String(), strconv.Itoa(int(peer.Port)))
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, err
	}

	err = handshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = checkHandshakeMsg(conn, infoHash)
	if err != nil {
		conn.Close()
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
		return nil, err
	}

	return c, nil
}

func handshake(conn net.Conn, infoHash [INFO_HASH_LEN]byte, peerID string) error {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	msg := struct {
		PreStr   string
		InfoHash [INFO_HASH_LEN]byte
		PeerID   string
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

func checkHandshakeMsg(r io.Reader, targetInfoHash [INFO_HASH_LEN]byte) error {
	lenBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lenBuf)
	if err != nil {
		return err
	}

	preLen := int(lenBuf[0])
	if preLen == 0 {
		err := fmt.Errorf("prelen can not be 0")
		return err
	}

	msgBuf := make([]byte, 48+preLen)
	_, err = io.ReadFull(r, msgBuf)
	if err != nil {
		return err
	}

	// var peerID [PEER_ID_LEN]byte
	var infoHash [INFO_HASH_LEN]byte

	copy(infoHash[:], msgBuf[preLen+8:preLen+8+INFO_HASH_LEN])
	// copy(peerID[:], msgBuf[preLen+8+INFO_HASH_LEN:])

	// preStr := string(msgBuf[0:preLen])

	if !bytes.Equal(infoHash[:], targetInfoHash[:]) {
		return fmt.Errorf("handshake msg error: " + string(infoHash[:]))
	}

	return nil
}

func fillBitfield(c *PeerConn) error {
	c.SetDeadline(time.Now().Add(5 * time.Second))
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

	c.PiecesMap = msg.Payload
	return nil
}

// endregion

func CopyPieceData(index int, buf []byte, msg *PeerMsg) (int, error) {
	if msg.Type != PEER_MSG_TYPE_PIECE {
		return 0, fmt.Errorf("expected PEER_MSG_TYPE_PIECE[%d], got %d", PEER_MSG_TYPE_PIECE, msg.Type)
	}

	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("payload too short. %d < 8", len(msg.Payload))
	}

	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != index {
		return 0, fmt.Errorf("expected index %d, got %d", index, parsedIndex)
	}

	offset := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if offset >= len(buf) {
		return 0, fmt.Errorf("offset too high. %d >= %d", offset, len(buf))
	}

	data := msg.Payload[8:]
	if offset+len(data) > len(buf) {
		return 0, fmt.Errorf("data too large [%d] for offset %d with length %d", len(data), offset, len(buf))
	}

	copy(buf[offset:], data)
	return len(data), nil
}

func GetHaveIndex(msg *PeerMsg) (int, error) {
	if msg.Type != PEER_MSG_TYPE_HAVE {
		return 0, fmt.Errorf("expected PEER_MSG_TYPE_HAVE[%d], got %d", PEER_MSG_TYPE_HAVE, msg.Type)
	}

	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("expected payload length 4, got length %d", len(msg.Payload))
	}

	index := int(binary.BigEndian.Uint32(msg.Payload))
	return index, nil
}

func NewRequestMsg(index, offset, length int) *PeerMsg {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(offset))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))

	return &PeerMsg{
		Type:    PEER_MSG_TYPE_REQUEST,
		Payload: payload,
	}
}
