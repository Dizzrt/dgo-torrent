package torrent

import (
	"encoding/binary"
	"net"

	"github.com/Dizzrt/dgo-torrent/model"
)

const (
	IpLen   int = 4
	PortLen int = 2
	PeerLen int = IpLen + PortLen
)

type Peer struct {
	Ip   net.IP
	Port uint16
}

func GetPeers(tf *model.Torrent) []Peer {
	resp, err := getTrackerResp(tf)
	if err != nil {
		panic(err)
	}

	var peersBytes []byte
	rawPeers, ok := resp["peers"].(string)
	if ok {
		peersBytes = []byte(rawPeers)
	}

	peersBytesLen := len(peersBytes)
	peersCnt := peersBytesLen / PeerLen
	if peersBytesLen%PeerLen != 0 {
		// TODO received malformed peers
		return nil
	}

	peers := make([]Peer, peersCnt)
	for i := 0; i < peersCnt; i++ {
		offset := i * PeerLen
		peers[i].Ip = net.IP(peersBytes[offset : offset+IpLen])
		peers[i].Port = binary.BigEndian.Uint16(peersBytes[offset+IpLen : offset+PeerLen])
	}

	return peers
}
