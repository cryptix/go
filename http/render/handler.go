/*
Package render implements template inheritance and exposes functions to render HTML.

inspired by http://elithrar.github.io/article/approximating-html-template-inheritance and https://github.com/sourcegraph/thesrc/blob/master/app/handler.go

It also exports two types Binary and HMTL.
Both wrap a http.HandlerFunc-like function with an error return value and argument the response.
*/
package render

import (
	"log"
	"net/http"

	"github.com/pkg/errors"
)

// Binary sets Content-Description and Content-Transfer-Encoding
// if h returns an error it returns http status 500
type Binary func(resp http.ResponseWriter, req *http.Request) error

func (h Binary) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Description", "File Transfer")
	resp.Header().Set("Content-Transfer-Encoding", "binary")
	if err := h(resp, req); err != nil {
		err = errors.Wrapf(err, "render/binary: handler error serving %s", req.URL)
		// TOODO: request injection
		log.Println("Binary/ServeHTTP:", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
	}
}

// PlainError helps rendering user errors
func PlainError(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	log.Println("PlainError:", err)
	// TOODO: request injection
	// xlog.FromContext(r.Context()).Error("PlainError", xlog.F{
	// 	"status": statusCode,
	// 	"err":    errgo.Details(err),
	// })
	http.Error(w, err.Error(), statusCode)
}
