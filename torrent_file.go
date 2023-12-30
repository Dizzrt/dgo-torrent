package dgotorrent

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"

	"github.com/Dizzrt/dgo-torrent/bencode"
	"github.com/google/uuid"
)

var (
	ErrInvalidTorrentFile = errors.New("invalid torrent file")
)

const PIECE_LEN = 20
const INFO_HASH_LEN = 20

type TMF_TYPE uint8

const (
	TMF_TYPE_DIRECTORY = iota
	TMF_TYPE_FILE
)

// type TorrentMutiFIleType uint8

// const (
// 	TMFT_DIRECTORY = iota
// 	TMFT_FILE
// )

// type TorrentMutiFile struct {
// 	Type      TorrentMutiFIleType
// 	Name      string
// 	Size      int64
// 	Subs      map[string]TorrentMutiFile
// 	SubsOrder []string
// }

type TorrentMutiFile struct {
	Type      TMF_TYPE
	Name      string
	Length    int64
	Subs      map[string]TorrentMutiFile
	SubsOrder []string
}

type iMutiFile struct {
	Path   []string
	Length int64
}

type TorrentInfo struct {
	Name        string
	IsMutiFile  bool
	Length      int64
	MutiFiles   TorrentMutiFile
	PieceLength int64
	PieceHashes [][PIECE_LEN]byte
	Hash        [INFO_HASH_LEN]byte
}

type TorrentFile struct {
	Announce     string
	AnnounceList []string
	Comment      string
	CreatedBy    string
	CreatedAt    int64
	Info         TorrentInfo
}

func parseAnnounceList(announceList []any) []string {
	list := make([]string, 0, len(announceList))

	for _, value := range announceList {
		if v, ok := value.(string); ok {
			list = append(list, v)
			continue
		}

		if v, ok := value.([]any); ok {
			for _, vv := range v {
				if vvv, ok := vv.(string); ok {
					list = append(list, vvv)
				}
			}
		}
	}

	return list
}

func parseBase(tf *TorrentFile, tfMap map[string]any) error {
	// announce
	if v, ok := tfMap["announce"]; ok {
		value, ok := v.(string)
		if !ok {
			return ErrInvalidTorrentFile
		}

		tf.Announce = value
	} else {
		return ErrInvalidTorrentFile
	}

	// announce-list
	if v, ok := tfMap["announce-list"]; ok {
		value, ok := v.([]any)
		if ok {
			tf.AnnounceList = parseAnnounceList(value)
		} else {
			tf.AnnounceList = make([]string, 0)
		}
	}

	// comment
	if v, ok := tfMap["comment"]; ok {
		value, ok := v.(string)
		if ok {
			tf.Comment = value
		} else {
			tf.Comment = ""
		}
	}

	// created_by
	if v, ok := tfMap["created by"]; ok {
		value, ok := v.(string)
		if ok {
			tf.CreatedBy = value
		} else {
			tf.CreatedBy = ""
		}
	}

	// get created_at
	if v, ok := tfMap["creation date"]; ok {
		value, ok := v.(int64)
		if ok {
			tf.CreatedAt = value
		} else {
			tf.CreatedAt = 0
		}
	}

	return nil
}

func buildIMutiFile(info *TorrentInfo, fileList []any) ([]iMutiFile, error) {
	filesCount := len(fileList)
	ret := make([]iMutiFile, 0, filesCount)

	for _, _file := range fileList {
		file, ok := _file.(map[string]any)
		if !ok {
			continue
		}

		imf := iMutiFile{}
		if v, ok := file["length"]; ok {
			value, ok := v.(int64)
			if ok {
				imf.Length = value
			} else {
				return nil, ErrInvalidTorrentFile
			}
		} else {
			return nil, ErrInvalidTorrentFile
		}

		if v, ok := file["path"]; ok {
			value, ok := v.([]any)
			if ok {
				path := make([]string, 0, len(value))
				for _, _p := range value {
					if p, ok := _p.(string); ok {
						path = append(path, p)
					}
				}
				imf.Path = path
			} else {
				return nil, ErrInvalidTorrentFile
			}
		} else {
			return nil, ErrInvalidTorrentFile
		}

		ret = append(ret, imf)
	}

	return ret, nil
}

