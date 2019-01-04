// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package test

import (
	"context"
	"sync/atomic"
	"testing"
)

type tKey struct{}

// ContextWithT saves the testing.T in the context.
func ContextWithT(ctx context.Context, t *testing.T) context.Context {
	return context.WithValue(ctx, tKey{}, t)
}

// TFromContext returns the testing.T saved using ContextWithT from the context.
func TFromContext(ctx context.Context) (*testing.T, bool) {
	t, ok := ctx.Value(tKey{}).(*testing.T)
	if !ok {
		return nil, false
	}
	return t, true
}

// MustTFromContext returns the test state from the context, and panics if it was not saved in the context.
func MustTFromContext(ctx context.Context) *testing.T {
	t, ok := TFromContext(ctx)
	if !ok {
		panic("*testing.T not present in the context")
	}
	return t
}

// ContextWithCounter adds a counter to ctx under key specified.
func ContextWithCounter(ctx context.Context, key interface{}) context.Context {
	var i int64
	return context.WithValue(ctx, key, &i)
}

// CounterFromContext gets the counter from context.
func CounterFromContext(ctx context.Context, key interface{}) (int64, bool) {
	i, ok := ctx.Value(key).(*int64)
	if !ok {
		return 0, false
	}
	return *i, true
}

// MustCounterFromContext gets the counter from context, and panics if it is not present in the context
func MustCounterFromContext(ctx context.Context, key interface{}) int64 {
	i, ok := CounterFromContext(ctx, key)
	if !ok {
		panic("counter not present in the context")
	}
	return i
}

// IncrementContextCounter increments the counter in the context.
func IncrementContextCounter(ctx context.Context, key interface{}, v int64) (int64, bool) {
	i, ok := ctx.Value(key).(*int64)
	if !ok {
		return 0, false
	}
	return atomic.AddInt64(i, v), true
}

// MustIncrementContextCounter increments the counter in the context, and panics if it is not present in the context.
func MustIncrementContextCounter(ctx context.Context, key interface{}, v int64) int64 {
	i, ok := IncrementContextCounter(ctx, key, v)
	if !ok {
		panic("counter not present in the context")
	}
	return i
}
