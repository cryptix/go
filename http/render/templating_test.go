package render

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestRender(t *testing.T) {
	r := mux.NewRouter()
	Init(http.Dir("tests"), []string{"/base.tmpl", "/extra.tmpl"})
	AddTemplates([]string{
		"/test1.tmpl",
		"/error.tmpl",
	})
	SetAppRouter(r)
	Load()

	rw := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := Render(rw, req, "/test1.tmpl", http.StatusOK, nil); err != nil {
		t.Fatal(err)
	}
	// TODO(cryptix): parse html with goquery
}
