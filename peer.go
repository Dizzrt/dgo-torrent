package dgotorrent

import (
	"encoding/binary"
	"fmt"
	"net"
)

// ip_len:4 port_len:2
const IP_LEN = 4
const PORT_LEN = 4
const PEER_LEN = IP_LEN + PORT_LEN

const PEER_ID_LEN = 20

type Peer struct {
	IP   net.IP
	Port uint16
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
