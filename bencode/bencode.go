package bencode

import (
	"bufio"
	"io"
)

func Unmarshal(r io.Reader) (any, error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}

	res, err := parse(br)
	return res, err
}
