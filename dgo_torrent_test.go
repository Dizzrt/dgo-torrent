package dgotorrent_test

import (
	"encoding/json"
	"os"
	"testing"

	dgotorrent "github.com/Dizzrt/dgo-torrent"
)

func unmarshaTorrentFile(filePath string, t string) error {
	file, _ := os.Open(filePath)
	defer file.Close()

	tf, err := dgotorrent.Unmarshal(file)
	if err != nil {
		return err
	}

	jbytes, err := json.Marshal(tf)
	if err != nil {
		return err
	}

	if _, err := os.Stat("./test/out"); os.IsNotExist(err) {
		os.Mkdir("./test/out", os.ModePerm)
	}

	out, _ := os.OpenFile("test/out/torrent_"+t+"_file_umarshal_test_resutl.json", os.O_WRONLY|os.O_CREATE, 0666)
	defer out.Close()
	out.Write(jbytes)

	return nil
}

func TestTorrentFile(t *testing.T) {
	if err := unmarshaTorrentFile("test/singleFile.torrent", "single"); err != nil {
		t.Error(err)
	}

	if err := unmarshaTorrentFile("test/mutiFile.torrent", "muti"); err != nil {
		t.Error(err)
	}
}
