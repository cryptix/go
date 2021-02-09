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
	defaultSessionName = "AuthSession"

	userKey sessionKey = iota
	userTimeout
)

// errors to be checked against returned
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

	errorHandler         ErrorHandler
	notAuthorizedHandler http.Handler

	redirLanding string // the url to redirect to after login
	redirLogout  string // the url to redirect to after logout

	// how long should a session life
	lifetime time.Duration

	// the name of the cookie
	sessionName string
}

func NewHandler(a Auther, options ...Option) (*Handler, error) {
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

	if ah.redirLanding == "" {
		ah.redirLanding = "/"
	}

	if ah.redirLogout == "" {
		ah.redirLogout = ah.redirLanding
	}

	if ah.sessionName == "" {
		ah.sessionName = defaultSessionName
	}

	if ah.errorHandler == nil {
		ah.errorHandler = func(w http.ResponseWriter, r *http.Request, err error, code int) {
			http.Error(w, err.Error(), code)
		}
	}

	if ah.notAuthorizedHandler == nil {
		ah.notAuthorizedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ah.errorHandler(w, r, ErrNotAuthorized, http.StatusUnauthorized)
		})
	}

	return &ah, nil
}

func (ah Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	session, err := ah.store.Get(r, ah.sessionName)
	if err != nil {
		ah.errorHandler(w, r, err, http.StatusInternalServerError)
		return
	}
	if err := r.ParseForm(); err != nil {
		ah.errorHandler(w, r, err, http.StatusInternalServerError)
		return
	}

	user := r.Form.Get("user")
	pass := r.Form.Get("pass")
	if user == "" || pass == "" {
		ah.errorHandler(w, r, ErrBadLogin, http.StatusBadRequest)
		return
	}

	id, err := ah.auther.Check(user, pass)
	if err != nil {
		var code = http.StatusInternalServerError
		if err == ErrBadLogin {
			code = http.StatusBadRequest
		}
		ah.errorHandler(w, r, err, code)
		return
	}

	session.Values[userKey] = id
	session.Values[userTimeout] = time.Now().Add(ah.lifetime)
	if err := session.Save(r, w); err != nil {
		ah.errorHandler(w, r, err, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, ah.redirLanding, http.StatusSeeOther)
	return
}

func (ah Handler) Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := ah.AuthenticateRequest(r); err != nil {
			ah.notAuthorizedHandler.ServeHTTP(w, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func (ah Handler) AuthenticateRequest(r *http.Request) (interface{}, error) {
	session, err := ah.store.Get(r, ah.sessionName)
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
	session, err := ah.store.Get(r, ah.sessionName)
	if err != nil {
		ah.errorHandler(w, r, err, http.StatusInternalServerError)
		return
	}

	session.Values[userTimeout] = time.Now().Add(-ah.lifetime)
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		ah.errorHandler(w, r, err, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, ah.redirLogout, http.StatusSeeOther)
	return
}
