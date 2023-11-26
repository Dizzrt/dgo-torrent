package dgotorrent_test

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	dgotorrent "github.com/Dizzrt/dgo-torrent"
	"github.com/Dizzrt/dgo-torrent/config"
	"github.com/Dizzrt/dgo-torrent/dlog"
)

func TestDownload(t *testing.T) {
	file, _ := os.Open("test/debian.torrent")
	defer file.Close()

	tf, err := dgotorrent.NewTorrentFile(file)
	if err != nil {
		t.Error(err)
	}

	task := &dgotorrent.Task{
		ID:        1,
		Name:      tf.Info.Name,
		Path:      config.Instance().GetDefaultDonwloadPath(),
		Status:    make(map[string]any),
		State:     dgotorrent.TASK_STATE_PAUSED,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PeerID:    config.Instance().GetPeerID(),
		Torrent:   *tf,
	}

	process := dgotorrent.NewProcess(task)
	process.Start()

	// peers, err := tf.FindPeers()
	// if err != nil {
	// 	t.Error(err)
	// }

	// task := dgotorrent.TorrentTask{
	// 	PeerID:      config.Instance().GetPeerID(),
	// 	Peers:       peers,
	// 	InfoHash:    tf.Info.Hash,
	// 	FileName:    "/Users/dizzrt/Downloads/debian.iso",
	// 	FileLength:  int(tf.Info.Length),
	// 	PieceLength: int(tf.Info.PieceLength),
	// 	PiecesHash:  tf.Info.PieceHashes,
	// }

	// dgotorrent.Download(&task)
	dlog.L().Sync()
}

func checkTestDir() {
	if _, err := os.Stat("./test/out"); os.IsNotExist(err) {
		os.Mkdir("./test/out", os.ModePerm)
	}
}

func unmarshaTorrentFile(filePath string, t string) error {
	file, _ := os.Open(filePath)
	defer file.Close()

	tf, err := dgotorrent.NewTorrentFile(file)
	if err != nil {
		return err
	}

	jbytes, err := json.Marshal(tf)
	if err != nil {
		return err
	}

	checkTestDir()

	out, _ := os.OpenFile("test/out/torrent_"+t+"_file_umarshal_test_resutl.json", os.O_WRONLY|os.O_CREATE, 0666)
	defer out.Close()
	out.Write(jbytes)

	return nil
}

func TestTest(t *testing.T) {
	fmt.Println(config.Instance().GetDefaultDonwloadPath())
}

func TestTorrentFile(t *testing.T) {
	if err := unmarshaTorrentFile("test/debian.torrent", "debian"); err != nil {
		t.Error(err)
	}

	dlog.L().Sync()
	// if err := unmarshaTorrentFile("test/mutiFile.torrent", "muti"); err != nil {
	// 	t.Error(err)
	// }
}

func TestTrackerRequest(t *testing.T) {
	// announceList := []string{}
	// tf := &dgotorrent.TorrentFile{AnnounceList: announceList}

	file, _ := os.Open("./test/debian.torrent")
	defer file.Close()

	tf, err := dgotorrent.NewTorrentFile(file)
	if err != nil {
		t.Error(err)
	}

	_, err = tf.RequestTrackers()
	if err != nil {
		t.Error(err)
	}

	// jbyte, err := json.Marshal(res)
	// if err != nil {
	// 	fmt.Printf("%v\n", res)
	// 	t.Error(err)
	// }

	// checkTestDir()
	// out, _ := os.OpenFile("test/out/torrent_tracker_request_test_result.json", os.O_WRONLY|os.O_CREATE, 0666)
	// defer out.Close()
	// out.Write(jbyte)
}

func TestFindPeer(t *testing.T) {
	file, _ := os.Open("./test/fs.torrent")
	defer file.Close()

	tf, err := dgotorrent.NewTorrentFile(file)
	if err != nil {
		t.Error(err)
	}

	peers, err := tf.FindPeers()
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%+v\n", peers)
	fmt.Printf("peer count: %d\n", len(peers))
}

func TestPeerConn(t *testing.T) {
	file, _ := os.Open("./test/fs.torrent")
	defer file.Close()

	tf, err := dgotorrent.NewTorrentFile(file)
	if err != nil {
		t.Error(err)
	}

	peers, err := tf.FindPeers()
	if err != nil {
		t.Error(err)
	}

	var wg sync.WaitGroup
	broadcast := net.ParseIP("255.255.255.255")
	self := net.ParseIP("0.0.0.0")

	dlog.Infof("peers: %+v\npeer count:%d", peers, len(peers))
	for _, pp := range peers {
		if pp.IP.Equal(broadcast) || pp.IP.Equal(self) {
			continue
		}

		wg.Add(1)
		go func(p dgotorrent.Peer) {
			defer wg.Done()

			pc, err := dgotorrent.NewConn(p, tf.Info.Hash, config.Instance().GetPeerID())
			if err != nil {
				// t.Error(err)
				dlog.Errorf("error: %v", err)
			} else {
				defer pc.Close()
				dlog.Infof("bitfield: %+v", pc.PiecesMap)
			}
		}(pp)
	}

	wg.Wait()
	dlog.L().Sync()
}
