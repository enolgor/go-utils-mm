package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

type KeyDerivator interface {
	Derive(password string, salt []byte) []byte
}

type keyDerivator struct {
	iterations int
	size       KeySize
}

type KeySize int

const Key256 KeySize = 256

func (kd *keyDerivator) Derive(password string, salt []byte) []byte {
	return deriveKey(password, kd.size, salt, kd.iterations)
}

func deriveKey(password string, size KeySize, salt []byte, iterations int) []byte {
	hash := hashAlgorithmBySize(size)
	return pbkdf2.Key([]byte(password), salt, iterations, int(size)/8, hash)
}

func (size KeySize) NewDerivator(iterations int) *keyDerivator {
	return &keyDerivator{iterations: iterations, size: size}
}

func (size KeySize) OptimalIterations(target time.Duration) int {
	iter := minimumIterations(size)
	hash := hashAlgorithmBySize(size)
	salt := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic(err)
	}
	start := time.Now()
	pbkdf2.Key([]byte("microbenchmark"), salt, iter, int(size)/8, hash)
	end := time.Now()
	dur := end.Sub(start)
	for dur < target {
		iter = iter * 2
		dur = dur * 2
	}
	iter = iter / 2
	return iter
}

func minimumIterations(size KeySize) int {
	if size == 256 {
		return 4096
	}
	panic(fmt.Sprintf("unsupported size %d", int(size)))
}

func hashAlgorithmBySize(size KeySize) func() hash.Hash {
	if size == 256 {
		return sha256.New
	}
	panic(fmt.Sprintf("unsupported size %d", int(size)))
}
