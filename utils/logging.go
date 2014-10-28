package utils

import (
	"log"
	"path/filepath"
	"runtime"
	"strings"
)

var basePath = "/home/cryptix/"

var (
	LogErr  = log.New(LogWriter{}, "ERROR: ", 0)
	LogWarn = log.New(LogWriter{}, "WARN: ", 0)
	LogInfo = log.New(LogWriter{}, "INFO: ", 0)
)

type LogWriter struct{}

func (f LogWriter) Write(p []byte) (n int, err error) {
	pc, file, line, ok := runtime.Caller(4)
	if !ok {
		file = "?"
		line = 0
	} else {
		file = strings.TrimPrefix(file, basePath)
	}
	fn := runtime.FuncForPC(pc)
	var fnName string
	if fn == nil {
		fnName = "?()"
	} else {
		dotName := filepath.Ext(fn.Name())
		fnName = strings.TrimLeft(dotName, ".") + "()"
	}
	log.Printf("%s:%d %s\n", file, line, fnName)
	log.Print(string(p))

	return len(p), nil
}
