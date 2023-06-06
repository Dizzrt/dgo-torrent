package torrent

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	bencode "github.com/Dizzrt/dgo-bencode"
	"github.com/Dizzrt/dgo-torrent/model"
)

// const _IDLEN int = 20

// TODO
// func GetTrackers(tt *model.Torrent) any {

// }

func buildQuery(tt *model.Torrent) (ret string, err error) {
	var peerId [20]byte
	_, err = rand.Read(peerId[:])
	if err != nil {
		return
	}

	base, err := url.Parse(tt.Announce)
	if err != nil {
		return
	}

	params := url.Values{
		"info_hash":  []string{string(tt.InfoSHA[:])},
		"peer_id":    []string{string(peerId[:])},
		"port":       []string{strconv.Itoa(6666)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(tt.Info.Length)},
	}

	base.RawQuery = params.Encode()
	return base.String(), nil
}

func getTrackerResp(tt *model.Torrent) (map[string]any, error) {
	url, err := buildQuery(tt)
	if err != nil {
		fmt.Println("Build Tracker Url Error: " + err.Error())
		return nil, ErrUnknown
	}

	cli := &http.Client{
		Timeout: 15 * time.Second,
	}

	trackerResp, err := cli.Get(url)
	if err != nil {
		fmt.Println("failed to get connect to tracker: " + err.Error())
		return nil, ErrUnknown
	}
	defer trackerResp.Body.Close()

	resp, err := bencode.Unmarshal(trackerResp.Body)
	if err != nil {
		fmt.Println("tracker response error: " + err.Error())
		return nil, ErrUnknown
	}

	if resMap, ok := resp.(map[string]any); ok {
		return resMap, nil
	}

	return nil, ErrUnknown
}
