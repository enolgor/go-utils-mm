package cryp

import "encoding/base64"

type Crypto interface {
	Encrypt(plain []byte) (string, error)
	Decrypt(encrypted string) ([]byte, error)
}

func encodeString(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

func decodeString(data string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(data)
}
