package logging

import (
	"net/http"
	"time"

	"github.com/codegangsta/negroni"
	logpkg "github.com/cryptix/go-logging"
)

type HTTPLogger struct {
	*logpkg.Logger
}

func NewNegroni(name string) *HTTPLogger {
	return &HTTPLogger{Logger(name)}
}

func (l *HTTPLogger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	l.Infof("Started %s %s", r.Method, r.URL.Path)

	next(rw, r)

	res := rw.(negroni.ResponseWriter)
	l.Infof("Completed %v %s in %v", res.Status(), http.StatusText(res.Status()), time.Since(start))
}
