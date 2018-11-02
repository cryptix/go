package logging

import (
	stdlog "log"

	kitlog "github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

type Option func(*InfoLog) error

func With(keyvals ...interface{}) Option {
	return func(il *InfoLog) error {
		if len(keyvals)%2 != 0 {
			return errors.Errorf("logging/With: odd number of keyvals")
		}
		//il.Interface = kitlog.With(il.Interface, "time", kitlog.DefaultTimestamp, "caller", kitlog.DefaultCaller)
		il.Interface = kitlog.With(il.Interface, keyvals...)
		return nil
	}
}

func ToStdlib() Option {
	return func(il *InfoLog) error {
		stdlog.SetOutput(kitlog.NewStdlibAdapter(kitlog.With(il.Interface, "module", "stdlib")))
		return nil
	}
}
