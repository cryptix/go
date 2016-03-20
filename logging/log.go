package logging

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rs/xlog"
	"gopkg.in/errgo.v1"
)

var (
	conf *xlog.Config

	closeChan chan<- os.Signal
)

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
		xlog.Errorf("%s:%d %s", file, line, fnName)
		xlog.Error("Fatal Error:", errgo.Details(err))
		if closeChan != nil {
			xlog.Warn("Sending close message")
			closeChan <- os.Interrupt
		}
		os.Exit(1)
	}
}

// SetupLogging will initialize the logger backend and set the flags.
func SetupLogging(c *xlog.Config) {
	if conf != nil {
		xlog.Error("logging: initializing twice! skipping...")
		return
	}

	// defaults
	if c == nil {
		c = new(xlog.Config)
		c.Output = xlog.NewConsoleOutput()
	}

	c.Level = xlog.LevelInfo
	if lvl := os.Getenv("CRYPTIX_LOGLVL"); lvl != "" {
		var ok bool
		c.Level, ok = map[string]xlog.Level{
			"debug": xlog.LevelDebug,
			"info":  xlog.LevelInfo,
			"error": xlog.LevelError,
		}[strings.ToLower(lvl)]
		if !ok {
			xlog.Warn("logging: could not match loglvl from env, defaulting to debug")
			c.Level = xlog.LevelDebug
		}
	}
	conf = c
	l := xlog.New(*conf)
	// Plug the xlog handler's input to Go's default logger
	log.SetFlags(0)
	log.SetOutput(l)

	xlog.SetLogger(l)
}

// Logger returns an Entry where the module field is set to name
func Logger(name string) xlog.Logger {
	if name == "" {
		xlog.Warn("loging: missing name parameter")
		name = "undefined"
	}
	var thisConf = conf
	if thisConf == nil {
		xlog.Warn("logging: not initialized yet.", xlog.F{"name": name})
		thisConf = &xlog.Config{
			Output: xlog.NewConsoleOutput(),
		}
	}
	l := xlog.New(*thisConf)
	l.SetField("unit", name)
	return l
}
