// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

import "context"

type loggerKeyType struct{}

var loggerKey = &loggerKeyType{}

// NewContext returns a derived context with the logger set.
func NewContext(ctx context.Context, logger Interface) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// NewContextWithField returns a derived context with the given field added to the logger.
func NewContextWithField(ctx context.Context, k string, v interface{}) context.Context {
	return NewContext(ctx, FromContext(ctx).WithField(k, v))
}

// NewContextWithFields returns a derived context with the given fields added to the logger.
func NewContextWithFields(ctx context.Context, f Fielder) context.Context {
	return NewContext(ctx, FromContext(ctx).WithFields(f))
}

// FromContext returns the logger that is attached to the context or returns the Noop logger if it does not exist
func FromContext(ctx context.Context) Interface {
	if v := ctx.Value(loggerKey); v != nil {
		if logger, ok := v.(Interface); ok {
			return logger
		}
	}

	return Noop
}
