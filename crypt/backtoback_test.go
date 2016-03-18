package crypt

import (
	"bytes"
	"crypto/rand"
	"io"
	"io/ioutil"
	"testing"
)

func TestBackToBack(t *testing.T) {
	r := io.LimitReader(rand.Reader, 1024*1024)
	indata, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	key, err := GetKey(bytes.NewReader(indata))
	if err != nil {
		t.Fatal(err)
	}

	e, err := NewCrypter(key)
	if err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	cipherW, err := e.MakePipe(buf)
	if err != nil {
		t.Fatal(err)
	}

	_, err = io.Copy(cipherW, bytes.NewReader(indata))
	if err != nil {
		t.Fatal(err)
	}

	if buf.Len() != len(indata) {
		t.Fatalf("didn't write enough to the ciphertext buffer. got:%d", buf.Len())
	}

	d, err := NewCrypter(key)
	if err != nil {
		t.Fatal(err)
	}

	clearBuf := new(bytes.Buffer)
	clearW, err := d.MakePipe(clearBuf)
	if err != nil {
		t.Fatal(err)
	}

	_, err = io.Copy(clearW, buf)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(indata, clearBuf.Bytes()) != 0 {
		t.Fatal("didn't decrypt data correctly")
	}
}
