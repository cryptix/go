package auth

import (
	"errors"
	"net/http"
	"testing"

	"go.mindeco.de/toolbelt/http/tester"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var (
	testMux          *http.ServeMux
	testClient       *tester.Tester
	testAuthProvider mockProvider
	testStore        sessions.Store
	testOptions      []Option
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

	var opts = append([]Option{SetStore(testStore), SetLanding("/landingRedir")}, testOptions...)
	ah, err := NewHandler(&testAuthProvider, opts...)
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
	checkMock func(string, string) (interface{}, error)
}

func (m mockProvider) Check(user, pass string) (interface{}, error) {
	if m.checkMock != nil {
		return m.checkMock(user, pass)
	}
	return nil, errors.New("Not mocked!")
}
