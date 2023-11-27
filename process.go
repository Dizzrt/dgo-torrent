package dgotorrent

import (
	"bytes"
	"crypto/sha1"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/Dizzrt/dgo-torrent/dlog"
	"github.com/schollz/progressbar/v3"
)

const (
	BLOCKSIZE  = 16384
	MAXBACKLOG = 5
)

type pJob struct {
	index  int
	hash   [20]byte
	length int
}

type pJobState struct {
	index      int
	conn       *PeerConn
	requested  int
	downloaded int
	backlog    int
	data       []byte
}

type pJobResult struct {
	index int
	data  []byte
}

type Process struct {
	Task  *Task
	Peers []Peer
}

func NewProcess(task *Task) *Process {
	p := &Process{
		Task:  task,
		Peers: make([]Peer, 0),
	}

	return p
}

func (p *Process) Start() error {
	jobs := make(chan *pJob, len(p.Task.Torrent.Info.PieceHashes))
	results := make(chan *pJobResult)

	var wg sync.WaitGroup
	wg.Add(2)

	// search peers
	go func() {
		defer wg.Done()

		peers, err := p.Task.Torrent.FindPeers()
		if err != nil {
			dlog.Errorf("search peer failed with error: %v", err)
		}

		p.Peers = peers
	}()

	// init jobs
	go func() {
		wg.Done()

		for index, hash := range p.Task.Torrent.Info.PieceHashes {
			begin, end := p.Task.GetPieceBounds(index)
			jobs <- &pJob{
				index:  index,
				hash:   hash,
				length: end - begin,
			}
		}
	}()

	wg.Wait()

	for _, peer := range p.Peers {
		go p.peerRoutine(peer, jobs, results)
	}

	bar := progressbar.DefaultBytes(p.Task.Torrent.Info.Length, "downloading")

	// copy data
	count := 0
	buf := make([]byte, p.Task.Torrent.Info.Length)
	for count < len(p.Task.Torrent.Info.PieceHashes) {
		res := <-results
		begin, end := p.Task.GetPieceBounds(res.index)
		copy(buf[begin:end], res.data)
		count++

		io.Copy(bar, bytes.NewReader(res.data))
		// io.Copy(io.MultiWriter(bar), bytes.NewReader(res.data))
		// percent := float64(count) / float64(len(p.Task.Torrent.Info.PieceHashes)) * 100
		// fmt.Printf("downloading progress: %0.2f%%\n", percent)
	}

	close(jobs)
	close(results)

	file, err := os.Create(path.Join(p.Task.Path, p.Task.Name))
	if err != nil {
		dlog.Errorf("failed to create file: %v", p.Task.Name)
		return err
	}

	_, err = file.Write(buf)
	if err != nil {
		dlog.Error("failed to write data")
		return err
	}

	return nil
}

func (p *Process) peerRoutine(peer Peer, jobs chan *pJob, results chan *pJobResult) {
	conn, err := NewConn(peer, p.Task.Torrent.Info.Hash, p.Task.PeerID)
	if err != nil {
		dlog.Infof("fail to connect peer: %s:%d", peer.IP.String(), peer.Port)
		return
	}
	defer conn.Close()

	conn.WriteMsg(&PeerMsg{
		Type:    PEER_MSG_TYPE_INTERESTED,
		Payload: nil,
	})

	for job := range jobs {
		if !conn.PiecesMap.Test(job.index) {
			jobs <- job
			continue
		}

		res, err := downloadPiece(conn, job)
		if err != nil {
			jobs <- job
			dlog.Error("failed to download piece with error: %v", err)
			return
		}

		if !checkPiece(job, res) {
			jobs <- job
			continue
		}

		results <- res
	}
}

func downloadPiece(conn *PeerConn, job *pJob) (*pJobResult, error) {
	state := pJobState{
		index:   job.index,
		conn:    conn,
		data:    make([]byte, job.length),
		backlog: 0,
	}

	conn.SetDeadline(time.Now().Add(15 * time.Second))
	defer conn.SetDeadline(time.Time{})

	for state.downloaded < job.length {
		if !conn.Choked {
			for state.backlog < MAXBACKLOG && state.requested < job.length {
				length := BLOCKSIZE
				if job.length-state.requested < length {
					length = job.length - state.requested
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

	return &pJobResult{
		index: state.index,
		data:  state.data,
	}, nil
}

func (s *pJobState) handleMsg() error {
	msg, err := s.conn.ReadMsg()
	if err != nil {
		return nil
	}

	if msg == nil {
		return nil
	}

	switch msg.Type {
	case PEER_MSG_TYPE_CHOKE:
		s.conn.Choked = true
	case PEER_MSG_TYPE_UNCHOKE:
		s.conn.Choked = false
	case PEER_MSG_TYPE_HAVE:
		index, err := GetHaveIndex(msg)
		if err != nil {
			return err
		}

		s.conn.PiecesMap.Set(index)
	case PEER_MSG_TYPE_PIECE:
		n, err := CopyPieceData(s.index, s.data, msg)
		if err != nil {
			return err
		}

		s.downloaded += n
		s.backlog--
	}

	return nil
}

func checkPiece(job *pJob, res *pJobResult) bool {
	hash := sha1.Sum(res.data)
	if !bytes.Equal(job.hash[:], hash[:]) {
		dlog.Infof("check integrity failed, index: %v", res.index)
		return false
	}

	return true
}
