package dgotorrent

import (
	"bytes"
	"crypto/sha1"
	"time"

	"github.com/Dizzrt/dgo-torrent/dlog"
)

type TorrentTask struct {
	PeerID      string
	Peers       []Peer
	InfoHash    [INFO_HASH_LEN]byte
	FileName    string
	FileLength  int
	PieceLength int
	PiecesHash  [][20]byte
}

type TorrentWork struct {
	index  int
	hash   [20]byte
	length int
}

type TorrentWorkState struct {
	index      int
	conn       *PeerConn
	requested  int
	downloaded int
	backlog    int
	data       []byte
}

type TorrentWorkResult struct {
	index int
	data  []byte
}

const BLOCKSIZE = 16384
const MAXBACKLOG = 5

func downloadPiece(conn *PeerConn, work *TorrentWork) (*TorrentWorkResult, error) {
	state := &TorrentWorkState{
		index:   work.index,
		conn:    conn,
		data:    make([]byte, work.length),
		backlog: 0,
	}

	conn.SetDeadline(time.Now().Add(15 * time.Second))
	defer conn.SetDeadline(time.Time{})

	for state.downloaded < work.length {
		if !conn.Choked {
			for state.backlog < MAXBACKLOG && state.requested < work.length {
				length := BLOCKSIZE
				if work.length-state.requested < length {
					length = work.length - state.requested
				}

				msg := NewRequestMsg(state.index, state.requested, length)
				_, err := state.conn.WriteMsg(msg)
				if err != nil {
					return nil, err
				}

				state.backlog++
				state.requested += length
			}
		}

		err := state.handleMsg()
		if err != nil {
			return nil, err
		}
	}

	return &TorrentWorkResult{
		index: state.index,
		data:  state.data,
	}, nil
}

func checkPiece(work *TorrentWork, res *TorrentWorkResult) bool {
	hash := sha1.Sum(res.data)
	if !bytes.Equal(work.hash[:], hash[:]) {
		dlog.Infof("check integrity failed, index: %v", res.index)
		return false
	}

	return true
}

func (t *TorrentTask) getPieceBounds(index int) (begin, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength

	if end > t.FileLength {
		end = t.FileLength
	}

	return
}

func (t *TorrentTask) peerRoutine(peer Peer, works chan *TorrentWork, results chan *TorrentWorkResult) {
	conn, err := NewConn(peer, t.InfoHash, t.PeerID)
	if err != nil {
		dlog.Infof("fail to connect peer: %s:%d", peer.IP.String(), peer.Port)
		return
	}
	defer conn.Close()

	conn.WriteMsg(&PeerMsg{
		Type:    PEER_MSG_TYPE_INTERESTED,
		Payload: nil,
	})

	for work := range works {
		if !conn.PiecesMap.Test(work.index) {
			works <- work
			continue
		}

		res, err := downloadPiece(conn, work)
		if err != nil {
			works <- work
			dlog.Errorf("fail to download piece: %v", err)
			return
		}

		if !checkPiece(work, res) {
			works <- work
			continue
		}

		results <- res
	}

}

func (state *TorrentWorkState) handleMsg() error {
	msg, err := state.conn.ReadMsg()
	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}

	switch msg.Type {
	case PEER_MSG_TYPE_CHOKE:
		state.conn.Choked = true
	case PEER_MSG_TYPE_UNCHOKE:
		state.conn.Choked = false
	case PEER_MSG_TYPE_HAVE:
		index, err := GetHaveIndex(msg)
		if err != nil {
			return err
		}

		state.conn.PiecesMap.Set(index)
	case PEER_MSG_TYPE_PIECE:
		n, err := CopyPieceData(state.index, state.data, msg)
		if err != nil {
			return err
		}

		state.downloaded += n
		state.backlog--
	}

	return nil
}
