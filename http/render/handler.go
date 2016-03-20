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

	"github.com/rs/xlog"
	"golang.org/x/net/context"
	"gopkg.in/errgo.v1"
)

// Binary sets Content-Description and Content-Transfer-Encoding
// if h returns an error it returns http status 500
type Binary func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error

func (h Binary) ServeHTTPC(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Description", "File Transfer")
	resp.Header().Set("Content-Transfer-Encoding", "binary")
	if err := h(ctx, resp, req); err != nil {
		fmt.Fprintf(resp, "Error serving %s: %s", req.URL, err)
		xlog.FromContext(ctx).Error(err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
	}
}

// PlainError helps rendering user errors
func PlainError(ctx context.Context, w http.ResponseWriter, statusCode int, err error) {
	xlog.FromContext(ctx).Error("PlainError", xlog.F{
		"status": statusCode,
		"err":    errgo.Details(err),
	})
	http.Error(w, err.Error(), statusCode)
}
