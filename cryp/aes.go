package cryp

import (
	"crypto/aes"
)

func AES(key []byte) Crypto {
	block, _ := aes.NewCipher(key[:])
	return &cfb{block}
}
