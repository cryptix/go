package render

import (
	"fmt"
	"html/template"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/oxtoacart/bpool"
	"github.com/rs/xhandler"
	"github.com/rs/xlog"
	"github.com/shurcooL/httpfs/html/vfstemplate"
	"golang.org/x/net/context"
	"gopkg.in/errgo.v1"
)

type Renderer struct {
	assets http.FileSystem

	// files
	templateFiles []string
	baseTemplate  string

	funcMap template.FuncMap

	// bufpool is shared between all render() calls
	bufpool *bpool.BufferPool

	doReload bool // Reload is whether to reload templates on each request.

	mu        sync.RWMutex // protect concurrent map access
	reloading bool
	templates map[string]*template.Template
}

// New creates a new Renderer
func New(fs http.FileSystem, base string, opts ...Option) (*Renderer, error) {
	r := &Renderer{
		assets:       fs,
		baseTemplate: base,
		bufpool:      bpool.NewBufferPool(64),
		templates:    make(map[string]*template.Template),
	}
	for i, o := range opts {
		if err := o(r); err != nil {
			return nil, errgo.Notef(err, "render: option %i failed.", i)
		}
	}
	return r, r.parseHTMLTemplates()
}

func (r *Renderer) GetReloader() func(xhandler.HandlerC) xhandler.HandlerC {
	r.doReload = true
	return func(next xhandler.HandlerC) xhandler.HandlerC {
		return xhandler.HandlerFuncC(func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
			if err := r.Reload(); err != nil {
				err = errgo.Notef(err, "could not parse template")
				r.Error(ctx, rw, req, http.StatusInternalServerError, err)
				return
			}
			next.ServeHTTPC(ctx, rw, req)
		})
	}
}

func (r *Renderer) Reload() error {
	if r.doReload {
		r.mu.RLock()
		if r.reloading {
			return nil
		}
		r.mu.RUnlock()
		return r.parseHTMLTemplates()
	}
	return nil
}

type RenderFunc func(ctx context.Context, w http.ResponseWriter, req *http.Request) (interface{}, error)

func (r *Renderer) HTML(name string, f RenderFunc) xhandler.HandlerC {
	return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		data, err := f(ctx, w, req)
		if err != nil {
			r.Error(ctx, w, req, http.StatusInternalServerError, err)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		if err := r.Render(ctx, w, req, name, http.StatusOK, data); err != nil {
			r.Error(ctx, w, req, http.StatusInternalServerError, err)
			return
		}
	})
}

func (r *Renderer) StaticHTML(name string) xhandler.HandlerC {
	return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		err := r.Render(ctx, w, req, name, http.StatusOK, nil)
		if err != nil {
			r.Error(ctx, w, req, http.StatusInternalServerError, err)
		}
	})
}

func (r *Renderer) Render(ctx context.Context, w http.ResponseWriter, req *http.Request, name string, status int, data interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	l := xlog.FromContext(ctx)
	t, ok := r.templates[name]
	if !ok {
		return errgo.New("Could not find template:" + name)
	}
	start := time.Now()
	l.SetField("tpl", name)
	buf := r.bufpool.Get()
	err := t.ExecuteTemplate(buf, r.baseTemplate, data)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, err = buf.WriteTo(w)
	r.bufpool.Put(buf)
	xlog.FromContext(ctx).Debug("Rendered", xlog.F{
		"name":   name,
		"status": status,
		"took":   time.Since(start),
	})
	return err
}

func (r *Renderer) Error(ctx context.Context, w http.ResponseWriter, req *http.Request, status int, err error) {
	r.logError(ctx, req, err, nil)
	w.Header().Set("cache-control", "no-cache")
	err2 := r.Render(ctx, w, req, "/error.tmpl", status, map[string]interface{}{
		"StatusCode": status,
		"Status":     http.StatusText(status),
		"Err":        err,
	})
	if err2 != nil {
		err = errgo.WithCausef(err, err2, "render: during execution of error template.")
		r.logError(ctx, req, err, nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (r *Renderer) parseHTMLTemplates() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reloading = true
	var err error
	funcTpl := template.New("").Funcs(r.funcMap)
	for _, tf := range r.templateFiles {
		ftc, err := funcTpl.Clone()
		if err != nil {
			return errgo.Notef(err, "render: could not clone func template")
		}
		t, err := vfstemplate.ParseFiles(r.assets, ftc, r.baseTemplate, tf)
		if err != nil {
			return errgo.Notef(err, "render: failed to parse template %s", tf)
		}
		r.templates[tf] = t
		xlog.Debug("parsed", xlog.F{
			"name": t.Name(),
			"tf":   tf,
		})
	}
	r.reloading = false
	return err
}

func (r *Renderer) logError(ctx context.Context, req *http.Request, err error, rv interface{}) {
	if err != nil {
		buf := r.bufpool.Get()
		fmt.Fprintf(buf, "Error serving %s: %s", req.URL, err)
		if rv != nil {
			fmt.Fprintln(buf, rv)
			buf.Write(debug.Stack())
		}
		xlog.FromContext(ctx).Error(buf.String())
		r.bufpool.Put(buf)
	}
}
