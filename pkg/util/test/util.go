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
	"fmt"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/kr/pretty"
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

// DiffEqual reports if pretty.Diff of x and y is empty.
func DiffEqual(x, y interface{}) bool {
	return len(pretty.Diff(x, y)) == 0
}

// Ranger represents an entity, which can be ranged over(e.g. sync.Map).
type Ranger interface {
	Range(f func(k, v interface{}) bool)
}

type indexRanger struct {
	reflect.Value
}

func (rv indexRanger) Range(f func(k, v interface{}) bool) {
	for i := 0; i < rv.Len(); i++ {
		if !f(nil, rv.Index(i).Interface()) {
			return
		}
	}
}

type mapRanger struct {
	reflect.Value
}

func (rv mapRanger) Range(f func(k, v interface{}) bool) {
	for _, k := range rv.MapKeys() {
		if !f(k.Interface(), rv.MapIndex(k).Interface()) {
			return
		}
	}
}

func wrapRanger(v interface{}) (Ranger, bool) {
	r, ok := v.(Ranger)
	if ok {
		return r, ok
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Array:
		return indexRanger{rv}, true
	case reflect.Map:
		return mapRanger{rv}, true
	}
	return nil, false
}

// SameElements reports whether xs and ys represent the same multiset of elements
// under equality given by eq.
// Signature of eq must be func(A, B) bool, where A, B are types, which
// elements of xs and ys can be assigned to respectively.
// It panics if either xs or ys is not one of:
// 1. string, slice, array or map kind
// 2. value, which implements Ranger interface(e.g. sync.Map)
func SameElements(eq interface{}, xs, ys interface{}) bool {
	if xs == nil || ys == nil {
		return xs == ys
	}

	ev := reflect.ValueOf(eq)
	if ev.Kind() != reflect.Func {
		panic(fmt.Errorf("expected kind of eq to be a function, got: %s", ev.Kind()))
	}

	xr, ok := wrapRanger(xs)
	if !ok {
		panic(fmt.Errorf("cannot range over values of type %T", xs))
	}

	yr, ok := wrapRanger(ys)
	if !ok {
		panic(fmt.Errorf("cannot range over values of type %T", ys))
	}

	// NOTE: A hashmap cannot be used directly here, as []byte is unhashable.
	type entry struct {
		key    interface{}
		values []reflect.Value
		found  map[int]bool
	}
	var entries []*entry

	findEntry := func(k interface{}) *entry {
		for _, e := range entries {
			if reflect.DeepEqual(e.key, k) {
				return e
			}
		}
		return nil
	}

	xr.Range(func(k, v interface{}) bool {
		e := findEntry(k)
		if e == nil {
			e = &entry{
				key:   k,
				found: map[int]bool{},
			}
			entries = append(entries, e)
		}
		e.values = append(e.values, reflect.ValueOf(v))
		return true
	})

	ok = true
	yr.Range(func(k, yv interface{}) bool {
		e := findEntry(k)
		if e == nil {
			ok = false
			return false
		}

		for i, v := range e.values {
			if e.found[i] {
				continue
			}

			if ev.Call([]reflect.Value{v, reflect.ValueOf(yv)})[0].Bool() {
				ok = true
				e.found[i] = true
				return true
			}
		}
		ok = false
		return false
	})

	if !ok {
		return false
	}

	for _, e := range entries {
		for i := range e.values {
			if !e.found[i] {
				return false
			}
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
