package dgotorrent

type Bitfield []byte

func (field Bitfield) Test(index int) bool {
	offset := index % 8
	byteOffset := index / 8
	if byteOffset < 0 || byteOffset > len(field) {
		return false
	}

	return field[byteOffset]>>uint(7-offset)&1 != 0
}
