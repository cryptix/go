package utils

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var closeChan chan<- os.Signal

func SetCloseChan(c chan<- os.Signal) {
	closeChan = c
}

// CheckFatal if err != nil: log.Fatalf() is called with file and line information from runtime.Caller()
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
		log.Printf("%s:%d %s\n", file, line, fnName)
		log.Print("Fatal Error:", err.Error())
		if closeChan != nil {
			log.Println("Sending close message")
			closeChan <- os.Interrupt
		}
		log.Fatal("Stopping")
	}
}
