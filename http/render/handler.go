package render

import (
	"fmt"
	"net/http"
)

type Binary func(resp http.ResponseWriter, req *http.Request) error

func (h Binary) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Description", "File Transfer")
	resp.Header().Set("Content-Transfer-Encoding", "binary")
	if err := h(resp, req); err != nil {
		logError(req, err, nil)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
	}
}

// Html wrapps a htpp.HandlerFunc-like function with an error return value
type Html func(resp http.ResponseWriter, req *http.Request) error

func (h Html) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if Reload {
		Load()
	}

	if err := h(resp, req); err != nil {
		logError(req, err, nil)
		handleError(resp, req, http.StatusInternalServerError, err)
	}
}

func handleError(w http.ResponseWriter, r *http.Request, status int, err error) {
	w.Header().Set("cache-control", "no-cache")
	err2 := Render(w, r, "error.tmpl", status, map[string]interface{}{
		"StatusCode": status,
		"Status":     http.StatusText(status),
		"Err":        err,
	})
	if err2 != nil {
		logError(r, fmt.Errorf("during execution of error template: %s", err2), nil)
	}
}
