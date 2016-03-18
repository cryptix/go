package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"io"

	"gopkg.in/errgo.v1"
)

// Crypter can only be used once
type Crypter struct {
	key   []byte
	block cipher.Block
	used  bool
}

// NewCrypter creates a new Crypter, or nil if there is an error
func NewCrypter(key []byte) (e *Crypter, err error) {
	e = new(Crypter)

	if len(key) != sha512.Size256 {
		return nil, errgo.Newf("whatwhat: wrong key length. Got:%d", len(key))
	}

	e.key = key

	e.block, err = aes.NewCipher(e.key)
	if err != nil {
		return nil, errgo.Notef(err, "whatwhat: couldn't create AES cipher")
	}

	return e, nil
}

// MakePipe takes an output (for the cipher text) writer and returns a writer to which you writer your cleartext
func (e *Crypter) MakePipe(out io.Writer) (io.Writer, error) {
	if e.used == true {
		return nil, errgo.Newf("whatwhat: crypter was used twice")
	}

	// we only use Crypter once, its ok to have a zero IV
	var iv [aes.BlockSize]byte
	stream := cipher.NewCTR(e.block, iv[:])

	e.used = true

	return &cipher.StreamWriter{S: stream, W: out}, nil
}
