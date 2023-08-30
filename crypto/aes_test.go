package crypto

import (
	"testing"

	"github.com/enolgor/go-utils-mm/parse"
)

var plain string = "this is some text"

func TestAES(t *testing.T) {
	aes := AES(parse.Must(parse.HexBytes)("bc27bec0c4291b4e43a2ec657d8afc9b668e158c6acd4004ffb1faa16c5b88bf"))
	enc, err := aes.Encrypt([]byte(plain))
	if err != nil {
		t.Errorf("error: %s", err)
	}
	dec, err := aes.Decrypt(enc)
	if err != nil {
		t.Errorf("error: %s", err)
	}
	if string(dec) != plain {
		t.Errorf("got %s, want %s", string(dec), plain)
	}
}
