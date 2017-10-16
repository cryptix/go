package logging

import (
	"bufio"
	"io"

	"github.com/go-kit/kit/log"
)

func Writer(unit string, l log.Logger) io.WriteCloser {
	l = log.With(l, "unit", unit)
	pr, pw := io.Pipe()
	go func() {
		s := bufio.NewScanner(pr)
		for s.Scan() {
			l.Log("msg", s.Text())
		}
		if err := s.Err(); err != nil {
			l.Log("msg", "scanner error", "err", err)
		}
	}()
	return pw
}
