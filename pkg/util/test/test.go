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

// Package test provides various testing utilities.
package test

import (
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
)

func NewWithContext(ctx context.Context, tb testing.TB) (*assertions.Assertion, context.Context) {
	tb.Helper()
	return assertions.New(tb), ContextWithTB(
		log.NewContext(
			ctx, GetLogger(tb),
		),
		tb,
	)
}

func New(tb testing.TB) (*assertions.Assertion, context.Context) {
	tb.Helper()
	return NewWithContext(Context(), tb)
}

func NewTBFromContext(ctx context.Context) (testing.TB, *assertions.Assertion, bool) {
	tb, ok := TBFromContext(ctx)
	if !ok {
		return nil, nil, false
	}
	tb.Helper()
	return tb, assertions.New(tb), true
}

func MustNewTBFromContext(ctx context.Context) (testing.TB, *assertions.Assertion) {
	tb := MustTBFromContext(ctx)
	tb.Helper()
	return tb, assertions.New(tb)
}

func NewTFromContext(ctx context.Context) (*testing.T, *assertions.Assertion, bool) {
	t, ok := TFromContext(ctx)
	if !ok {
		return nil, nil, false
	}
	t.Helper()
	return t, assertions.New(t), true
}

func MustNewTFromContext(ctx context.Context) (*testing.T, *assertions.Assertion) {
	t := MustTFromContext(ctx)
	t.Helper()
	return t, assertions.New(t)
}
