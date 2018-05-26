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

package test_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	. "go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestContextParent(t *testing.T) {
	for _, tc := range []struct {
		Name       string
		NewContext func(context.Context) context.Context
		OK         bool
	}{
		{
			Name: "context.WithCancel",
			NewContext: func(ctx context.Context) context.Context {
				ctx, _ = context.WithCancel(ctx)
				return ctx
			},
			OK: true,
		},
		{
			Name: "context.WithValue",
			NewContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, struct{}{}, nil)
			},
			OK: true,
		},
		{
			Name: "context.WithDeadline",
			NewContext: func(ctx context.Context) context.Context {
				ctx, _ = context.WithDeadline(ctx, time.Now().Add(200*Delay))
				return ctx
			},
			OK: true,
		},
		{
			Name: "context.WithTimeout",
			NewContext: func(ctx context.Context) context.Context {
				ctx, _ = context.WithTimeout(ctx, 200*Delay)
				return ctx
			},
			OK: true,
		},
		{
			Name: "log.NewContext",
			NewContext: func(ctx context.Context) context.Context {
				return log.NewContext(ctx, log.Noop)
			},
			OK: true,
		},
		{
			Name: "rights.NewContext",
			NewContext: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, []ttnpb.Right{ttnpb.RIGHT_APPLICATION_DEVICES_WRITE})
			},
			OK: true,
		},
		{
			Name: "errorcontext.New",
			NewContext: func(ctx context.Context) context.Context {
				ctx, _ = errorcontext.New(ctx)
				return ctx
			},
			OK: true,
		},
		{
			Name: "context.Background",
			NewContext: func(ctx context.Context) context.Context {
				return context.Background()
			},
			OK: false,
		},
		{
			Name: "context.TODO",
			NewContext: func(ctx context.Context) context.Context {
				return context.TODO()
			},
			OK: false,
		},
		{
			Name: "nil",
			NewContext: func(context.Context) context.Context {
				return nil
			},
			OK: false,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			for _, parent := range []context.Context{
				nil, &MockContext{}, context.Background(),
			} {
				t.Run(fmt.Sprintf("%T", parent), func(t *testing.T) {
					a := assertions.New(t)

					defer func() {
						if r := recover(); r != nil {
							// Some functions in stdlib panic on nil context.Context passed.
							t.Skip("Skipping test case because of NewContext panic:", r)
						}
					}()
					ctx := tc.NewContext(parent)

					ctx, ok := ContextParent(ctx)
					if !tc.OK {
						a.So(ok, should.BeFalse)
						a.So(ctx, should.BeNil)
						return
					}

					a.So(ok, should.BeTrue)
					a.So(ctx, should.Equal, parent)
				})
			}
		})
	}
}

func TestContextRoot(t *testing.T) {
	for _, tc := range []struct {
		Name    string
		Context context.Context
		Root    context.Context
	}{
		{
			Name: "4",
			Context: context.WithValue(
				context.WithValue(
					context.WithValue(
						context.WithValue(
							Context(), "A", nil),
						"B", nil),
					struct{}{}, nil),
				struct{}{}, nil),
			Root: Context(),
		},
		{
			Name: "log.NewContext",
			Context: log.NewContext(
				log.NewContext(
					Context(), log.Noop),
				log.Noop,
			),
			Root: Context(),
		},
		{
			Name: "errorcontext.New",
			Context: log.NewContext(
				log.NewContext(
					Context(), log.Noop),
				log.Noop,
			),
			Root: Context(),
		},
		{
			Name:    "0",
			Context: Context(),
			Root:    Context(),
		},
		{
			Name:    "nil",
			Context: nil,
			Root:    nil,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			a.So(func() {
				a.So(ContextRoot(tc.Context), should.Resemble, tc.Root)
			}, should.NotPanic)
		})
	}
}

func TestContextHasParent(t *testing.T) {
	sharedCtx, _ := context.WithCancel(context.WithValue(Context(), struct{}{}, struct{}{}))

	for _, tc := range []struct {
		Name      string
		Context   context.Context
		Parent    context.Context
		HasParent bool
	}{
		{
			Name: "2",
			Context: context.WithValue(
				context.WithValue(
					sharedCtx, "A", nil),
				struct{ A int }{}, nil),
			Parent:    sharedCtx,
			HasParent: true,
		},
		{
			Name: "log.NewContext",
			Context: log.NewContext(
				log.NewContext(
					sharedCtx, log.Noop),
				log.Noop,
			),
			Parent:    sharedCtx,
			HasParent: true,
		},
		{
			Name: "not a parent",
			Context: context.WithValue(
				context.WithValue(
					context.WithValue(
						context.WithValue(
							Context(), "A", nil),
						"B", nil),
					struct{}{}, nil),
				struct{}{}, nil),
			Parent:    context.WithValue(context.Background(), struct{}{}, "asd"),
			HasParent: false,
		},
		{
			Name:      "0",
			Context:   sharedCtx,
			Parent:    sharedCtx,
			HasParent: false,
		},
		{
			Name:      "nil",
			Context:   nil,
			Parent:    nil,
			HasParent: false,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			a.So(func() {
				a.So(ContextHasParent(tc.Context, tc.Parent), should.Equal, tc.HasParent)
			}, should.NotPanic)
		})
	}
}

func TestContext(t *testing.T) {
	assertions.New(t).So(Context(), should.Equal, DefaultContext)
}
