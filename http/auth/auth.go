package auth

import (
	"encoding/gob"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

// custom sessionKey type to prevent collision
type sessionKey uint

func init() {
	// need to register our Key with gob so gorilla/sessions can (de)serialize it
	gob.Register(userKey)
	gob.Register(time.Time{})
}

const (
	sessionName = "AuthSession"

	userKey sessionKey = iota
	userTimeout
)

var (
	ErrBadLogin      = errors.New("Bad Login")
	ErrNotAuthorized = errors.New("Not Authorized")
)

// Auther allows for custom authentication backends
type Auther interface {
	// Check should return a non-nil error for failed requests (like ErrBadLogin)
	// and it can pass custom data that is saved in the cookie through the first return argument
	Check(user, pass string) (interface{}, error)
}

type AuthHandler struct {
	auther Auther
	store  sessions.Store
}

func NewHandler(a Auther, store sessions.Store) (ah AuthHandler) {
	ah.auther = a
	ah.store = store
	return
}

func (ah AuthHandler) Authorize(redir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := ah.store.Get(r, sessionName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := r.Form.Get("user")
		pass := r.Form.Get("pass")
		if user == "" || pass == "" {
			http.Error(w, ErrBadLogin.Error(), http.StatusBadRequest)
			return
		}

		id, err := ah.auther.Check(user, pass)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		session.Values[userKey] = id
		session.Values[userTimeout] = time.Now().Add(5 * time.Minute)
		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, redir, http.StatusTemporaryRedirect)
		return
	}
}

func (ah AuthHandler) Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := ah.AuthenticateRequest(r); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func (ah AuthHandler) AuthenticateRequest(r *http.Request) error {
	session, err := ah.store.Get(r, sessionName)
	if err != nil {
		return err
	}

	if session.IsNew {
		return ErrNotAuthorized
	}

	if _, ok := session.Values[userKey]; !ok {
		return ErrNotAuthorized
	}

	t, ok := session.Values[userTimeout]
	if !ok {
		return ErrNotAuthorized
	}

	tout, ok := t.(time.Time)
	if !ok {
		return ErrNotAuthorized
	}

	if time.Now().After(tout) {
		return ErrNotAuthorized
	}

	return nil
}

func (ah AuthHandler) Logout(redir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := ah.store.Get(r, sessionName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session.Values[userTimeout] = time.Now().Add(-5 * time.Minute)
		session.Options.MaxAge = -1
		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, redir, http.StatusTemporaryRedirect)
		return
	}
}
