package render

import "html/template"

type Option func(*Renderer) error

// AddTemplates adds filenames for the next call to parseTempaltes
func AddTemplates(files ...string) Option {
	return func(r *Renderer) error {
		r.templateFiles = files
		return nil
	}
}

func FuncMap(m template.FuncMap) Option {
	return func(r *Renderer) error {
		r.funcMap = m
		return nil
	}
}
