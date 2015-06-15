package http

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestCheck_valid(t *testing.T) {
	var valid = &http.Response{
		StatusCode: 200,
	}
	err := CheckResponse(valid)
	if err != nil {
		t.Error("marked valid response as OK")
	}
}

func TestCheck_404(t *testing.T) {
	var buf = bytes.NewBufferString("test not found")
	var valid = &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(buf),
	}
	err := CheckResponse(valid)
	if err == nil {
		t.Error("marked invalid response as OK")
	}
}

func TestCheck_500(t *testing.T) {
	var buf = bytes.NewBufferString("test err")
	var valid = &http.Response{
		StatusCode: 500,
		Body:       ioutil.NopCloser(buf),
	}
	err := CheckResponse(valid)
	if err == nil {
		t.Error("marked invalid response as OK")
	}
}
