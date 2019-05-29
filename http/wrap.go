package http

import (
	"net/http"

	kitlog "github.com/go-kit/kit/log"
)

type MiddlewareFunc func(next http.Handler) http.Handler

type HandlerFuncWithErr func(w http.ResponseWriter, r *http.Request) error

func WrapWithError(f HandlerFuncWithErr, log kitlog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			log.Log("event", "error", "msg", "could not serve HTTP request", "err", err, "path", r.URL.Path)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// Authorize does a very simple header check against a wanted value
// it returns http.StatusUnauthorized if it's false
func Authorize(headerName, wantHeader string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hv := r.Header.Get(headerName)
			if hv != wantHeader {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
