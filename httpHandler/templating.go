// Package templates implements template inheritance and exposes functions to render these
//
// inspired by http://elithrar.github.io/article/approximating-html-template-inheritance/
package httpHandler

import (
	"errors"
	"fmt"
	htmpl "html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/cryptix/go/utils"
	"github.com/gorilla/mux"
	"github.com/oxtoacart/bpool"
)

var (
	// Reload is whether to reload templates on each request.
	Reload bool

	// files
	templateFiles     [][]string
	baseTemplateFiles []string

	basePath  string
	appRouter *mux.Router

	// all the templates that we parsed
	templates = map[string]*htmpl.Template{}

	// bufpool is shared between all render() calls
	bufpool = bpool.NewBufferPool(64)

	// go.rice box for loading templates
	tmplBox *rice.Box
)

func SetRiceBox(b *rice.Box) {
	tmplBox = b
}

func SetBasePath(p ...string) {
	if len(p) == 1 {
		basePath = p[0]
	} else {
		basePath = filepath.Join(p...)
	}
}

func SetBaseTemplates(files []string) {
	baseTemplateFiles = append(baseTemplateFiles, files...)
}

func AddTemplates(files [][]string) {
	templateFiles = append(templateFiles, files...)
}

func SetAppRouter(r *mux.Router) {
	appRouter = r
}

// Load loads and parses all templates that are in templateDir
func Load() {
	if appRouter == nil {
		utils.CheckFatal(errors.New("No appRouter set"))
	}

	if len(baseTemplateFiles) == 0 {
		baseTemplateFiles = []string{"navbar.tmpl", "base.tmpl"}
	}

	if tmplBox == nil {
		utils.CheckFatal(errors.New("No riceBox set"))
	}
	utils.CheckFatal(parseHTMLTemplates(templateFiles))
}

func parseHTMLTemplates(sets [][]string) error {
	for _, set := range sets {
		t := htmpl.New("")
		t.Funcs(htmpl.FuncMap{
			"urlTo": urlTo,
			"itoa":  strconv.Itoa,
		})

		err := parseFilesFromBindata(t, set...)
		if err != nil {
			return fmt.Errorf("template %v: %s", set, err)
		}

		t = t.Lookup("base")
		if t == nil {
			return fmt.Errorf("base template not found in %v", set)
		}
		templates[set[0]] = t
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
	utils.LogInfo.Printf("Rendered '%s' Status:%d (took %v)", name, status, time.Since(start))
	return err
}

// PlainError helps rendering user errors
func PlainError(w http.ResponseWriter, statusCode int, err error) {
	utils.LogErr.Printf("PlainError(%d):%s\n", statusCode, err)
	http.Error(w, err.Error(), statusCode)
}

// copied from template.ParseFiles but dont use ioutil.ReadFile
func parseFilesFromBindata(t *htmpl.Template, filenames ...string) error {
	var err error

	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return errors.New("templates: no files named in call to parseFilesFromBindata")
	}
	files := append(filenames, baseTemplateFiles...)
	utils.LogInfo.Printf("parseFile - %q", files)
	for _, filename := range files {
		var s string
		s, err = tmplBox.String(filename)
		if err != nil {
			utils.LogInfo.Printf("parseFile - Error from tmplBox.String() - %v", err)
			return err
		}

		name := filepath.Base(filename)
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
		_, err = tmpl.Parse(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func urlTo(routeName string, ps ...interface{}) *url.URL {
	route := appRouter.Get(routeName)
	if route == nil {
		utils.LogErr.Printf("no such route: %q (params: %v)", routeName, ps)
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
			utils.LogErr.Printf("invalid param type %v in route %q", p, routeName)
			utils.CheckFatal(errors.New("invalid param"))
		}
	}

	u, err := route.URLPath(params...)
	if err != nil {
		utils.LogErr.Printf("Route error: failed to make URL for route %q (params: %v): %s", routeName, params, err)
		return &url.URL{}
	}
	return u
}
