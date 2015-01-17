package auth

import (
	"errors"
	"time"

	"github.com/gorilla/sessions"
)

func SetStore(s sessions.Store) func(h *Handler) error {
	return func(h *Handler) error {
		if s == nil {
			return errors.New("sessions.Store can't be nil")
		}
		h.store = s
		return nil
	}
}

func SetLanding(l string) func(h *Handler) error {
	return func(h *Handler) error {
		if l == "" {
			return errors.New("landing redirect can't be empty")
		}
		h.landing = l
		return nil
	}
}

func SetLifetime(d time.Duration) func(h *Handler) error {
	return func(h *Handler) error {
		h.lifetime = d
		return nil
	}
}
