package logging

import (
	"bytes"
	stdlog "log"
	"os"
	"strings"
	"testing"

	kitlog "github.com/go-kit/kit/log"
)

func ExampleNewDevelopment() {
	i, err := NewDevelopment(os.Stdout)
	checkPanic(err)
	i.Log("hello", "world")
	//Output:
	//hello=world
}

func ExampleWith() {
	i, err := NewDevelopment(os.Stdout, With("meta", "info", "caller", kitlog.DefaultCaller))
	checkPanic(err)
	i.Log("hello", "world")
	//Output:
	//meta=info caller=log_test.go:24 hello=world
}

func ExampleModule() {
	i, err := NewDevelopment(os.Stdout)
	checkPanic(err)
	i.Log("hello", "world")
	m := i.Module("foo")
	m.Log("hello", "again")
	//Output:
	//hello=world
	//module=foo hello=again
}

func TestToStdlib(t *testing.T) {
	var buf bytes.Buffer
	i, err := NewDevelopment(&buf, ToStdlib())
	checkPanic(err)
	i.Log("hello", "world")
	stdlog.Printf("throughStdlib")

	out := buf.String()
	if !strings.Contains(out, "hello=world\n") {
		t.Error("missing hello=world")
	}

	if !strings.Contains(out, "msg=throughStdlib\n") {
		t.Error("missing stdlib log")
	}
}

func checkPanic(err error) {
	if err != nil {
		panic(err)
	}
}
