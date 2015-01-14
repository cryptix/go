/*
Package render implements template inheritance and exposes functions to render HTML.

inspired by http://elithrar.github.io/article/approximating-html-template-inheritance and https://github.com/sourcegraph/thesrc/blob/master/app/handler.go

It also exports two types Binary and HMTL.
Both wrap a http.HandlerFunc-like function with an error return value and argument the response.
*/
package render

import (
	"fmt"
	"net/http"
)

// Binary sets Content-Description and Content-Transfer-Encoding
// if h returns an error it returns http status 500
type Binary func(resp http.ResponseWriter, req *http.Request) error

func (h Binary) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Description", "File Transfer")
	resp.Header().Set("Content-Transfer-Encoding", "binary")
	if err := h(resp, req); err != nil {
		logError(req, err, nil)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
	}
}

// HTML expects render.Render to be called in it's wrapped handler.
// if h returns an error, RenderError is called to render it.
type HTML func(resp http.ResponseWriter, req *http.Request) error

func (h HTML) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if Reload {
		Load()
	}

	// resp.Header().Set("Content-Type","text/html")
	if err := h(resp, req); err != nil {
		logError(req, err, nil)
		Error(resp, req, http.StatusInternalServerError, err)
	}
}

// StaticHTML just renders a template without any extra data
type StaticHTML string

func (tpl StaticHTML) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if Reload {
		Load()
	}

	err := Render(resp, req, string(tpl), http.StatusOK, nil)
	if err != nil {
		logError(req, err, nil)
		Error(resp, req, http.StatusInternalServerError, err)
	}
}

// Error uses 'error.tmpl' to output an error in HTML format
func Error(w http.ResponseWriter, r *http.Request, status int, err error) {
	w.Header().Set("cache-control", "no-cache")
	err2 := Render(w, r, "error.tmpl", status, map[string]interface{}{
		"StatusCode": status,
		"Status":     http.StatusText(status),
		"Err":        err,
	})
	if err2 != nil {
		logError(r, fmt.Errorf("during execution of error template: %s", err2), nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
