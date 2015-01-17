package auth

import (
	"errors"
	"net/http"
	"testing"

	"github.com/cryptix/go/http/tester"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var (
	testMux          *http.ServeMux
	testClient       *tester.Tester
	testAuthProvider mockProvider
	testStore        sessions.Store
)

func setup(t *testing.T) {
	testMux = http.NewServeMux()
	testClient = tester.New(testMux, t)
	testStore = &sessions.CookieStore{
		Codecs: securecookie.CodecsFromPairs(
			securecookie.GenerateRandomKey(32), // new key every time we startup
			securecookie.GenerateRandomKey(32),
		),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 30,
		},
	}
	ah, err := NewHandler(&testAuthProvider,
		SetStore(testStore),
		SetLanding("/landingRedir"))
	if err != nil {
		t.Fatal(err)
	}
	testMux.HandleFunc("/login", ah.Authorize)
	testMux.HandleFunc("/logout", ah.Logout)
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
