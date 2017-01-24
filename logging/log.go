package logging

import (
	"fmt"
	"io"
	"os"

	"github.com/go-kit/kit/log"
	"gopkg.in/errgo.v1"
)

var closeChan chan<- os.Signal

func SetCloseChan(c chan<- os.Signal) {
	closeChan = c
}

// CheckFatal exits the process if err != nil
func CheckFatal(err error) {
	if err != nil {
		l := Underlying
		if l == nil {
			l = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
			l = log.NewContext(l).With("module", "logging", log.DefaultCaller)
		}

		l.Log("check", "fatal", "err", errgo.Details(err))
		if closeChan != nil {
			l.Log("check", "notice", "msg", "Sending close message")
			closeChan <- os.Interrupt
		}
		os.Exit(1)
	}
}

var Underlying log.Logger

// SetupLogging will initialize the logger backend and set the flags.
func SetupLogging(w io.Writer) {
	if w != nil {
		w = io.MultiWriter(os.Stderr, w)
	}
	logger := log.NewLogfmtLogger(w)
	if lvl := os.Getenv("CRYPTIX_LOGLVL"); lvl != "" {
		logger.Log("module", "logging", "error", "CRYPTIX_LOGLVL is obsolete. levels are bad, mkay?")
	}
	// wrap logger to error-check the writes only once
	Underlying = log.LoggerFunc(func(keyvals ...interface{}) error {
		if err := logger.Log(keyvals...); err != nil {
			fmt.Fprintf(os.Stderr, "warning: logger.Write() failed! %s", err)
			panic(err) // no other way to escalate this
		}
		return nil
	})
}

// Logger returns an Entry where the module field is set to name
func Logger(name string) *log.Context {
	if name == "" {
		Underlying.Log("module", "logger", "error", "missing name parameter")
		name = "undefined"
	}

	return log.NewContext(Underlying).With("ts", log.DefaultTimestamp, "caller", log.DefaultCaller, "module", name)
}
