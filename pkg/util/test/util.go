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
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

func FormatError(err error) string {
	if err == nil {
		return "nil"
	}
	var s string
	for i, err := range errors.Stack(err) {
		s += fmt.Sprintf(`
%s-> %s (attributes: %v)`,
			strings.Repeat("-", i), err, errors.Attributes(err),
		)
	}
	return s
}

// Delay is the value, which can be used to slowdown the execution of time-dependent tests.
// You can assume, that most function calls will return in at most Delay time.
// It can(and should) be used to construct other time variables used in testing.
// Value may vary from machine to machine and can be overridden by TEST_SLOWDOWN environment variable.
var Delay = time.Millisecond * func() time.Duration {
	env := os.Getenv("TEST_SLOWDOWN")
	if env == "" {
		return 1
	}

	v, err := strconv.Atoi(env)
	if err != nil {
		return 1
	}
	return time.Duration(v)
}()

// Must returns v if err is nil and panics otherwise.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("Must received error: %v", FormatError(err)))
	}
	return v
}

// WaitTimeout returns true if f returns after at most d or false otherwise.
// An example of a f, for which this is useful would be Wait method of sync.WaitGroup.
// Note, this function leaks a goroutine if f never returns.
func WaitTimeout(d time.Duration, f func()) bool {
	done := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		f()
		close(done)
	}()
	wg.Wait()

	select {
	case <-time.After(d):
		return false
	case <-done:
		return true
	}
}

// WaitDeadline returns WaitTimeout(time.Until(t), f).
func WaitDeadline(t time.Time, f func()) bool {
	return WaitTimeout(time.Until(t), f)
}

// WaitContext returns true if f returns before <-ctx.Done() or false otherwise.
// An example of a f, for which this is useful would be Wait method of sync.WaitGroup.
// Note, this function leaks a goroutine if f never returns.
func WaitContext(ctx context.Context, f func()) bool {
	done := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		f()
		close(done)
	}()
	wg.Wait()

	select {
	case <-ctx.Done():
		return false
	case <-done:
		return true
	}
}

// AllTrue returns true iff v == true for each v in vs.
func AllTrue(vs ...bool) bool {
	for _, v := range vs {
		if !v {
			return false
		}
	}
	return true
}

// JoinStringsMap maps contents of xs to strings using f and joins them with sep.
func JoinStringsMap(f func(k interface{}, v interface{}) string, sep string, xs interface{}) string {
	r, ok := WrapRanger(xs)
	if !ok {
		panic(fmt.Sprintf("cannot range over values of type %T", xs))
	}
	var ss []string
	r.Range(func(k, v interface{}) bool {
		ss = append(ss, f(k, v))
		return true
	})
	return strings.Join(ss, sep)
}

// JoinStringsf formats contents of xs using format and joins them with sep.
func JoinStringsf(format, sep string, withKeys bool, xs interface{}) string {
	return JoinStringsMap(func(k, v interface{}) string {
		if withKeys {
			return fmt.Sprintf(format, k, v)
		}
		return fmt.Sprintf(format, v)
	}, sep, xs)
}
