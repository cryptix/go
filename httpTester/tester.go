package httpTester

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

type Tester struct {
	mux *http.ServeMux
	t   *testing.T
}

func New(mux *http.ServeMux, t *testing.T) *Tester {
	return &Tester{
		mux: mux,
		t:   t,
	}
}

func (t *Tester) GetHTML(uri *url.URL) (*goquery.Document, *httptest.ResponseRecorder) {
	req, err := http.NewRequest("GET", uri.String(), nil)
	if err != nil {
		t.t.Fatal(err)
	}

	rw := httptest.NewRecorder()
	t.mux.ServeHTTP(rw, req)

	doc, err := goquery.NewDocumentFromReader(rw.Body)
	if err != nil {
		t.t.Fatal(err)
	}
	return doc, rw
}

func (t *Tester) GetBody(uri *url.URL) (rw *httptest.ResponseRecorder) {
	req, err := http.NewRequest("GET", uri.String(), nil)
	if err != nil {
		t.t.Fatal(err)
	}

	rw = httptest.NewRecorder()
	t.mux.ServeHTTP(rw, req)
	return
}

func (t *Tester) GetJSON(uri *url.URL, v interface{}) (rw *httptest.ResponseRecorder) {
	req, err := http.NewRequest("GET", uri.String(), nil)
	if err != nil {
		t.t.Fatal(err)
	}

	rw = httptest.NewRecorder()
	t.mux.ServeHTTP(rw, req)
	body := rw.Body.Bytes()
	if rw.Code == 200 {
		if err = json.Unmarshal(body, v); err != nil {
			t.t.Log("Body:", string(body))
			t.t.Fatal(err)
		}
	}

	return
}

func (t *Tester) SendJSON(uri *url.URL, v interface{}) (rw *httptest.ResponseRecorder) {
	blob, err := json.Marshal(v)
	if err != nil {
		t.t.Fatal(err)
	}

	req, err := http.NewRequest("POST", uri.String(), bytes.NewReader(blob))
	if err != nil {
		t.t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	rw = httptest.NewRecorder()
	t.mux.ServeHTTP(rw, req)
	return
}
