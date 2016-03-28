package logtest

import (
	"bufio"
	"io"
	"testing"
)

// Logger logs every line it is written to it. t.Log(prefix: line)
func Logger(prefix string, t *testing.T) io.WriteCloser {
	pr, pw := io.Pipe()
	go func() {
		s := bufio.NewScanner(pr)
		for s.Scan() {
			t.Logf("%s: %q", prefix, s.Text())
		}
		if err := s.Err(); err != nil {
			t.Errorf("%s: scanner error:%s", prefix, err)
		}
	}()
	return pw
}
