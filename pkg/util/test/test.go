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
	"time"

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

var defaultTestTimeout = (1 << 15) * Delay

type TestConfig struct {
	Parallel bool
	Timeout  time.Duration
	Func     func(context.Context, *assertions.Assertion)
}

func runTestFromContext(ctx context.Context, conf TestConfig) {
	t := MustTFromContext(ctx)
	t.Helper()

	if conf.Parallel {
		// TODO: Enable once https://github.com/TheThingsNetwork/lorawan-stack/pull/3052 is merged.
		// t.Parallel()
	}
	timeout := conf.Timeout
	if timeout == 0 {
		timeout = defaultTestTimeout
	}
	a, ctx := New(t)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dl, ok := ctx.Deadline()
	if !ok {
		panic("missing deadline in context")
	}
	timeout = time.Until(dl)

	start := time.Now()
	doneCh := make(chan struct{})
	defer func() {
		t.Helper()
		close(doneCh)
		if d := time.Since(start); d > timeout {
			t.Errorf("%s took too long to execute. Expected execution time below %v, ran for %v", t.Name(), timeout, d)
		}
	}()
	go func() {
		for {
			select {
			case <-doneCh:
				return
			case <-time.Tick(timeout / 4):
				t.Logf("%s is taking a long time to execute. Expected execution time below %v, running already for: %v", t.Name(), timeout, time.Since(start))
			}
		}
	}()
	conf.Func(ctx, a)
}

func RunTest(t *testing.T, conf TestConfig) {
	t.Helper()
	_, ctx := New(t)
	runTestFromContext(ctx, conf)
}

var defaultSubtestTimeout = defaultTestTimeout / 4

type SubtestConfig struct {
	Name     string
	Parallel bool
	Timeout  time.Duration
	Func     func(context.Context, *testing.T, *assertions.Assertion)
}

func RunSubtestFromContext(ctx context.Context, conf SubtestConfig) bool {
	t := MustTFromContext(ctx)
	t.Helper()
	return t.Run(conf.Name, func(t *testing.T) {
		t.Helper()

		timeout := conf.Timeout
		if timeout == 0 {
			timeout = defaultSubtestTimeout
		}
		_, ctx = NewWithContext(ctx, t)
		runTestFromContext(ctx, TestConfig{
			Parallel: conf.Parallel,
			Timeout:  timeout,
			Func: func(ctx context.Context, a *assertions.Assertion) {
				t.Helper()
				conf.Func(ctx, t, a)
			},
		})
	})
}

func RunSubtest(t *testing.T, conf SubtestConfig) bool {
	t.Helper()
	_, ctx := New(t)
	return RunSubtestFromContext(ctx, conf)
}
