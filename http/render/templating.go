package render

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	"github.com/cryptix/go/logging"
	"github.com/go-kit/kit/log"
	"github.com/oxtoacart/bpool"
	"github.com/pkg/errors"
	"github.com/shurcooL/httpfs/html/vfstemplate"
)

// TODO: make interface
type Renderer struct {
	assets http.FileSystem
	log    log.Logger

	// files
	templateFiles []string
	baseTemplates []string

	funcMap template.FuncMap

	// bufpool is shared between all render() calls
	bufpool *bpool.BufferPool

	doReload bool // Reload is whether to reload templates on each request.

	mu        sync.RWMutex // protect concurrent map access
	reloading bool
	templates map[string]*template.Template
}

// New creates a new Renderer
func New(fs http.FileSystem, opts ...Option) (*Renderer, error) {
	r := &Renderer{
		assets:    fs,
		bufpool:   bpool.NewBufferPool(64),
		templates: make(map[string]*template.Template),
	}

	for i, o := range opts {
		if err := o(r); err != nil {
			return nil, errors.Wrapf(err, "render: option %d failed.", i)
		}
	}

	// todo defaults
	if r.log == nil {
		r.log = logging.Logger("render")
	}

	if len(r.baseTemplates) == 0 {
		r.baseTemplates = []string{"base.tmpl"}
	}
	return r, r.parseHTMLTemplates()
}

func (r *Renderer) GetReloader() func(http.Handler) http.Handler {
	r.doReload = true
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if err := r.Reload(); err != nil {
				err = errors.Wrapf(err, "render: could not reload templates")
				r.log.Log("event", "error", "msg", "reload failed", "err", err)
				r.Error(rw, req, http.StatusInternalServerError, err)
				return
			}
			next.ServeHTTP(rw, req)
		})
	}
}

func (r *Renderer) Reload() error {
	if r.doReload {
		r.mu.RLock()
		if r.reloading {
			r.mu.RUnlock()
			return nil
		}
		r.mu.RUnlock()
		return r.parseHTMLTemplates()
	}
	return nil
}

type RenderFunc func(w http.ResponseWriter, req *http.Request) (interface{}, error)

func (r *Renderer) HTML(name string, f RenderFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		data, err := f(w, req)
		if err != nil {
			r.log.Log("event", "error", "msg", "handler failed", "err", err)
			r.Error(w, req, http.StatusInternalServerError, err)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		if err := r.Render(w, req, name, http.StatusOK, data); err != nil {
			r.log.Log("event", "error", "msg", "HTML render failed", "err", err)
			r.Error(w, req, http.StatusInternalServerError, err)
			return
		}
	}
}

func (r *Renderer) StaticHTML(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := r.Render(w, req, name, http.StatusOK, nil)
		if err != nil {
			r.log.Log("event", "error", "msg", "static HTML failed", "err", err)
			r.Error(w, req, http.StatusInternalServerError, err)
		}
	})
}

func (r *Renderer) Render(w http.ResponseWriter, req *http.Request, name string, status int, data interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.templates[name]
	if !ok {
		return errors.Errorf("render: could not find template: %s", name)
	}
	start := time.Now()
	l := log.With(r.log, "tpl", name)
	buf := r.bufpool.Get()
	err := t.ExecuteTemplate(buf, filepath.Base(r.baseTemplates[0]), data)
	if err != nil {
		return errors.Wrapf(err, "render: template(%s) execution failed.", name)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, err = buf.WriteTo(w)
	r.bufpool.Put(buf)
	l.Log("level", "debug", "event", "rendered",
		"name", name,
		"status", status,
		"took", time.Since(start),
	)
	return err
}

func (r *Renderer) Error(w http.ResponseWriter, req *http.Request, status int, err error) {
	r.logError(req, err, nil)
	w.Header().Set("cache-control", "no-cache")
	err2 := r.Render(w, req, "/error.tmpl", status, map[string]interface{}{
		"StatusCode": status,
		"Status":     http.StatusText(status),
		"Err":        err,
	})
	if err2 != nil {
		err2 = errors.Wrap(err2, "render: during execution of error template")
		err = errors.Wrap(err, err2.Error())
		r.logError(req, err, nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (r *Renderer) parseHTMLTemplates() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reloading = true
	funcTpl := template.New("").Funcs(r.funcMap)
	for _, tf := range r.templateFiles {
		ftc, err := funcTpl.Clone()
		if err != nil {
			return errors.Wrapf(err, "render: could not clone func template")
		}
		t, err := vfstemplate.ParseFiles(r.assets, ftc, append(r.baseTemplates, tf)...)
		if err != nil {
			return errors.Wrapf(err, "render: failed to parse template %s", tf)
		}
		r.templates[tf] = t
	}
	r.reloading = false
	return nil
}

func (r *Renderer) logError(req *http.Request, err error, rv interface{}) {
	if err != nil {
		buf := r.bufpool.Get()
		fmt.Fprintf(buf, "Error serving %s: %s", req.URL, err)
		if rv != nil {
			fmt.Fprintln(buf, rv)
			buf.Write(debug.Stack())
		}
		r.log.Log("event", "error", "msg", "logError", "err", err)
		r.bufpool.Put(buf)
	}
}
