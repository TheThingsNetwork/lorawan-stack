// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	v := ctx.Value(loggerKey)
	if v == nil {
		return Noop
	}

	logger, ok := v.(Interface)
	if !ok {
		return Noop
	}
	return logger
}
