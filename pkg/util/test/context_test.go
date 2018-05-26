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
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestParentContext(t *testing.T) {
	for _, tc := range []struct {
		Name       string
		Parent     context.Context
		NewContext func(context.Context) context.Context
		OK         bool
	}{
		{
			Name:   "context.WithCancel",
			Parent: &MockContext{},
			NewContext: func(ctx context.Context) context.Context {
				ctx, _ = context.WithCancel(ctx)
				return ctx
			},
			OK: true,
		},
		{
			Name:   "context.WithValue",
			Parent: &MockContext{},
			NewContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, struct{}{}, nil)
			},
			OK: true,
		},
		{
			Name:   "context.WithDeadline",
			Parent: &MockContext{},
			NewContext: func(ctx context.Context) context.Context {
				ctx, _ = context.WithDeadline(ctx, time.Now().Add(200*Delay))
				return ctx
			},
			OK: true,
		},
		{
			Name:   "context.WithTimeout",
			Parent: &MockContext{},
			NewContext: func(ctx context.Context) context.Context {
				ctx, _ = context.WithTimeout(ctx, 200*Delay)
				return ctx
			},
			OK: true,
		},
		{
			Name:   "context.Background",
			Parent: nil,
			NewContext: func(ctx context.Context) context.Context {
				return context.Background()
			},
			OK: false,
		},
		{
			Name:   "context.TODO",
			Parent: nil,
			NewContext: func(ctx context.Context) context.Context {
				return context.TODO()
			},
			OK: false,
		},
		{
			Name:   "nil",
			Parent: nil,
			NewContext: func(context.Context) context.Context {
				return nil
			},
			OK: false,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var ctx context.Context
			var ok bool
			if !a.So(func() {
				ctx, ok = ParentContext(tc.NewContext(tc.Parent))
			}, should.NotPanic) {
				t.FailNow()
			}

			if !tc.OK {
				a.So(ok, should.BeFalse)
				a.So(ctx, should.BeNil)
				return
			}

			a.So(ok, should.BeTrue)
			a.So(ctx, should.Equal, tc.Parent)
		})
	}
}

func TestContext(t *testing.T) {
	assertions.New(t).So(Context(), should.Equal, globalContext)
}
