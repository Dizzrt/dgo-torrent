package model

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"reflect"

	bencode "github.com/Dizzrt/dgo-bencode"
	"github.com/google/uuid"
)

const SHALEN int = 20

var (
	ErrTorrentType          = errors.New("invalid type, taget file may not be a valid torrent file")
	ErrTorrentNoAnnounce    = errors.New("expected field \"announce\"")
	ErrTorrentNoInfo        = errors.New("expected field \"info\"")
	ErrTorrentNoPieceLength = errors.New("expected field \"info/piece length\"")
	ErrTorrentNoPiecesSHA   = errors.New("expected field \"info/pieces\"")
	ErrTorrentNoLength      = errors.New("expected file length")
	ErrTorrentNoPath        = errors.New("expected file path")
)

type TorrentMutiFileDetail struct {
	Length int
	Path   []string
}

type TorrentInfo struct {
	// common fields

	// root-directory name if mutifiles
	Name        string
	PieceLength int
	PiecesSHA   [][SHALEN]byte

	// single file fields

	Length int

	// muti-files fields

	IsMutiFiles bool
	Files       []TorrentMutiFileDetail
}

type Torrent struct {
	Announce     string
	AnnounceList []string
	CreatedBy    string
	CreationDate int64
	Comment      string
	Info         TorrentInfo
	InfoSHA      [SHALEN]byte
}

func NewTorrent(tfPath string) (ret Torrent, err error) {
	file, err := os.Open(tfPath)
	if err != nil {
		return
	}

	_tf, err := bencode.Unmarshal(file)
	if err != nil {
		return
	}

	tf, ok := _tf.(map[string]any)
	if !ok {
		return ret, ErrTorrentType
	}

	if v, ok := tf["announce"]; ok {
		ret.Announce = v.(string)
	} else {
		return ret, ErrTorrentNoAnnounce
	}

	if v, ok := tf["announce-list"]; ok {
		list := v.([]any)
		if _err := parseAnnounceList(&ret, &list); _err != nil {
			ret.AnnounceList = make([]string, 0)
			fmt.Println("parse announce list failed: " + _err.Error())
		}
	} else {
		ret.AnnounceList = make([]string, 0)
	}

	if v, ok := tf["created by"]; ok {
		ret.CreatedBy = v.(string)
	} else {
		ret.CreatedBy = ""
	}

	if v, ok := tf["creation date"]; ok {
		ret.CreationDate = int64(v.(int))
	} else {
		ret.CreationDate = 0
	}

	if v, ok := tf["comment"]; ok {
		ret.Comment = v.(string)
	} else {
		ret.Comment = ""
	}

	if v, ok := tf["info"]; ok {
		info := v.(map[string]any)
		if _err := parseInfo(&ret, &info); _err != nil {
			return ret, _err
		}
	} else {
		return ret, ErrTorrentNoInfo
	}

	return
}

func parseAnnounceList(tt *Torrent, list *[]any) error {
	for _, val := range *list {
		_type := reflect.TypeOf(val).Kind()
		switch _type {
		case reflect.String:
			tt.AnnounceList = append(tt.AnnounceList, val.(string))
		case reflect.Slice:
			tt.AnnounceList = append(tt.AnnounceList, val.([]any)[0].(string))
		default:
			fmt.Printf("can't parse announce: %#v", val)
		}
	}
	return nil
}

func parseInfo(tt *Torrent, info *map[string]any) error {
	infoBencode, err := bencode.Marshal(*info)
	if err != nil {
		return err
	}
	tt.InfoSHA = sha1.Sum([]byte(infoBencode))

	if v, ok := (*info)["name"]; ok {
		tt.Info.Name = v.(string)
	} else {
		tt.Info.Name = uuid.New().String()
	}

	if v, ok := (*info)["piece length"]; ok {
		tt.Info.PieceLength = v.(int)
	} else {
		return ErrTorrentNoPieceLength
	}

	if v, ok := (*info)["pieces"]; ok {
		bytes := []byte(v.(string))
		if _err := parsePieces(tt, &bytes); _err != nil {
			fmt.Println("parse pieces failed: " + _err.Error())
			return _err
		}
	} else {
		return ErrTorrentNoPiecesSHA
	}

	_files, ok := (*info)["files"]
	if ok {
		tt.Info.IsMutiFiles = true
		files := _files.([]any)
		if _err := parseMutiFiles(tt, &files); _err != nil {
			fmt.Println("parse files failed: " + _err.Error())
			return _err
		}
	} else {
		// single file
		tt.Info.IsMutiFiles = false

		if v, ok := (*info)["length"]; ok {
			tt.Info.Length = v.(int)
		} else {
			return ErrTorrentNoLength
		}
	}

	return nil
}

func parseMutiFiles(tt *Torrent, files *[]any) error {
	for _, _file := range *files {
		file := _file.(map[string]any)

		path := make([]string, 0)
		for _, p := range file["path"].([]any) {
			path = append(path, p.(string))
		}

		tt.Info.Files = append(tt.Info.Files, TorrentMutiFileDetail{
			Length: file["length"].(int),
			Path:   path,
		})
	}

	return nil
}

func parsePieces(tt *Torrent, bytes *[]byte) error {
	cnt := len(*bytes) / SHALEN

	hashes := make([][SHALEN]byte, cnt)
	for i := 0; i < cnt; i++ {
		copy(hashes[i][:], (*bytes)[i*SHALEN:(i+1)*SHALEN])
	}

	tt.Info.PiecesSHA = hashes
	return nil
}
