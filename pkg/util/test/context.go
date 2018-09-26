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

package test

import (
	"context"
	"testing"
)

type tKey struct{}

// ContextWithT saves the test state in the context.
func ContextWithT(ctx context.Context, t *testing.T) context.Context {
	return context.WithValue(ctx, tKey{}, t)
}

// TFromContext returns the test state from the context.
func TFromContext(ctx context.Context) (*testing.T, bool) {
	t, ok := ctx.Value(tKey{}).(*testing.T)
	return t, ok
}

// MustTFromContext returns the test state from the context, and panics if it was not saved in the context.
func MustTFromContext(ctx context.Context) *testing.T {
	t, ok := TFromContext(ctx)
	if !ok {
		panic("*testing.T not present in the context")
	}
	return t
}
