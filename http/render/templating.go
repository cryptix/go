package render

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/oxtoacart/bpool"
	"github.com/shurcooL/httpfs/html/vfstemplate"
	"go.mindeco.de/logging"
)

type Renderer struct {
	assets http.FileSystem
	log    log.Logger

	// files
	templateFiles []string
	baseTemplates []string

	funcMap template.FuncMap

	tplFuncInjectors map[string]FuncInjector

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

		tplFuncInjectors: make(map[string]FuncInjector),
	}

	for i, o := range opts {
		if err := o(r); err != nil {
			return nil, fmt.Errorf("render: option %d failed: %w", i, err)
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
				level.Error(r.log).Log("event", "reload failed", "err", err)
				err = fmt.Errorf("render: could not reload templates: %w", err)
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
			level.Error(r.log).Log("event", "handler failed", "err", err)
			r.Error(w, req, http.StatusInternalServerError, err)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		if err := r.Render(w, req, name, http.StatusOK, data); err != nil {
			level.Error(r.log).Log("event", "HTML render failed", "err", err)
			r.Error(w, req, http.StatusInternalServerError, err)
			return
		}
	}
}

func (r *Renderer) StaticHTML(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := r.Render(w, req, name, http.StatusOK, nil)
		if err != nil {
			level.Error(r.log).Log("msg", "static HTML failed", "err", err)
			r.Error(w, req, http.StatusInternalServerError, err)
		}
	})
}

func (r *Renderer) Render(w http.ResponseWriter, req *http.Request, name string, status int, data interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("render: could not find template: %s", name)
	}

	// create request scoped functions
	var scopedFuncs = make(template.FuncMap, len(r.tplFuncInjectors))
	for name, fn := range r.tplFuncInjectors {
		scopedFuncs[name] = fn(req)
	}

	// need to clone the template to not bork it for future requests
	scopedTpl, err := t.Clone()
	if err != nil {
		return err
	}

	// assign the scoped functions
	scopedTpl = scopedTpl.Funcs(scopedFuncs)

	start := time.Now()
	l := log.With(r.log, "tpl", name)
	buf := r.bufpool.Get()

	err = scopedTpl.ExecuteTemplate(buf, filepath.Base(r.baseTemplates[0]), data)
	if err != nil {
		return fmt.Errorf("render: template(%s) execution failed: %w", name, err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	sz := buf.Len()
	_, err = buf.WriteTo(w)

	r.bufpool.Put(buf)
	level.Debug(l).Log("event", "rendered",
		"name", name,
		"status", status,
		"took", time.Since(start),
		"size", sz,
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
		err2 = fmt.Errorf("render: during execution of error template: %w", err2)
		err = fmt.Errorf("meant to return %s but ran into %w", err, err2)
		r.logError(req, err, nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (r *Renderer) parseHTMLTemplates() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reloading = true

	parseFuncs := make(template.FuncMap, len(r.funcMap)+len(r.tplFuncInjectors))
	for k, v := range r.funcMap {
		parseFuncs[k] = v
	}

	// these are just placeholders so that the functions are not undefined.
	// they are repaced in Render() after the template is cloned.
	for k, _ := range r.tplFuncInjectors {
		parseFuncs[k] = func(...interface{}) string { return k }
	}

	funcTpl := template.New("").Funcs(parseFuncs)

	for _, tf := range r.templateFiles {
		ftc, err := funcTpl.Clone()
		if err != nil {
			return fmt.Errorf("render: could not clone func template: %w", err)
		}
		t, err := vfstemplate.ParseFiles(r.assets, ftc, append(r.baseTemplates, tf)...)
		if err != nil {
			return fmt.Errorf("render: failed to parse template %s: %w", tf, err)
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
		level.Error(r.log).Log("event", "logError", "err", err)
		r.bufpool.Put(buf)
	}
}
