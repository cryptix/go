package httpHandler

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/cryptix/go/utils"
)

type Binary func(resp http.ResponseWriter, req *http.Request) error

func (h Binary) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Description", "File Transfer")
	resp.Header().Set("Content-Transfer-Encoding", "binary")
	runBinaryHandler(resp, req, h)
}

func runBinaryHandler(w http.ResponseWriter, r *http.Request, fn func(http.ResponseWriter, *http.Request) error) {
	var err error

	defer func() {
		if rv := recover(); rv != nil {
			err = errors.New("Html panic")
			logError(r, err, rv)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}()

	err = fn(w, r)
	if err != nil {
		logError(r, err, nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Html wrapps a htpp.HandlerFunc-like function with an error return value
type Html func(resp http.ResponseWriter, req *http.Request) error

func (h Html) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if Reload {
		Load()
	}
	runHandler(resp, req, h)
}

func runHandler(w http.ResponseWriter, r *http.Request, fn func(http.ResponseWriter, *http.Request) error) {
	var err error

	defer func() {
		if rv := recover(); rv != nil {
			err = errors.New("Html panic")
			logError(r, err, rv)
			handleError(w, r, http.StatusInternalServerError, err)
		}
	}()

	err = fn(w, r)
	if err != nil {
		logError(r, err, nil)
		handleError(w, r, http.StatusInternalServerError, err)
	}
}

func handleError(w http.ResponseWriter, r *http.Request, status int, err error) {
	w.Header().Set("cache-control", "no-cache")
	err2 := Render(w, r, "error.tmpl", status, &struct {
		StatusCode int
		Status     string
		Err        error
	}{
		StatusCode: status,
		Status:     http.StatusText(status),
		Err:        err,
	})
	if err2 != nil {
		logError(r, fmt.Errorf("during execution of error template: %s", err2), nil)
	}
}

func logError(req *http.Request, err error, rv interface{}) {
	if err != nil {
		buf := bufpool.Get()
		fmt.Fprintf(buf, "Error serving %s: %s\n", req.URL, err)
		if rv != nil {
			fmt.Fprintln(buf, rv)
			buf.Write(debug.Stack())
		}
		utils.LogErr.Println(buf.String())
		bufpool.Put(buf)
	}
}
