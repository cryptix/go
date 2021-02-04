package render

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"go.mindeco.de/logging"
	"go.mindeco.de/logging/logtest"
)

func TestRender(t *testing.T) {
	logging.SetupLogging(logtest.Logger("Rendere", t))
	log := logging.Logger("TestRender")
	r, err := New(http.Dir("tests"),
		AddTemplates("test1.tmpl"),
		SetLogger(log),
	)
	if err != nil {
		t.Fatal("New() failed", err)
	}
	rw := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Render(rw, req, "test1.tmpl", http.StatusOK, nil); err != nil {
		t.Fatal(err)
	}
	if rw.Code != http.StatusOK {
		t.Fatal("wrong status")
	}
	doc, err := goquery.NewDocumentFromReader(rw.Body)
	if err != nil {
		t.Fatal(err)
	}
	if title := doc.Find("title").Text(); title != "render - tests" {
		t.Fatalf("wrong Title. got: %s", title)
	}
	if hello := doc.Find("#hello").Text(); hello != "Hello" {
		t.Fatalf("wrong hello. got: %s", hello)
	}
	if testID := doc.Find("#testID").Text(); testID != "Test2" {
		t.Fatalf("wrong testID. got: %s", testID)
	}
}

func TestFuncMap(t *testing.T) {
	logging.SetupLogging(logtest.Logger("Rendere", t))
	log := logging.Logger("TestFuncMap")
	r, err := New(http.Dir("tests"),
		SetLogger(log),
		AddTemplates("testFuncMap.tmpl"),
		FuncMap(template.FuncMap{
			"itoa": strconv.Itoa,
		}),
	)
	if err != nil {
		t.Fatal("New() failed", err)
	}
	rw := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Render(rw, req, "testFuncMap.tmpl", http.StatusOK, nil); err != nil {
		t.Fatal(err)
	}
	if rw.Code != http.StatusOK {
		t.Fatal("wrong status")
	}
}

func TestBugOverride(t *testing.T) {
	logging.SetupLogging(logtest.Logger("Rendere", t))
	log := logging.Logger("TestBugOverride")
	r, err := New(http.Dir("tests"),
		SetLogger(log),
		AddTemplates("testFuncMap.tmpl", "bug1.tmpl"),
		FuncMap(template.FuncMap{"itoa": strconv.Itoa}),
	)
	if err != nil {
		t.Fatal("New() failed", err)
	}
	rw := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Render(rw, req, "testFuncMap.tmpl", http.StatusOK, nil); err != nil {
		t.Fatal(err)
	}
	if rw.Code != http.StatusOK {
		t.Fatal("wrong status")
	}
	if !strings.Contains(rw.Body.String(), "42") {
		t.Fatal("first doesn't contain 42")
	}
}

func TestBaseTmpl(t *testing.T) {
	logging.SetupLogging(logtest.Logger("Rendere", t))
	log := logging.Logger("TestBugOverride")
	r, err := New(http.Dir("tests"),
		SetLogger(log),
		BaseTemplates("subdir/base2.tmpl"),
		AddTemplates("test1.tmpl"),
		FuncMap(template.FuncMap{"itoa": strconv.Itoa}),
	)
	if err != nil {
		t.Fatal("New() failed", err)
	}
	rw := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Render(rw, req, "test1.tmpl", http.StatusOK, nil); err != nil {
		t.Fatal(err)
	}
	if rw.Code != http.StatusOK {
		t.Fatal("wrong status")
	}
	doc, err := goquery.NewDocumentFromReader(rw.Body)
	if err != nil {
		t.Fatal(err)
	}
	if heading := doc.Find("#baseHead").Text(); heading != "Alternative base in a subdir" {
		t.Fatalf("wrong heading. got: %s", heading)
	}
}

func TestMultileBaseTmpls(t *testing.T) {
	logging.SetupLogging(logtest.Logger("Rendere", t))
	log := logging.Logger("TestMultileBaseTmpls")
	r, err := New(http.Dir("tests"),
		SetLogger(log),
		BaseTemplates("subdir/base2.tmpl", "extra.tmpl"),
		AddTemplates("test2.tmpl"),
	)
	if err != nil {
		t.Fatal("New() failed", err)
	}
	rw := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Render(rw, req, "test2.tmpl", http.StatusOK, nil); err != nil {
		t.Fatal(err)
	}
	if rw.Code != http.StatusOK {
		t.Fatal("wrong status")
	}
	doc, err := goquery.NewDocumentFromReader(rw.Body)
	if err != nil {
		t.Fatal(err)
	}
	if ex := doc.Find("#extra").Text(); ex != "additional base tpl" {
		t.Fatalf("wrong ex. got: %s", ex)
	}
}

func TestFuncInjection(t *testing.T) {
	logging.SetupLogging(logtest.Logger("Render", t))
	log := logging.Logger("TestFuncMap")
	r, err := New(http.Dir("tests"),
		SetLogger(log),
		AddTemplates("testInject.tmpl"),
		InjectTemplateFunc("addr", func(r *http.Request) interface{} {
			return func() string {
				return r.Header.Get("X-Test-Addr")
			}
		}),
	)
	if err != nil {
		t.Fatal("New() failed", err)
	}
	rw := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Test-Addr", "localhost:1234")

	if err := r.Render(rw, req, "testInject.tmpl", http.StatusOK, nil); err != nil {
		t.Fatal(err)
	}
	if rw.Code != http.StatusOK {
		t.Fatal("wrong status")
	}
	doc, err := goquery.NewDocumentFromReader(rw.Body)
	if err != nil {
		t.Fatal(err)
	}
	if ex := doc.Find("#addr").Text(); ex != "localhost:1234" {
		t.Fatalf("wrong ex. got: %s", ex)
	}
}
