package bencode_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/Dizzrt/dgo-torrent/bencode"
)

func checkOutPath() {
	if _, err := os.Stat("../test/out"); os.IsNotExist(err) {
		os.Mkdir("../test/out", os.ModePerm)
	}
}

func TestUnmarshal(t *testing.T) {
	file, _ := os.Open("../test/mutiFile.torrent")
	defer file.Close()

	res, err := bencode.Unmarshal(file)
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}

	jbytes, err := json.Marshal(res)
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}

	checkOutPath()

	out, _ := os.OpenFile("../test/out/bencode_unmarshal_test_result.json", os.O_WRONLY|os.O_CREATE, 0666)
	defer out.Close()
	out.Write(jbytes)
}

func TestMarshal(t *testing.T) {
	val := struct {
		Length      int64  `bencode:"length"`
		Name        string `bencode:"name"`
		PieceLength int64  `bencode:"piece length"`
		Pieces      string `bencode:"pieces"`
	}{
		Length:      123,
		Name:        "xxx",
		PieceLength: 456,
		Pieces:      "vvv",
	}

	res, err := bencode.Marshal(val)
	if err != nil {
		t.Error(err)
	}

	checkOutPath()

	out, _ := os.OpenFile("../test/out/bencode_marshal_test_result.bencode", os.O_WRONLY|os.O_CREATE, 0666)
	defer out.Close()
	out.Write([]byte(res))
}
