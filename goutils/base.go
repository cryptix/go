package goutils

import (
	"go/build"
	"os"
	"path/filepath"

	"go.mindeco.de/toolbelt/logging"
	"github.com/pkg/errors"
)

// LocatePackage searches for the import path and returns the os filesystem path location
func LocatePackage(path string) (string, error) {
	p, err := build.Default.Import(path, "", build.FindOnly)
	if err != nil {
		return "", errors.Wrap(err, "LocatePackage: failed to find import")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "LocatePackage: could not get working directory")
	}

	p.Dir, err = filepath.Rel(cwd, p.Dir)
	if err != nil {
		return "", errors.Wrap(err, "LocatePackage: could not construct relative path")
	}

	return p.Dir, nil
}

func MustLocatePackage(path string) string {
	p, err := LocatePackage(path)
	logging.CheckFatal(err)

	return p
}
