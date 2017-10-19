// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

import "context"

// Key is the key where the logger will live in the context.
var Key = &key{}

type key struct{}

// WithLogger sets the logger in the context.
func WithLogger(ctx context.Context, logger Interface) context.Context {
	return context.WithValue(ctx, Key, logger)
}

// FromContext returns the logger that is attached to the context or returns the Noop logger if it does not exist
func FromContext(ctx context.Context) Interface {
	if v := ctx.Value(Key); v != nil {
		if logger, ok := v.(Interface); ok {
			return logger
		}
	}

	return Noop
}
