package dgotorrent_test

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"

	dgotorrent "github.com/Dizzrt/dgo-torrent"
	"github.com/Dizzrt/dgo-torrent/dlog"
)

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

func TestTorrentFile(t *testing.T) {
	if err := unmarshaTorrentFile("test/fs.torrent", "fs"); err != nil {
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

	file, _ := os.Open("./test/fs.torrent")
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
	file, _ := os.Open("./test/debian.torrent")
	defer file.Close()

	tf, err := dgotorrent.NewTorrentFile(file)
	if err != nil {
		t.Error(err)
	}

	peers, err := tf.FindPeers()
	if err != nil {
		t.Error(err)
	}

	var peerId [20]byte
	_, _ = rand.Read(peerId[:])

	var wg sync.WaitGroup
	broadcast := net.ParseIP("255.255.255.255")

	dlog.Infof("peers: %+v\npeer count:%d", peers, len(peers))
	for _, pp := range peers {
		if pp.IP.Equal(broadcast) {
			continue
		}

		wg.Add(1)
		go func(p dgotorrent.Peer) {
			defer wg.Done()

			pc, err := dgotorrent.NewConn(p, tf.Info.Hash, dgotorrent.Config().GetPeerID())
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
