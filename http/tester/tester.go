package tester

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/cryptix/go/logging"
	"github.com/cryptix/go/logging/logtest"
)

type Tester struct {
	mux http.Handler
	t   *testing.T
}

func New(mux *http.ServeMux, t *testing.T) *Tester {
	l, _ := logtest.KitLogger("http/tester", t)
	return &Tester{
		mux: logging.InjectHandler(l)(mux),
		t:   t,
	}
}

func (t *Tester) GetHTML(u string, h *http.Header) (*goquery.Document, *httptest.ResponseRecorder) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		t.t.Fatal(err)
	}

	if h != nil {
		req.Header = *h
	}

	rw := httptest.NewRecorder()
	t.mux.ServeHTTP(rw, req)

	doc, err := goquery.NewDocumentFromReader(rw.Body)
	if err != nil {
		t.t.Fatal(err)
	}
	return doc, rw
}

func (t *Tester) GetBody(u string, h *http.Header) (rw *httptest.ResponseRecorder) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		t.t.Fatal(err)
	}

	if h != nil {
		req.Header = *h
	}

	rw = httptest.NewRecorder()
	t.mux.ServeHTTP(rw, req)
	return
}

func (t *Tester) GetJSON(u string, v interface{}) (rw *httptest.ResponseRecorder) {
	req, err := http.NewRequest("GET", u, nil)
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

func (t *Tester) SendJSON(u string, v interface{}) (rw *httptest.ResponseRecorder) {
	blob, err := json.Marshal(v)
	if err != nil {
		t.t.Fatal(err)
	}

	req, err := http.NewRequest("POST", u, bytes.NewReader(blob))
	if err != nil {
		t.t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	rw = httptest.NewRecorder()
	t.mux.ServeHTTP(rw, req)
	return
}

func (t *Tester) PostForm(u string, v url.Values) (rw *httptest.ResponseRecorder) {
	req, err := http.NewRequest("POST", u, strings.NewReader(v.Encode()))
	if err != nil {
		t.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rw = httptest.NewRecorder()
	t.mux.ServeHTTP(rw, req)
	return
}
