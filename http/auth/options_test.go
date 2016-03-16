package auth

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOption_landing(t *testing.T) {
	const want = "/someLocation"
	testOptions = []Option{
		SetLanding(want),
	}
	setup(t)
	defer teardown()
	a := assert.New(t)

	vals := url.Values{
		"user": {"testUser"},
		"pass": {"testPassw"},
	}
	called := false
	testAuthProvider.check_ = func(u, p string) (interface{}, error) {
		called = true
		if !(u == "testUser" && p == "testPassw") {
			return nil, ErrBadLogin
		}
		return 23, nil
	}
	resp := testClient.PostForm("/login", vals)
	a.Equal(http.StatusFound, resp.Code)
	a.Equal(want, resp.Header().Get("Location"))
	a.True(called)
	newCookie := resp.Header().Get("Set-Cookie")
	a.Contains(newCookie, defaultSessionName)
}

func TestOption_logout(t *testing.T) {
	const want = "/someLocation"
	testOptions = []Option{
		SetLogout(want),
	}
	setup(t)
	defer teardown()
	a := assert.New(t)

	vals := url.Values{
		"user": {"testUser"},
		"pass": {"testPassw"},
	}
	called := false
	testAuthProvider.check_ = func(u, p string) (interface{}, error) {
		called = true
		if !(u == "testUser" && p == "testPassw") {
			return nil, ErrBadLogin
		}
		return 23, nil
	}
	resp := testClient.PostForm("/login", vals)
	a.Equal(http.StatusFound, resp.Code)
	a.Equal("/landingRedir", resp.Header().Get("Location"))
	a.True(called)
	newCookie := resp.Header().Get("Set-Cookie")
	a.Contains(newCookie, defaultSessionName)

	resp = testClient.PostForm("/logout", nil)
	a.Equal(http.StatusFound, resp.Code)
	a.Equal(want, resp.Header().Get("Location"))
}
