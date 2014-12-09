package auth

import (
	"errors"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/mholt/binding"
)

const (
	SessionName = "AP_Session"
	UserKey     = "AP_User"
)

var ErrBadLogin = errors.New("Bad Login")

type Auther interface {
	Check(string, string) (interface{}, error)
}

type loginRequest struct {
	User, Pass string
}

func (lr *loginRequest) FieldMap() binding.FieldMap {
	return binding.FieldMap{
		&lr.User: "user",
		&lr.Pass: "pass",
	}
}

type AuthHandler struct {
	auther Auther
	store  sessions.Store
}

func NewHandler(a Auther) (ah AuthHandler) {
	ah.auther = a
	ah.store = &sessions.CookieStore{
		Codecs: securecookie.CodecsFromPairs(
			securecookie.GenerateRandomKey(32), // new key every time we startup
			securecookie.GenerateRandomKey(32),
		),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
			// BUG(Henry): Add flag to toggle SSL-Only
			// Secure:   true,
			HttpOnly: true,
		},
	}
	return
}

func (ah AuthHandler) Authorize(redir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := ah.store.Get(r, SessionName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		lr := new(loginRequest)
		errs := binding.Bind(r, lr)
		if errs.Handle(w) {
			return
		}

		if lr.User == "" || lr.Pass == "" {
			http.Error(w, ErrBadLogin.Error(), http.StatusBadRequest)
			return
		}

		id, err := ah.auther.Check(lr.User, lr.Pass)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		session.Values[UserKey] = id
		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, redir, http.StatusFound)
		return
	}
}

func (ah AuthHandler) Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := ah.store.Get(r, SessionName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, ok := session.Values[UserKey]; !ok {
			http.Error(w, "Not Authorized", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}
