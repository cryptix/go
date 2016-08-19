package render

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestRender(t *testing.T) {
	r, err := New(http.Dir("tests"),
		AddTemplates("test1.tmpl"),
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
	r, err := New(http.Dir("tests"),
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
	r, err := New(http.Dir("tests"),
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
