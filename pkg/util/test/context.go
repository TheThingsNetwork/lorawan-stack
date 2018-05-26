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
	"time"
)

type MockContext struct {
	DeadlineFunc func() (deadline time.Time, ok bool)
	DoneFunc     func() <-chan struct{}
	ErrFunc      func() error
	ValueFunc    func(interface{}) interface{}
}

func (ctx *MockContext) Deadline() (deadline time.Time, ok bool) {
	if ctx.DeadlineFunc == nil {
		return time.Time{}, false
	}
	return ctx.DeadlineFunc()
}

func (ctx *MockContext) Done() <-chan struct{} {
	if ctx.DoneFunc == nil {
		return nil
	}
	return ctx.DoneFunc()
}
func (ctx *MockContext) Err() error {
	if ctx.ErrFunc == nil {
		return nil
	}
	return ctx.ErrFunc()
}
func (ctx *MockContext) Value(key interface{}) interface{} {
	if ctx.ValueFunc == nil {
		return nil
	}
	return ctx.ValueFunc(key)
}

var globalContext = &MockContext{}

func Context() context.Context {
	return globalContext
}

var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()

// ContextParent returns the parent context of ctx and true if one is found, nil and false otherwise.
func ContextParent(ctx context.Context) (context.Context, bool) {
	rv := reflect.ValueOf(ctx)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return nil, false
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

	return rv.FieldByName("Context").Interface().(context.Context), true
}

// ContextRoot returns the root context of ctx.
func ContextRoot(ctx context.Context) context.Context {
	for ok := true; ok; {
		p, ok := ContextParent(ctx)
		if !ok {
			return ctx
		}
		ctx = p
	}
	panic("Unreachable")
}
