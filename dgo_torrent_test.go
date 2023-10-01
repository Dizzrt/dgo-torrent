package dgotorrent_test

import (
	"encoding/json"
	"os"
	"testing"

	dgotorrent "github.com/Dizzrt/dgo-torrent"
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
