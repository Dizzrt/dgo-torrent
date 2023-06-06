package torrent

import (
	"errors"

	"github.com/Dizzrt/dgo-torrent/model"
)

var (
	ErrUnknown = errors.New("unknown error")
)

func NewTorrent(tfPath string) (model.Torrent, error) {
	return model.NewTorrent(tfPath)
}
