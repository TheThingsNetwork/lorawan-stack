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

package assertions

import (
	"context"
	"fmt"
	"reflect"

	"go.thethings.network/lorawan-stack/pkg/errorcontext"
)

const (
	needContext             = "This assertion requires context.Context as comparison type (you provided %T)."
	shouldHaveParentContext = "Expected context to have parent '%v' (but it didn't)!"
)

var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()

// contextParent returns the parent context of ctx and true if one is found, nil and false otherwise.
// contextParent assumes that ctx has a parent iff it's located at field named Context.
func contextParent(ctx context.Context) (context.Context, bool) {
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
		return contextParent(ctx.Context)
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

// contextHasParent reports whether parent is one of ctx's parents.
// contextHasParent assumes that ctx has a parent iff it's located at field named Context.
func contextHasParent(ctx, parent context.Context) bool {
	for {
		p, ok := contextParent(ctx)
		if !ok {
			return false
		}
		if p == parent {
			return true
		}
		ctx = p
	}
}

// contextRoot returns the root context of ctx.
// contextRoot assumes that ctx has a parent iff it's located at field named Context.
func contextRoot(ctx context.Context) context.Context {
	for {
		p, ok := contextParent(ctx)
		if !ok {
			return ctx
		}
		ctx = p
	}
}

// ShouldHaveParentContext takes as argument a context.Context and context.Context.
// If the arguments are valid and the actual context has the expected context as
// parent, this function returns an empty string.
// Otherwise, it returns a string describing the error.
func ShouldHaveParentContext(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf(needExactValues, 1, len(expected))
	}

	ctx, ok := actual.(context.Context)
	if !ok {
		return fmt.Sprintf(needContext, actual)
	}

	parent, ok := expected[0].(context.Context)
	if !ok {
		return fmt.Sprintf(needContext, expected[0])
	}

	if !contextHasParent(ctx, parent) {
		return fmt.Sprintf(shouldHaveParentContext, parent)
	}
	return success
}

// ShouldHaveParentContextOrEqual takes as argument a context.Context and context.Context.
// If the arguments are valid and the actual context has the expected context as
// parent or if they are equal, this function returns an empty string.
// Otherwise, it returns a string describing the error.
func ShouldHaveParentContextOrEqual(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf(needExactValues, 1, len(expected))
	}

	ctx, ok := actual.(context.Context)
	if !ok {
		return fmt.Sprintf(needContext, actual)
	}

	parent, ok := expected[0].(context.Context)
	if !ok {
		return fmt.Sprintf(needContext, expected[0])
	}

	if ctx != parent && !contextHasParent(ctx, parent) {
		return fmt.Sprintf(shouldHaveParentContext, parent)
	}
	return success
}
