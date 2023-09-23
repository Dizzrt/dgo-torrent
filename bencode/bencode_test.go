package bencode_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/Dizzrt/dgo-torrent/bencode"
)

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

	if _, err := os.Stat("../test/out"); os.IsNotExist(err) {
		os.Mkdir("../test/out", os.ModePerm)
	}

	out, _ := os.OpenFile("../test/out/bencode_unmarshal_test_result.json", os.O_WRONLY|os.O_CREATE, 0666)
	defer out.Close()
	out.Write(jbytes)
}
