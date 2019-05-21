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
	"sync"
	"time"
)

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
func Must(v interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return v
}

// MustMultiple is like Must, but operates on arbitrary amount of values.
// It assumes that last value in vs is an error.
// It panics if len(vs) == 0.
func MustMultiple(vs ...interface{}) []interface{} {
	n := len(vs)
	if n == 0 {
		panic("MustMultiple requires at least 1 argument")
	}

	err, ok := vs[n-1].(error)
	if !ok && vs[n-1] != nil {
		panic(fmt.Sprintf("MustMultiple expected last argument to be an error, got %T", vs[n-1]))
	}

	if err != nil {
		panic(err)
	}
	return vs[:n-1]
}

// WaitTimeout returns true if f returns after at most d or false otherwise.
// An example of a f, for which this is useful would be Wait method of sync.WaitGroup.
// Note, this function leaks a goroutine if f never returns.
func WaitTimeout(d time.Duration, f func()) (ok bool) {
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
func WaitDeadline(t time.Time, f func()) (ok bool) {
	return WaitTimeout(time.Until(t), f)
}

// WaitContext returns true if f returns before <-ctx.Done() or false otherwise.
// An example of a f, for which this is useful would be Wait method of sync.WaitGroup.
// Note, this function leaks a goroutine if f never returns.
func WaitContext(ctx context.Context, f func()) (ok bool) {
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
