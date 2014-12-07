package auth

import (
	"errors"
	"net/http"
	"testing"

	"github.com/cryptix/go/http/tester"
)

var (
	testMux          *http.ServeMux
	testClient       *tester.Tester
	testAuthProvider mockProvider
)

func setup(t *testing.T) {
	testMux = http.NewServeMux()
	testClient = tester.New(testMux, t)

	ah := NewHandler(&testAuthProvider)
	testMux.Handle("/login", ah.Authorize("/todoRedir"))
	testMux.Handle("/profile", ah.Authenticate(http.HandlerFunc(restricted)))

}

func restricted(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	return
}

func teardown() {
	testMux = nil
}

type mockProvider struct {
	check_ func(string, string) (interface{}, error)
}

func (m mockProvider) Check(user, pass string) (interface{}, error) {
	if m.check_ != nil {
		return m.check_(user, pass)
	}
	return nil, errors.New("Not mocked!")
}
