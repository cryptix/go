package crypt

import (
	"crypto/sha512"
	"io"

	"github.com/pkg/errors"
)

func GetKey(r io.Reader) ([]byte, error) {
	h := sha512.New512_256()

	_, err := io.Copy(h, r)
	if err != nil {
		return nil, errors.Wrap(err, "GetKey: could not copy data in the hasher")
	}

	return h.Sum(nil), nil
}
