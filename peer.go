package torrent

import (
	"net"
	"strconv"
	"time"
)

const (
	SHALEN int = 20
	IDLEN  int = 20
)

type PeerID_t [20]byte

type PeerMsgType uint8

const (
	PeerMsgChoke PeerMsgType = iota
	PeerMsgUnchoke
	PeerMsgInterested
	PeerMsgNotInterest
	PeerMsgHave
	PeerMsgBitfield
	PeerMsgRequest
	PeerMsgPiece
	PeerMsgCancel
)

type PeerMsg struct {
	Type    PeerMsgType
	Payload []byte
}

type Peer struct {
	IP   net.IP
	Port uint16
}

type PeerConn struct {
	net.Conn
	Choked  bool
	Field   Bitfield
	peer    Peer
	peerID  [IDLEN]byte
	infoSHA [SHALEN]byte
}

func NewPeerConn(peer Peer, infoSHA [SHALEN]byte, peerID [IDLEN]byte) (*PeerConn, error) {
	addr := net.JoinHostPort(peer.IP.String(), strconv.Itoa(int(peer.Port)))
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, err
	}

	// TODO handshake

	c := &PeerConn{
		Conn:    conn,
		Choked:  true,
		peer:    peer,
		peerID:  peerID,
		infoSHA: infoSHA,
	}

	// TODO fill bitfield

	return c, nil
}
