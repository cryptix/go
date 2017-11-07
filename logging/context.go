package logging

import (
	"context"
)

type logctxKeyT string

var LogCTXKey logctxKeyT = "loggingContextKey"

func NewContext(ctx context.Context, log Interface) context.Context {
	return context.WithValue(ctx, LogCTXKey, log)
}

func FromContext(ctx context.Context) Interface {
	v, ok := ctx.Value(LogCTXKey).(Interface)
	if !ok {
		internal.Log("warning", "no logger inside context")
		return internal
	}
	return v
}
