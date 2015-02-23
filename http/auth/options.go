package auth

import (
	"errors"
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

// SetLanding sets the url to where a client is redirect
func SetLanding(l string) Option {
	return func(h *Handler) error {
		if l == "" {
			return errors.New("landing redirect can't be empty")
		}
		h.landing = l
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
