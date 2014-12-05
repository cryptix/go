package logging

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	logpkg "github.com/jbenet/go-logging"
)

func init() {
	SetupLogging()
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
		log.Criticalf("%s:%d %s", file, line, fnName)
		log.Critical("Fatal Error:", err.Error())

		os.Exit(1)
	}
}

// ErrNoSuchLogger is returned when the util pkg is asked for a non existant logger
var ErrNoSuchLogger = errors.New("Error: No such logger")

var log = Logger("util")

var ansiGray = "\033[0;37m"
var ansiBlue = "\033[0;34m"

// LogFormats is a map of formats used for our logger, keyed by name.
var LogFormats = map[string]string{
	"nocolor": "%{time:2006-01-02 15:04:05.000000} %{level} %{module} %{shortfile}: %{message}",
	"color": ansiGray + "%{time:15:04:05.000} %{color}%{level:5.5s} " + ansiBlue +
		"%{module:10.10s}: %{color:reset}%{message} " + ansiGray + "%{shortfile}%{color:reset}",
}
var defaultLogFormat = "color"

// Logging environment variables
const (
	envLogging    = "CRYPTIX_LOGGING"
	envLoggingFmt = "CRYPTIX_LOGGING_FMT"
)

// loggers is the set of loggers in the system
var loggers = map[string]*logpkg.Logger{}

// POut is a shorthand printing function to output to Stdout.
func POut(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, format, a...)
}

// SetupLogging will initialize the logger backend and set the flags.
func SetupLogging() {

	fmt := LogFormats[os.Getenv(envLoggingFmt)]
	if fmt == "" {
		fmt = LogFormats[defaultLogFormat]
	}

	backend := logpkg.NewLogBackend(os.Stderr, "", 0)
	logpkg.SetBackend(backend)
	logpkg.SetFormatter(logpkg.MustStringFormatter(fmt))

	lvl := logpkg.NOTICE

	if logenv := os.Getenv(envLogging); logenv != "" {
		var err error
		lvl, err = logpkg.LogLevel(logenv)
		if err != nil {
			log.Errorf("logpkg.LogLevel() Error: %q", err)
			lvl = logpkg.ERROR // reset to ERROR, could be undefined now(?)
		}
	}

	SetAllLoggers(lvl)

}

// SetAllLoggers changes the logpkg.Level of all loggers to lvl
func SetAllLoggers(lvl logpkg.Level) {
	logpkg.SetLevel(lvl, "")
	for i := range loggers {
		logpkg.SetLevel(lvl, i)
	}
}

// Logger retrieves a particular logger
func Logger(name string) *logpkg.Logger {
	log := logpkg.MustGetLogger(name)
	loggers[name] = log
	return log
}

// SetLogLevel changes the log level of a specific subsystem
// name=="*" changes all subsystems
func SetLogLevel(name, level string) error {
	lvl, err := logpkg.LogLevel(level)
	if err != nil {
		return err
	}

	// wildcard, change all
	if name == "*" {
		SetAllLoggers(lvl)
		return nil
	}

	// Check if we have a logger by that name
	// logpkg.SetLevel() can't tell us...
	_, ok := loggers[name]
	if !ok {
		return ErrNoSuchLogger
	}

	logpkg.SetLevel(lvl, name)

	return nil
}
