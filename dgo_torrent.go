package torrent

import "os"

type TorrentTask struct {
	PeerID   PeerID_t
	PeerList []Peer
	SHA      [SHALEN]byte
	FileName string
	FileLen  int
	PieceLen int
	PicesSHA [][SHALEN]byte
}

type pieceTask struct {
	index  int
	sha    [SHALEN]byte
	length int
}

type pieceResult struct {
	index int
	data  []byte
}

func (task *TorrentTask) getPieceBounds(index int) (begin, end int) {
	begin = index * task.PieceLen
	end = begin + task.PieceLen
	if end > task.FileLen {
		end = task.FileLen
	}
	return
}

func (task *TorrentTask) peerRoutine(peer Peer, taskQueue chan *pieceTask, resultQueue chan *pieceResult) error {
	conn, err := NewPeerConn(peer, task.SHA, task.PeerID)
	if err != nil {
		return err
	}
	defer conn.Close()

	// TODO
	return nil
}

func Download(task *TorrentTask) error {
	taskQueue := make(chan *pieceTask, len(task.PicesSHA))
	resultQueue := make(chan *pieceResult)

	for index, sha := range task.PicesSHA {
		begin, end := task.getPieceBounds(index)
		taskQueue <- &pieceTask{
			index:  index,
			sha:    sha,
			length: end - begin,
		}
	}

	for _, peer := range task.PeerList {
		go task.peerRoutine(peer, taskQueue, resultQueue)
	}

	count := 0
	buf := make([]byte, task.FileLen)
	for count < len(task.PicesSHA) {
		res := <-resultQueue
		begin, end := task.getPieceBounds(res.index)
		copy(buf[begin:end], res.data)
		count++
	}

	close(taskQueue)
	close(resultQueue)

	file, err := os.Create(task.FileName)
	if err != nil {
		return err
	}

	_, err = file.Write(buf)
	if err != nil {
		return err
	}

	return nil
}
