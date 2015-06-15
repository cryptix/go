package http

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"gopkg.in/errgo.v1"
)

// An ErrorResponse reports errors caused by an API request.
type ErrorResponse struct {
	Response *http.Response `json:",omitempty"`
	Body     []byte
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d\n%v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, string(r.Body))
}

// CheckResponse checks the API response for errors, and returns them if
// present. A response is considered an error if it has a status code outside
// the 200 range. API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse. Any other
// response body will be silently ignored.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	var err error
	errorResponse.Body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return errgo.Notef(err, "cryptix/http: ReadAll(resp.Body) failed. URL: %s", r.Request.URL.String())
	}

	return errorResponse
}
