package random

import (
	"crypto/rand"
	"io"
)

func MustReadBytes(v []byte) {
	if _, err := io.ReadFull(rand.Reader, v); err != nil {
		panic(err)
	}
}

func Bytes(size int) []byte {
	b := make([]byte, size)
	MustReadBytes(b)
	return b
}
