package debug

import "io"

type RWC struct {
	io.Reader
	io.Writer
	c io.Closer
}

func WrapRWC(c io.ReadWriteCloser) io.ReadWriteCloser {
	rl := NewReadLogger("<", c)
	wl := NewWriteLogger(">", c)

	return &RWC{
		Reader: rl,
		Writer: wl,
		c:      c,
	}
}

func (c *RWC) Close() error {
	return c.c.Close()
}
