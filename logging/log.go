package logging

import (
	"fmt"
	"io"
	"os"

	kitlog "github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

var closeChan chan<- os.Signal

// SetCloseChan sets a signal channel that is sent to when CheckFatal is used
func SetCloseChan(c chan<- os.Signal) {
	closeChan = c
}

// CheckFatal exits the process if err != nil
func CheckFatal(err error) {
	if err != nil {
		l := kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stderr))
		l = kitlog.With(l, "module", "logging", "caller", kitlog.DefaultCaller)
		l.Log("check", "fatal", "err", err)
		if e2 := LogPanicWithStack(l, "CheckFatal", err); e2 != nil {
			fmt.Fprintf(os.Stderr, "CheckFatal Error:\n%#v", err)
			panic(errors.Wrap(e2, "CheckFatal could not dump error"))
		}
		if closeChan != nil {
			l.Log("check", "notice", "msg", "Sending close message")
			closeChan <- os.Interrupt
		}
		os.Exit(1)
	}
}

type Interface interface {
	kitlog.Logger
}

type InfoLog struct {
	Interface
}

// NewDevelopment gives you a new information logging helper
func NewDevelopment(w io.Writer, opts ...Option) (*InfoLog, error) {
	var il InfoLog

	if w == nil {
		w = os.Stderr
		//? w = ioutil.Discard
	}

	logger := kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(w))

	// wrap logger to error-check the writes only once
	il.Interface = kitlog.LoggerFunc(func(keyvals ...interface{}) error {
		if err := logger.Log(keyvals...); err != nil {
			fmt.Fprintf(w, "warning: logger.Write() failed! %s - vals: %v", err, keyvals)
			closeChan <- os.Interrupt
			panic(err) // no other way to escalate this
		}
		return nil
	})

	for i, o := range opts {
		err := o(&il)
		if err != nil {
			return nil, errors.Wrapf(err, "logging: NewDevelopment option %d failed", i)
		}
	}

	return &il, nil
}

// Logger returns an Entry where the module field is set to name
func (il InfoLog) Module(name string) Interface {
	if name == "" {
		il.Log("module", "logger", "error", "missing name parameter")
		name = "undefined"
	}
	return kitlog.With(il, "module", name)
}
