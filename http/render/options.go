package render

import (
	"html/template"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

type Option func(*Renderer) error

// AddTemplates adds filenames for the next call to parseTempaltes
func AddTemplates(files ...string) Option {
	return func(r *Renderer) error {
		if len(files) == 0 {
			return errors.New("render: no templates passed")
		}
		r.templateFiles = files
		return nil
	}
}

func BaseTemplates(bases ...string) Option {
	return func(r *Renderer) error {
		r.baseTemplates = bases
		return nil
	}
}

func FuncMap(m template.FuncMap) Option {
	return func(r *Renderer) error {
		r.funcMap = m
		return nil
	}
}

func SetLogger(l log.Logger) Option {
	return func(r *Renderer) error {
		if l == nil {
			return errors.New("render: nil logger passed")
		}
		r.log = l
		return nil
	}
}
