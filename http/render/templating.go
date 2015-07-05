package render

import (
	"errors"
	"fmt"
	htmpl "html/template"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cryptix/go/logging"
	"github.com/gorilla/mux"
	"github.com/oxtoacart/bpool"
	"github.com/shurcooL/go/vfs/httpfs/html/vfstemplate"
	"gopkg.in/errgo.v1"
)

var (
	// Reload is whether to reload templates on each request.
	Reload bool

	log = logging.Logger("render")

	assets http.FileSystem

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
func Init(fs http.FileSystem, base []string) {
	assets = fs
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
		logging.CheckFatal(errgo.New("No appRouter set"))
	}

	if len(baseTemplateFiles) == 0 {
		logging.CheckFatal(errgo.New("No base tempaltes"))
		// baseTemplateFiles = []string{"navbar.tmpl", "base.tmpl"}
	}

	logging.CheckFatal(parseHTMLTemplates())
}

func parseHTMLTemplates() error {
	for _, file := range templateFiles {
		t := htmpl.New("").Funcs(htmpl.FuncMap{
			"urlTo": urlTo,
			"itoa":  strconv.Itoa,
		})
		var err error
		t, err = vfstemplate.ParseFiles(assets, t, append(baseTemplateFiles, file)...)
		if err != nil {
			return errgo.Notef(err, "template %s", file)
		}

		t = t.Lookup("base")
		if t == nil {
			return errgo.Newf("base template not found in %v", file)
		}
		// TODO(cryptix): refactor all of this.. maybe templateName > path?
		templates[strings.TrimPrefix(file, "/tmpl")] = t
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
	log.WithFields(logrus.Fields{
		"name":   name,
		"status": status,
		"took":   time.Since(start),
	}).Debug("Rendered")
	return err
}

// PlainError helps rendering user errors
func PlainError(w http.ResponseWriter, statusCode int, err error) {
	log.WithFields(logrus.Fields{
		"status": statusCode,
		"err":    errgo.Details(err),
	}).Error("PlainError")
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

func urlTo(routeName string, ps ...interface{}) *url.URL {
	route := appRouter.Get(routeName)
	if route == nil {
		log.WithFields(logrus.Fields{
			"route":  routeName,
			"params": ps,
		}).Warning("no such route")
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
		log.WithFields(logrus.Fields{
			"route":  routeName,
			"params": params,
			"error":  err,
		}).Warning("no such route")
		log.Error("Route error: failed to make URL")
		return &url.URL{}
	}
	return u
}
