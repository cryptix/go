package logging

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"
	"gopkg.in/errgo.v1"
)

var closeChan chan<- os.Signal

func SetCloseChan(c chan<- os.Signal) {
	closeChan = c
}

// CheckFatal exits the process if err != nil
func CheckFatal(err error) {
	if err != nil {
		pc, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "?"
			line = 0
		}
		fn := runtime.FuncForPC(pc)
		var fnName string
		if fn == nil {
			fnName = "?()"
		} else {
			dotName := filepath.Ext(fn.Name())
			fnName = strings.TrimLeft(dotName, ".") + "()"
		}
		l.Errorf("%s:%d %s", file, line, fnName)
		l.Error("Fatal Error:", errgo.Details(err))
		if closeChan != nil {
			l.Warn("Sending close message")
			closeChan <- os.Interrupt
		}
		os.Exit(1)
	}
}

var l = logrus.New()

// SetupLogging will initialize the logger backend and set the flags.
func SetupLogging(w io.Writer) {
	if w != nil {
		l.Out = io.MultiWriter(os.Stderr, w)
	}
}

// Logger returns a logger where the module field is set to name
// https://github.com/Sirupsen/logrus/issues/144
func Logger(name string) *logrus.Entry {
	if len(name) == 0 {
		l.Warnf("missing name parameter")
		name = "undefined"
	}
	return l.WithField("module", name)
}
