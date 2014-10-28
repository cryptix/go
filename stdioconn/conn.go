package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type StdioConn struct {
	Out io.ReadCloser
	In  io.WriteCloser
}

func (s StdioConn) Close() (err error) {

	if err = s.In.Close(); err != nil {
		return err
	}

	if err = s.Out.Close(); err != nil {
		return err
	}

	return nil
}

func (s *StdioConn) Read(p []byte) (n int, err error) {
	// log.Println("called read")
	n, err = s.Out.Read(p)
	// log.Printf("read[%d] bytes '%s'", n, string(p))
	return
}

func (s *StdioConn) Write(p []byte) (n int, err error) {
	n, err = s.In.Write(p)
	// log.Printf("wrote[%d] '%s'", n, string(p))
	return
}

func StartStdioProcess(path string) (*StdioConn, error) {
	var (
		err  error
		cmd  = exec.Command(path)
		conn StdioConn
	)

	conn.Out, err = cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	conn.In, err = cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	go func() {
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			fmt.Fprintln(os.Stderr, "stdioproc err:", s.Text())
		}
		if err := s.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "stdioproc failed:", err)
		}
	}()

	go cmd.Wait()

	return &conn, nil
}
