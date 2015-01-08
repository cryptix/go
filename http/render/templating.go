package render

import (
	"errors"
	"fmt"
	htmpl "html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/cryptix/go/logging"
	"github.com/gorilla/mux"
	"github.com/oxtoacart/bpool"
)

type assetFunc func(name string) ([]byte, error)

var (
	// Reload is whether to reload templates on each request.
	Reload bool

	log = logging.Logger("render")

	// asset
	asset assetFunc

	// files
	templateFiles     []string
	baseTemplateFiles []string

	// all the templates that we parsed
	templates = map[string]*htmpl.Template{}

	// bufpool is shared between all render() calls
	bufpool = bpool.NewBufferPool(64)

	appRouter *mux.Router
)

// Init takes a go-bindata Asset function and base tempaltes, which are used to render other templates
func Init(fn assetFunc, base []string) {
	asset = fn
	baseTemplateFiles = append(baseTemplateFiles, base...)
}

// AddTemplates adds filenames for the next call to parseTempaltes
func AddTemplates(files []string) {
	templateFiles = append(templateFiles, files...)
}

// SetAppRouter is used to specify toe mux.Router, it's needed for the {{urlTo}} template func
func SetAppRouter(r *mux.Router) {
	appRouter = r
}

// Load loads and parses all templates that are in the assetFunc
func Load() {
	if appRouter == nil {
		logging.CheckFatal(errors.New("No appRouter set"))
	}

	if len(baseTemplateFiles) == 0 {
		logging.CheckFatal(errors.New("No base tempaltes"))
		// baseTemplateFiles = []string{"navbar.tmpl", "base.tmpl"}
	}

	logging.CheckFatal(parseHTMLTemplates())
}

func parseHTMLTemplates() error {
	for _, file := range templateFiles {
		t := htmpl.New("")
		t.Funcs(htmpl.FuncMap{
			"urlTo": urlTo,
			"itoa":  strconv.Itoa,
		})

		err := parseFilesFromBindata(t, file)
		if err != nil {
			return fmt.Errorf("template %v: %s", file, err)
		}

		t = t.Lookup("base")
		if t == nil {
			return fmt.Errorf("base template not found in %v", file)
		}
		templates[strings.TrimPrefix(file, "tmpl/")] = t
	}
	return nil
}

// Render takes a template name and any kind of named data
// renders the template to a buffer from the pool
// and writes that to the http response
func Render(w http.ResponseWriter, r *http.Request, name string, status int, data interface{}) error {
	tmpl, ok := templates[name]
	if !ok {
		return errors.New("Could not find template:" + name)
	}
	start := time.Now()

	buf := bufpool.Get()
	err := tmpl.ExecuteTemplate(buf, "base", data)
	if err != nil {
		return err
	}

	start = time.Now()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, err = buf.WriteTo(w)
	bufpool.Put(buf)
	log.Infof("Rendered %q Status:%d (took %v)", name, status, time.Since(start))
	return err
}

// PlainError helps rendering user errors
func PlainError(w http.ResponseWriter, statusCode int, err error) {
	log.Errorf("PlainError(%d):%s\n", statusCode, err)
	http.Error(w, err.Error(), statusCode)
}

func logError(req *http.Request, err error, rv interface{}) {
	if err != nil {
		buf := bufpool.Get()
		fmt.Fprintf(buf, "Error serving %s: %s", req.URL, err)
		if rv != nil {
			fmt.Fprintln(buf, rv)
			buf.Write(debug.Stack())
		}
		log.Error(buf.String())
		bufpool.Put(buf)
	}
}

// copied from template.ParseFiles but dont use ioutil.ReadFile
func parseFilesFromBindata(t *htmpl.Template, file string) error {
	var err error

	files := make([]string, len(baseTemplateFiles)+1)
	files[0] = file
	copy(files[1:], baseTemplateFiles)
	log.Debugf("parseFile - %q", files)

	for _, filename := range files {
		var tmplBytes []byte
		tmplBytes, err = asset(filename)
		if err != nil {
			log.Noticef("parseFile - Error from Asset() - %v", err)
			return err
		}

		var name = filepath.Base(filename)
		// First template becomes return value if not already defined,
		// and we use that one for subsequent New calls to associate
		// all the templates together. Also, if this file has the same name
		// as t, this file becomes the contents of t, so
		//  t, err := New(name).Funcs(xxx).ParseFiles(name)
		// works. Otherwise we create a new template associated with t.
		var tmpl *htmpl.Template
		if t == nil {
			t = htmpl.New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		_, err = tmpl.Parse(string(tmplBytes))
		if err != nil {
			return err
		}
	}
	return nil
}

func urlTo(routeName string, ps ...interface{}) *url.URL {
	route := appRouter.Get(routeName)
	if route == nil {
		log.Warningf("no such route: %q (params: %v)", routeName, ps)
		return &url.URL{}
	}

	var params []string
	for _, p := range ps {
		switch v := p.(type) {
		case string:
			params = append(params, v)
		case int:
			params = append(params, strconv.Itoa(v))
		case int64:
			params = append(params, strconv.FormatInt(v, 10))

		default:
			log.Errorf("invalid param type %v in route %q", p, routeName)
			logging.CheckFatal(errors.New("invalid param"))
		}
	}

	u, err := route.URLPath(params...)
	if err != nil {
		log.Errorf("Route error: failed to make URL for route %q (params: %v): %s", routeName, params, err)
		return &url.URL{}
	}
	return u
}
