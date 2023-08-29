package crypto

import (
	"crypto/cipher"
	"crypto/rand"
	"io"
)

type cfb struct {
	block cipher.Block
}

func (c *cfb) Encrypt(plain []byte) (string, error) {
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	blocksize := c.block.BlockSize()
	ciphertext := make([]byte, blocksize+len(plain))
	iv := ciphertext[:blocksize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	stream := cipher.NewCFBEncrypter(c.block, iv)
	stream.XORKeyStream(ciphertext[blocksize:], plain)
	return encodeString(ciphertext), nil
}

func (c *cfb) Decrypt(encrypted string) ([]byte, error) {
	data, err := decodeString(encrypted)
	if err != nil {
		return nil, err
	}
	blocksize := c.block.BlockSize()
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(data) < blocksize {
		return nil, err
	}
	iv := data[:blocksize]
	data = data[blocksize:]
	stream := cipher.NewCFBDecrypter(c.block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(data, data)
	return data, nil
}
