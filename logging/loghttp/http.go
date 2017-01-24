package loghttp

import (
	"net/http"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/go-kit/kit/log"
)

type HTTPLogger struct {
	log.Logger
}

func NewNegroni(l log.Logger) *HTTPLogger {
	return &HTTPLogger{l}
}

func NewNegroniWithName(l log.Logger, name string) *HTTPLogger {
	return &HTTPLogger{log.NewContext(l).With("module", name)}
}

func (l *HTTPLogger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	next(rw, r)
	res := rw.(negroni.ResponseWriter)
	l.Log("method", r.Method, "path", r.URL.Path, "status", res.Status(), "took", time.Since(start))
}
