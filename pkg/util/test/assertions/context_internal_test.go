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

package assertions

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/pkg/log"
)

func TestContextParent(t *testing.T) {
	var cancels []context.CancelFunc
	defer func() {
		for _, cancel := range cancels {
			cancel()
		}
	}()

	for _, tc := range []struct {
		Name       string
		NewContext func(context.Context) context.Context
		OK         bool
	}{
		{
			Name: "context.WithCancel",
			NewContext: func(ctx context.Context) context.Context {
				ctx, cancel := context.WithCancel(ctx)
				cancels = append(cancels, cancel)
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
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(200*time.Hour))
				cancels = append(cancels, cancel)
				return ctx
			},
			OK: true,
		},
		{
			Name: "context.WithTimeout",
			NewContext: func(ctx context.Context) context.Context {
				ctx, cancel := context.WithTimeout(ctx, 200*time.Hour)
				cancels = append(cancels, cancel)
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
			Name: "rights.NewContextWithFetcher",
			NewContext: func(ctx context.Context) context.Context {
				return rights.NewContextWithFetcher(ctx, nil)
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
				nil, context.Background(),
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

					ctx, ok := contextParent(ctx)
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
	var root = context.Background()

	for _, tc := range []struct {
		Name    string
		Context context.Context
	}{
		{
			Name: "4",
			Context: context.WithValue(
				context.WithValue(
					context.WithValue(
						context.WithValue(
							root, "A", nil),
						"B", nil),
					struct{}{}, nil),
				struct{}{}, nil),
		},
		{
			Name: "log.NewContext",
			Context: log.NewContext(
				log.NewContext(
					root, log.Noop),
				log.Noop,
			),
		},
		{
			Name: "errorcontext.New",
			Context: func() (ctx context.Context) {
				ctx, _ = errorcontext.New(root)
				ctx, _ = errorcontext.New(ctx)
				return
			}(),
		},
		{
			Name:    "0",
			Context: root,
		},
		{
			Name:    "nil",
			Context: nil,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			a.So(func() {
				expected := root
				if tc.Context == nil {
					expected = nil
				}
				a.So(contextRoot(tc.Context), should.Equal, expected)
			}, should.NotPanic)
		})
	}
}

func TestContextHasParent(t *testing.T) {
	sharedCtx, cancel := context.WithCancel(context.WithValue(context.Background(), struct{}{}, struct{}{}))
	defer cancel()

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
							context.Background(), "A", nil),
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
				a.So(contextHasParent(tc.Context, tc.Parent), should.Equal, tc.HasParent)
			}, should.NotPanic)
		})
	}
}

func TestShouldHaveParentContext(t *testing.T) {
	for i, tc := range []struct {
		Actual,
		Expected interface{}
		Test func(actual interface{}, expected ...interface{}) string
	}{
		{
			Actual:   "string",
			Expected: context.Background(),
			Test:     should.NotBeEmpty,
		},
		{
			Actual:   context.Background(),
			Expected: "string",
			Test:     should.NotBeEmpty,
		},
		{
			Actual:   context.Background(),
			Expected: context.Background(),
			Test:     should.NotBeEmpty,
		},
		{
			Actual:   context.WithValue(context.Background(), struct{}{}, struct{}{}),
			Expected: context.Background(),
			Test:     should.BeEmpty,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			msg := ShouldHaveParentContext(tc.Actual, tc.Expected)
			a.So(msg, tc.Test)
			if tc.Actual == tc.Expected {
				a.So(ShouldHaveParentContextOrEqual(tc.Actual, tc.Expected), should.BeEmpty)
			} else {
				a.So(ShouldHaveParentContextOrEqual(tc.Actual, tc.Expected), tc.Test)
			}
		})
	}
}
