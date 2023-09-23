package bencode

import (
	"bufio"
	"errors"
	"io"
)

var (
	ErrExpectedEndIdentifier    = errors.New("expected number identifier 'e'")
	ErrExpectedNumberIdentifier = errors.New("expected number identifier 'i'")
	ErrExpectedStringIdentifier = errors.New("expected string identifier ':'")
	ErrInvalidStringLength      = errors.New("invalid string length")
	ErrInvalidBencode           = errors.New("invalid bencode")
)

func readDecimal(br *bufio.Reader) (int64, int64) {
	var isNegative int64 = 1
	var val, count int64 = 0, 1

	b, _ := br.ReadByte()
	if b == '-' {
		isNegative = -1

		b, _ = br.ReadByte()
		count++
	}

	for {
		if b < '0' || b > '9' {
			br.UnreadByte()
			count--
			break
		}

		val = val*10 + int64(b-'0')
		b, _ = br.ReadByte()
		count++
	}

	return isNegative * val, count
}

// TODO func encodeInt(val int) (string,int){}

func decodeInt(br *bufio.Reader) (int64, error) {
	b, err := br.ReadByte()
	if err != nil {
		return 0, nil
	}

	if b != 'i' {
		return 0, ErrExpectedNumberIdentifier
	}

	res, _ := readDecimal(br)
	b, err = br.ReadByte()
	if err != nil {
		return 0, err
	}

	if b != 'e' {
		return 0, ErrExpectedEndIdentifier
	}

	return res, nil
}

// TODO func encodeStr(val string) (string, int) {}

func decodeStr(br *bufio.Reader) (string, error) {
	strLen, len := readDecimal(br)
	if len == 0 {
		return "", ErrInvalidStringLength
	}

	b, err := br.ReadByte()
	if err != nil {
		return "", err
	}

	if b != ':' {
		return "", ErrExpectedStringIdentifier
	}

	buf := make([]byte, strLen)
	_, err = io.ReadAtLeast(br, buf, int(strLen))
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func parse(br *bufio.Reader) (any, error) {
	var ret any
	var err error

	bs, err := br.Peek(1)
	if err != nil {
		return nil, err
	}
	b := bs[0]

	if b == 'i' {
		ret, err = decodeInt(br)
	} else if b >= '0' && b <= '9' {
		ret, err = decodeStr(br)
	} else if b == 'l' {
		br.ReadByte()
		var list []any

		for {
			bs, err = br.Peek(1)
			if err != nil {
				return nil, err
			}

			if bs[0] == 'e' {
				br.ReadByte()
				break
			}

			elem, err := parse(br)
			if err != nil {
				return nil, err
			}

			list = append(list, elem)
		}

		ret, err = list, nil
	} else if b == 'd' {
		br.ReadByte()
		dict := make(map[string]any)

		for {
			bs, err = br.Peek(1)
			if err != nil {
				return nil, err
			}

			if bs[0] == 'e' {
				br.ReadByte()
				break
			}

			key, err := decodeStr(br)
			if err != nil {
				return nil, err
			}

			val, err := parse(br)
			if err != nil {
				return nil, err
			}

			dict[key] = val
		}

		ret, err = dict, nil
	} else {
		ret = nil
		err = ErrInvalidBencode
	}

	return ret, err
}
