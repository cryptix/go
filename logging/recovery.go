package logging

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/pkg/errors"
)

func RecoveryHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			var err error
			defer func() {
				if r := recover(); r != nil {
					l := FromContext(req.Context())
					switch t := r.(type) {
					case string:
						err = errors.New(t)
					case error:
						err = t
					default:
						err = errors.Errorf("unkown error: %v", r)
					}
					os.Mkdir("panics", os.ModePerm)
					b, tmpErr := ioutil.TempFile("panics", "httpRecovery")
					if tmpErr != nil {
						panic(errors.Wrap(tmpErr, "failed to create httpRecovery log"))
					}
					fmt.Fprintf(b, "warning! httpRecovery!\nError: %s\n", err)
					fmt.Fprintf(b, "Stack:\n%s", debug.Stack())

					l.Log("event", "httpPanic", "panicLog", b.Name())

					if err := b.Close(); err != nil {
						panic(errors.Wrap(err, "failed to close httpRecovery log"))
					}
					http.Error(w, "internal processing error - please try again", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, req)
		})
	}
}
