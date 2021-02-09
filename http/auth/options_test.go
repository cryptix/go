package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
	testAuthProvider.checkMock = func(u, p string) (interface{}, error) {
		called = true
		if !(u == "testUser" && p == "testPassw") {
			return nil, ErrBadLogin
		}
		return 23, nil
	}
	resp := testClient.PostForm("/login", vals)
	a.Equal(http.StatusSeeOther, resp.Code)
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
	testAuthProvider.checkMock = func(u, p string) (interface{}, error) {
		called = true
		if !(u == "testUser" && p == "testPassw") {
			return nil, ErrBadLogin
		}
		return 23, nil
	}
	resp := testClient.PostForm("/login", vals)
	a.Equal(http.StatusSeeOther, resp.Code)
	a.Equal("/landingRedir", resp.Header().Get("Location"))
	a.True(called)
	newCookie := resp.Header().Get("Set-Cookie")
	a.Contains(newCookie, defaultSessionName)

	resp = testClient.PostForm("/logout", nil)
	a.Equal(http.StatusSeeOther, resp.Code)
	a.Equal(want, resp.Header().Get("Location"))
}

func TestOption_errhandler(t *testing.T) {

	var errh = func(rw http.ResponseWriter, req *http.Request, err error, code int) {
		rw.WriteHeader(code)
		fmt.Fprintln(rw, "custom error:", code)
		fmt.Fprintln(rw, err.Error())
	}
	testOptions = []Option{
		SetErrorHandler(errh),
	}
	setup(t)
	defer teardown()
	a := assert.New(t)

	testAuthProvider.checkMock = func(u, p string) (interface{}, error) {
		return nil, fmt.Errorf("simulated database outage")
	}

	vals := url.Values{"user": {"testUser"}, "pass": {"testPassw"}}
	resp := testClient.PostForm("/login", vals)

	a.Equal(http.StatusInternalServerError, resp.Code)
	body := resp.Body.String()
	a.True(strings.HasPrefix(body, "custom error: 500"))
}
