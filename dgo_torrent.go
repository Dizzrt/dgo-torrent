package dgotorrent

// func Download(task *TorrentTask) error {
// 	works := make(chan *TorrentWork, len(task.PiecesHash))
// 	results := make(chan *TorrentWorkResult)

// 	for index, hash := range task.PiecesHash {
// 		begin, end := task.getPieceBounds(index)
// 		works <- &TorrentWork{
// 			index:  index,
// 			hash:   hash,
// 			length: end - begin,
// 		}
// 	}

// 	for _, peer := range task.Peers {
// 		go task.peerRoutine(peer, works, results)
// 	}

// 	count := 0
// 	buf := make([]byte, task.FileLength)
// 	for count < len(task.PiecesHash) {
// 		res := <-results
// 		begin, end := task.getPieceBounds(res.index)
// 		copy(buf[begin:end], res.data)
// 		count++

// 		percent := float64(count) / float64(len(task.PiecesHash)) * 100
// 		fmt.Printf("downloading, progress: %0.2f%%\n", percent)
// 	}

// 	close(works)
// 	close(results)

// 	file, err := os.Create(task.FileName)
// 	if err != nil {
// 		dlog.Errorf("fail to create file: %v", task.FileName)
// 		return err
// 	}

// 	_, err = file.Write(buf)
// 	if err != nil {
// 		dlog.Error("fail to write data")
// 		return err
// 	}

// 	return nil
// }
