package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

// Option is a function that changes a handler in a certain way during initialization
type Option func(h *Handler) error

// SetStore sets the gorilla session.Store impl that should be used
func SetStore(s sessions.Store) Option {
	return func(h *Handler) error {
		if s == nil {
			return errors.New("sessions.Store can't be nil")
		}
		h.store = s
		return nil
	}
}

// SetSessionName sets the name to use for sessions (ie. cookie name)
func SetSessionName(name string) Option {
	return func(h *Handler) error {
		if len(name) == 0 {
			return errors.New("SessionName can't be empty")
		}
		h.sessionName = name
		return nil
	}
}

// SetLanding sets the url to where a client is redirect to after login
func SetLanding(l string) Option {
	return func(h *Handler) error {
		if l == "" {
			return errors.New("landing redirect can't be empty")
		}
		h.redirLanding = l
		return nil
	}
}

// SetLogout sets the url to where a client is redirect to after logout
// uses Landing location by default
func SetLogout(l string) Option {
	return func(h *Handler) error {
		if l == "" {
			return errors.New("logout redirect can't be empty")
		}
		h.redirLogout = l
		return nil
	}
}

// SetLifetime sets the duration of when a session expires after it is created
func SetLifetime(d time.Duration) Option {
	return func(h *Handler) error {
		h.lifetime = d
		return nil
	}
}

// SetNotAuthorizedHandler re-routes the _not authorized_ response to a different http handler
func SetNotAuthorizedHandler(nah http.Handler) Option {
	return func(h *Handler) error {
		h.notAuthorizedHandler = nah
		return nil
	}
}

// ErrorHandler is used to for the SetErrorHandler option.
// It is a classical http.HandleFunc plus the error and response code the auth system determained.
type ErrorHandler func(rw http.ResponseWriter, req *http.Request, err error, code int)

// SetErrorHandler can be used to customize the look up error responses
func SetErrorHandler(errh ErrorHandler) Option {
	return func(h *Handler) error {
		h.errorHandler = errh
		return nil
	}
}
