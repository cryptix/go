package countconn_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cryptix/go/logging/countconn"
)

func ExampleNewReader() {
	r := strings.NewReader("hello, world\n")
	rc := countconn.NewReader(r)

	n, err := io.Copy(os.Stdout, rc)
	fmt.Println(err)
	fmt.Println(n)
	fmt.Println(rc.N())

	// Output: hello, world
	// <nil>
	// 13
	// 13
}

func ExampleNewWriter() {
	var w bytes.Buffer
	wc := countconn.NewWriter(&w)

	n, err := io.Copy(wc, strings.NewReader("hello, world\n"))
	fmt.Print(w.String())
	fmt.Println(err)
	fmt.Println(n)
	fmt.Println(wc.N())

	// Output: hello, world
	// <nil>
	// 13
	// 13
}
