// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

import (
	"context"
	"fmt"
)

// Key is the key where the logger will live in the context.
var Key = &key{}

type key struct{}

// WithLogger sets the logger in the context.
func WithLogger(ctx context.Context, logger Interface) context.Context {
	return context.WithValue(ctx, Key, logger)
}

// FromContextMaybe returns the logger that is attached to the context or nil if does not exist.
func FromContextMaybe(ctx context.Context) Interface {
	if v := ctx.Value(Key); v != nil {
		if i, ok := v.(Interface); ok {
			return i
		}
	}

	return nil
}

// FromContext returns the logger that is attached to the context or panics if it does not exist.
func FromContext(ctx context.Context) Interface {
	logger := FromContextMaybe(ctx)
	if logger != nil {
		return logger
	}

	panic(fmt.Sprintf("No logger in context at key %s", Key))
}
