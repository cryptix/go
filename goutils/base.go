package goutils

import (
	"go/build"
	"os"
	"path/filepath"

	"github.com/cryptix/go/logging"
)

// LocatePackage searches for the import path and returns the os filesystem path location
func LocatePackage(path string) string {
	p, err := build.Default.Import(path, "", build.FindOnly)
	logging.CheckFatal(err)

	cwd, err := os.Getwd()
	logging.CheckFatal(err)

	p.Dir, err = filepath.Rel(cwd, p.Dir)
	logging.CheckFatal(err)

	return p.Dir
}
