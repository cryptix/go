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

type Handler struct {
	auther Auther
	store  sessions.Store

	// the url to redirect to after login/logout
	landing string

	// how long should a session life
	lifetime time.Duration
}

func NewHandler(a Auther, options ...func(*Handler) error) (*Handler, error) {
	var ah Handler
	ah.auther = a

	for _, o := range options {
		if err := o(&ah); err != nil {
			return nil, err
		}
	}

	if ah.store == nil {
		return nil, errors.New("please set a session.Store")
	}

	// defaults
	if ah.lifetime == 0 {
		ah.lifetime = 5 * time.Minute
	}

	if ah.landing == "" {
		ah.landing = "/"
	}
	return &ah, nil
}

func (ah Handler) Authorize(w http.ResponseWriter, r *http.Request) {
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
	session.Values[userTimeout] = time.Now().Add(ah.lifetime)
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, ah.landing, http.StatusTemporaryRedirect)
	return
}

func (ah Handler) Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := ah.AuthenticateRequest(r); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func (ah Handler) AuthenticateRequest(r *http.Request) (interface{}, error) {
	session, err := ah.store.Get(r, sessionName)
	if err != nil {
		return nil, err
	}

	if session.IsNew {
		return nil, ErrNotAuthorized
	}

	user, ok := session.Values[userKey]
	if !ok {
		return nil, ErrNotAuthorized
	}

	t, ok := session.Values[userTimeout]
	if !ok {
		return nil, ErrNotAuthorized
	}

	tout, ok := t.(time.Time)
	if !ok {
		return nil, ErrNotAuthorized
	}

	if time.Now().After(tout) {
		return nil, ErrNotAuthorized
	}

	return user, nil
}

func (ah Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := ah.store.Get(r, sessionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values[userTimeout] = time.Now().Add(-ah.lifetime)
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, ah.landing, http.StatusTemporaryRedirect)
	return
}
