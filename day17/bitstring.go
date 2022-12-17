package main

type Bitstring struct {
	size  int
	bytes string
}

func MakeBitstring(size int) (result Bitstring) {
	bytelen := size / 8
	if size%8 != 0 {
		bytelen++
	}

	result.size = size
	for i := 0; i < bytelen; i++ {
		result.bytes += string([]byte{0})
	}
	return
}

func location(pos int) (index int, flag byte) {
	index = pos / 8
	flag = 1
	flag <<= (pos % 8)
	return
}

func (bs *Bitstring) Size() int {
	return bs.size
}

func (bs *Bitstring) At(pos int) bool {
	if pos < 0 || pos >= bs.size {
		panic("Out of bounds")
	}
	index, flag := location(pos)
	return (bs.bytes[index] & flag) != 0
}

func (bs *Bitstring) Set(pos int, value bool) {
	if pos < 0 || pos >= bs.size {
		panic("Out of bounds")
	}
	index, flag := location(pos)

	b := bs.bytes[index]
	if value {
		b |= flag
	} else {
		b &= ^flag
	}
	newbytes := bs.bytes[:index]
	newbytes += string([]byte{b})
	newbytes += bs.bytes[index+1:]
	bs.bytes = newbytes
}