func buildTorrentMutiFile(upper TorrentMutiFile, path []string, length int64) TorrentMutiFile {
	if len(path) == 1 {
		upper.Subs[path[0]] = TorrentMutiFile{
			Type:      TMF_TYPE_FILE,
			Name:      path[0],
			Length:    length,
			Subs:      nil,
			SubsOrder: nil,
		}

		upper.SubsOrder = append(upper.SubsOrder, path[0])
		return upper
	}

	var tmf TorrentMutiFile
	if temp, ok := upper.Subs[path[0]]; ok {
		tmf = temp
	} else {
		tmf = TorrentMutiFile{
			Type:      TMF_TYPE_DIRECTORY,
			Name:      path[0],
			Length:    0,
			Subs:      make(map[string]TorrentMutiFile),
			SubsOrder: make([]string, 0),
		}

		upper.SubsOrder = append(upper.SubsOrder, path[0])
	}

	tmf = buildTorrentMutiFile(tmf, path[1:], length)
	upper.Subs[path[0]] = tmf

	return upper
}

func parseMutiFile(info *TorrentInfo, fileList []any) error {
	imfs, err := buildIMutiFile(info, fileList)
	if err != nil {
		return err
	}

	tmf := TorrentMutiFile{
		Type:      TMF_TYPE_DIRECTORY,
		Name:      info.Name,
		Length:    info.Length,
		Subs:      make(map[string]TorrentMutiFile),
		SubsOrder: make([]string, 0),
	}

	for _, v := range imfs {
		tmf = buildTorrentMutiFile(tmf, v.Path, v.Length)
	}

	info.MutiFiles = tmf
	return nil
}

func parseInfo(tf *TorrentFile, infoMap map[string]any) error {
	info := &tf.Info

	// name
	if v, ok := infoMap["name"]; ok {
		if value, ok := v.(string); ok {
			info.Name = value
		}
	}

	if len(info.Name) == 0 {
		info.Name = uuid.New().String()
	}

	// piece_length
	if v, ok := infoMap["piece length"]; ok {
		value, ok := v.(int64)
		if ok {
			info.PieceLength = value
		} else {
			fmt.Println("piece length is required")
			return ErrInvalidTorrentFile
		}
	}

	// pieces
	if v, ok := infoMap["pieces"]; ok {
		value, ok := v.(string)
		if ok {
			raw := []byte(value)
			count := len(raw) / PIECE_LEN
			pieces := make([][PIECE_LEN]byte, count)

			for i := 0; i < count; i++ {
				copy(pieces[i][:], raw[i*PIECE_LEN:(i+1)*PIECE_LEN])
			}
			info.PieceHashes = pieces
		} else {
			return ErrInvalidTorrentFile
		}
	}

	// files
	if v, ok := infoMap["files"]; ok {
		if value, ok := v.([]any); ok {
			if err := parseMutiFile(info, value); err != nil {
				return err
			}
		} else {
			return ErrInvalidTorrentFile
		}

		info.IsMutiFile = true
	} else {
		// single file
		if v, ok = infoMap["length"]; ok {
			if value, ok := v.(int64); ok {
				info.Length = value
			} else {
				return ErrInvalidTorrentFile
			}
		}

		info.IsMutiFile = false
	}

	return nil
}

func NewTorrentFile(r io.Reader) (*TorrentFile, error) {
	res, err := bencode.Unmarshal(r)
	if err != nil {
		return nil, err
	}

	tf := &TorrentFile{}
	tfMap, ok := res.(map[string]any)
	if !ok {
		return nil, ErrInvalidTorrentFile
	}

	if err = parseBase(tf, tfMap); err != nil {
		return nil, err
	}

	if v, ok := tfMap["info"]; ok {
		if value, ok := v.(map[string]any); ok {
			if err = parseInfo(tf, value); err != nil {
				return nil, err
			}

			infoBencode, err := bencode.Marshal(value)
			if err != nil {
				return nil, err
			}
			tf.Info.Hash = sha1.Sum([]byte(infoBencode))
		}
	} else {
		return nil, ErrInvalidTorrentFile
	}

	return tf, nil
}
