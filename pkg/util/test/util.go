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
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/kr/pretty"
	"go.thethings.network/lorawan-stack/pkg/errors"
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

func mustErr(err error) error {
	return errors.Errorf("Error: %s", err)
}

// Must returns v if err is nil and panics otherwise.
func Must(v interface{}, err error) interface{} {
	if err != nil {
		panic(mustErr(err))
	}
	return v
}

// MustMultiple is like Must, but operates on arbitrary amount of values.
// It assumes that last value in vs is an error.
// It panics if len(vs) == 0.
func MustMultiple(vs ...interface{}) []interface{} {
	n := len(vs)
	if n == 0 {
		panic(errors.Errorf("MustMultiple requires at least 1 argument"))
	}

	err, ok := vs[n-1].(error)
	if !ok && vs[n-1] != nil {
		panic(errors.Errorf("MustMultiple expected last argument to be an error, got %T", vs[n-1]))
	}

	if err != nil {
		panic(mustErr(err))
	}
	return vs[:n-1]
}

// DiffEqual reports if pretty.Diff of x and y is empty.
func DiffEqual(x, y interface{}) bool {
	return len(pretty.Diff(x, y)) == 0
}

// SameElements reports whether xs and ys represent the same multiset of elements
// under equality given by eq.
// Signature of eq must be func(A, B) bool, where A, B are types, which
// elements of xs and ys can be assigned to respectively.
// It panics if reflect.Kind of xs or ys is not a slice.
func SameElements(eq interface{}, xs, ys interface{}) bool {
	ev := reflect.ValueOf(eq)
	if ev.Kind() != reflect.Func {
		panic(errors.Errorf("Expected kind of eq to be a function, got: %s", ev.Kind()))
	}

	xv := reflect.ValueOf(xs)
	if xv.Kind() != reflect.Slice {
		panic(errors.Errorf("Expected kind of xs to be a slice, got: %s", xv.Kind()))
	}

	yv := reflect.ValueOf(ys)
	if yv.Kind() != reflect.Slice {
		panic(errors.Errorf("Expected kind of ys to be a slice, got: %s", yv.Kind()))
	}

	n := xv.Len()
	if n != yv.Len() {
		return false
	}

	bm := make([]bool, n)

outer:
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if !bm[j] && ev.Call([]reflect.Value{xv.Index(i), yv.Index(j)})[0].Bool() {
				bm[j] = true
				continue outer
			}
		}
		return false
	}

	// Check if all values in ys have been marked
	for _, v := range bm {
		if !v {
			return false
		}
	}
	return true
}

// SameElementsDeep is like SameElements, but uses reflect.DeepEqual as eq.
func SameElementsDeep(xs, ys interface{}) bool {
	return SameElements(reflect.DeepEqual, xs, ys)
}

// SameElementsDiff is like SameElements, but uses DiffEqual as eq.
func SameElementsDiff(xs, ys interface{}) bool {
	return SameElements(DiffEqual, xs, ys)
}

// WaitTimeout returns true if fn returns after at most d or false otherwise.
// An example of a fn, for which this is useful would be Wait method of sync.WaitGroup.
// Note, this function leaks a goroutine if fn never returns.
func WaitTimeout(d time.Duration, fn func()) (ok bool) {
	done := make(chan struct{})

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		wg.Done()

		fn()
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
