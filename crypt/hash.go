package crypt

import (
	"crypto/sha512"
	"io"

	"gopkg.in/errgo.v1"
)

func GetKey(r io.Reader) ([]byte, error) {
	h := sha512.New512_256()

	_, err := io.Copy(h, r)
	if err != nil {
		return nil, errgo.Notef(err, "")
	}

	return h.Sum(nil), nil
}
