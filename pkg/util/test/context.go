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
	"reflect"
	"testing"

	"go.thethings.network/lorawan-stack/pkg/errorcontext"
)

var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()

// ContextParent returns the parent context of ctx and true if one is found, nil and false otherwise.
// ContextParent assumes that ctx has a parent iff it's located at field named Context.
func ContextParent(ctx context.Context) (context.Context, bool) {
	rv := reflect.ValueOf(ctx)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return nil, false
	}

	switch ctx := rv.Interface().(type) {
	case errorcontext.ErrorContext:
		// ErrorContext wraps the context twice
		return ContextParent(ctx.Context)
	}

	rt := rv.Type()
	if rt.Kind() != reflect.Struct {
		return nil, false
	}

	f, ok := rt.FieldByName("Context")
	if !ok {
		return nil, false
	}
	if !f.Type.Implements(contextType) {
		return nil, false
	}

	fv := rv.FieldByName("Context")
	if (fv.Kind() == reflect.Ptr || fv.Kind() == reflect.Interface) && fv.IsNil() {
		return nil, true
	}
	return fv.Interface().(context.Context), true
}

// ContextHasParent reports whether parent is one of ctx's parents.
// ContextHasParent assumes that ctx has a parent iff it's located at field named Context.
func ContextHasParent(ctx context.Context, parent context.Context) bool {
	for {
		p, ok := ContextParent(ctx)
		if !ok {
			return false
		}
		if p == parent {
			return true
		}
		ctx = p
	}
}

// ContextRoot returns the root context of ctx.
// ContextRoot assumes that ctx has a parent iff it's located at field named Context.
func ContextRoot(ctx context.Context) context.Context {
	for {
		p, ok := ContextParent(ctx)
		if !ok {
			return ctx
		}
		ctx = p
	}
}

type tKey struct{}

func ContextWithT(ctx context.Context, t *testing.T) context.Context {
	return context.WithValue(ctx, tKey{}, t)
}

func TFromContext(ctx context.Context) (*testing.T, bool) {
	t, ok := ctx.Value(tKey{}).(*testing.T)
	return t, ok
}

func MustTFromContext(ctx context.Context) *testing.T {
	t, ok := TFromContext(ctx)
	if !ok {
		panic("*testing.T not present in the context")
	}
	return t
}
