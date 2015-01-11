package logging

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	logpkg "github.com/cryptix/go-logging"
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
		l.Criticalf("%s:%d %s", file, line, fnName)
		l.Critical("Fatal Error:", err.Error())
		if closeChan != nil {
			l.Debug("Sending close message")
			closeChan <- os.Interrupt
		}
		os.Exit(1)
	}
}

var (
	l = Logger("logging")

	// loggers is the set of loggers in the system
	loggers = map[string]*logpkg.Logger{}
)

// LogFormats is a map of formats used for our logger, keyed by name.
var LogFormats = map[string]string{
	"nocolor": "%{time:2006-01-02 15:04:05.000000} %{level} %{module} %{shortfile}: %{message}",
	"color":   ansiGray + "%{time:15:04:05.000} %{color}%{level:5.5s} " + ansiBlue + "%{module:10.10s}: %{color:reset}%{message} " + ansiGray + "%{shortfile}%{color:reset}",
}

const (
	// Logging environment variables
	envLoggingFmt = "CRYPTIX_LOGGING_FMT"

	// ansi colorcode constants
	ansiGray = "\033[0;37m"
	ansiBlue = "\033[0;34m"

	defaultLogFormat = "color"
)

// SetupLogging will initialize the logger backend and set the flags.
func SetupLogging(w io.Writer) {

	fmt := LogFormats[os.Getenv(envLoggingFmt)]
	if fmt == "" {
		fmt = LogFormats[defaultLogFormat]
	}

	// only output warnings and above to stderr
	var vis logpkg.Backend
	vis = logpkg.NewLogBackend(os.Stderr, "", 0)
	if w != nil {
		fileBackend := logpkg.NewLogBackend(w, "", 0)
		fileLog := logpkg.NewBackendFormatter(fileBackend, logpkg.MustStringFormatter(LogFormats["nocolor"]))
		vis = logpkg.MultiLogger(vis, fileLog)
	}
	leveld := logpkg.AddModuleLevel(vis)
	leveld.SetLevel(logpkg.NOTICE, "")

	logpkg.SetBackend(leveld)
	logpkg.SetFormatter(logpkg.MustStringFormatter(fmt))
}

// Logger retrieves a particular logger
func Logger(name string) *logpkg.Logger {
	log := logpkg.MustGetLogger(name)
	loggers[name] = log
	return log
}
