package dgotorrent

import (
	"errors"
	"fmt"
	"io"

	"github.com/Dizzrt/dgo-torrent/bencode"
	"github.com/google/uuid"
)

const PIECE_LEN = 20

var (
	ErrInvalidTorrentFile = errors.New("invalid torrent file")
)

type TorrentMutiFile struct {
	Path   []string
	Length int64
}

type TorrentInfo struct {
	Name        string
	IsMutiFile  bool
	Length      int64
	MutiFiles   []TorrentMutiFile
	PieceLength int64
	Pieces      [][PIECE_LEN]byte
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
		value, ok := v.(int)
		if ok {
			tf.CreatedAt = int64(value)
		} else {
			tf.CreatedAt = 0
		}
	}

	return nil
}

func parseMutiFile(info *TorrentInfo, fileList []any) error {
	filesCount := len(fileList)
	list := make([]TorrentMutiFile, 0, filesCount)

	for _, _file := range fileList {
		file, ok := _file.(map[string]any)
		if !ok {
			continue
		}

		tmf := TorrentMutiFile{}
		if v, ok := file["length"]; ok {
			value, ok := v.(int64)
			if ok {
				tmf.Length = value
			} else {
				return ErrInvalidTorrentFile
			}
		} else {
			return ErrInvalidTorrentFile
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

				tmf.Path = path
			} else {
				return ErrInvalidTorrentFile
			}
		} else {
			return ErrInvalidTorrentFile
		}

		list = append(list, tmf)
	}

	info.MutiFiles = list
	info.IsMutiFile = true

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
			info.Pieces = pieces
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

func Unmarshal(r io.Reader) (*TorrentFile, error) {
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
		}
	} else {
		return nil, ErrInvalidTorrentFile
	}

	return tf, nil
}
