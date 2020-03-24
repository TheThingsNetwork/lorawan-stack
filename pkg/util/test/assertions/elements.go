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
	"fmt"
	"reflect"

	"github.com/kr/pretty"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

const (
	shouldHaveBeenEqual          = "Expected: '%v'\nActual:   '%v'\n(Should be equal)!"
	shouldHaveBeenEqualDiff      = "Expected: '%v'\nActual:   '%v'\nDiff:     '%v'\n(Should be equal)!"
	shouldHaveBeenEqualErr       = "Expected: '%v'\nActual:   '%v'\n(Should be equal but equality check errored with '%v')!"
	shouldHaveHadSameElements    = "Expected: '%v'\nActual:   '%v'\n(Should have same elements)!"
	shouldNotHaveBeenEqual       = "Expected: '%v'\nActual:   '%v'\n(Should not be equal, but they were)!"
	shouldNotHaveBeenEqualErr    = "Expected: '%v'\nActual:   '%v'\n(Should not be equal but equality check errored with '%v')!"
	shouldNotHaveHadSameElements = "Expected: '%v'\nActual:   '%v'\n(Should not have same elements)!"
)

// diffEqual reports if pretty.Diff of x and y is empty.
func diffEqual(x, y interface{}) bool {
	return len(pretty.Diff(x, y)) == 0
}

// ranger represents an entity, which can be ranged over(e.g. sync.Map).
type ranger interface {
	Range(f func(k, v interface{}) bool)
}

type indexRanger struct {
	reflect.Value
}

func (rv indexRanger) Range(f func(k, v interface{}) bool) {
	for i := 0; i < rv.Len(); i++ {
		if !f(i, rv.Index(i).Interface()) {
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

func wrapRanger(v interface{}) (ranger, bool) {
	r, ok := v.(ranger)
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

// sameElements reports whether xs and ys represent the same multiset of elements
// under equality given by eq.
// Signature of eq must be func(A, B) bool, where A, B are types, which
// elements of xs and ys can be assigned to respectively.
// It panics if either xs or ys is not one of:
// 1. string, slice, array or map kind
// 2. value, which implements ranger interface(e.g. sync.Map)
// NOTE: Map key values are not taken into account.
func sameElements(eq interface{}, xs, ys interface{}) bool {
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

	type entry struct {
		value reflect.Value
		found int
	}
	entries := map[*entry]struct{}{}

	findEntry := func(v reflect.Value) *entry {
		for e := range entries {
			if ev.Call([]reflect.Value{e.value, v})[0].Bool() {
				return e
			}
		}
		return nil
	}

	xr.Range(func(_, v interface{}) bool {
		rv := reflect.ValueOf(v)
		e := findEntry(rv)
		if e == nil {
			entries[&entry{
				value: rv,
				found: 1,
			}] = struct{}{}
		} else {
			e.found++
		}
		return true
	})

	ok = true
	yr.Range(func(_, v interface{}) bool {
		rv := reflect.ValueOf(v)
		e := findEntry(rv)
		if e == nil {
			ok = false
			return false
		}
		if e.found == 1 {
			delete(entries, e)
		} else {
			e.found--
		}
		return true
	})
	return ok && len(entries) == 0
}

// ShouldHaveSameElementsFunc takes as arguments the actual value, a comparison function and the expected value.
// If the actual value equals the expected value using the comparison function, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldHaveSameElementsFunc(actual interface{}, expected ...interface{}) (message string) {
	defer func() {
		if r := recover(); r != nil {
			message = fmt.Sprintf(shouldHaveBeenEqualErr, expected[0], actual, r)
		}
	}()

	if message = need(2, expected); message != success {
		return
	}

	if !sameElements(expected[0], expected[1], actual) {
		return fmt.Sprintf(shouldHaveHadSameElements, expected[1], actual)
	}

	return success
}

// ShouldNotHaveSameElementsFunc takes as arguments the actual value, a comparison function and the expected value.
// If the actual value does not equal the expected value using the comparison function,
// this function returns an empty string. Otherwise, it returns a string describing the
// error.
func ShouldNotHaveSameElementsFunc(actual interface{}, expected ...interface{}) (message string) {
	defer func() {
		if r := recover(); r != nil {
			message = fmt.Sprintf(shouldNotHaveBeenEqualErr, expected[0], actual, r)
		}
	}()

	if message = need(2, expected); message != success {
		return
	}

	if sameElements(expected[0], expected[1], actual) {
		return fmt.Sprintf(shouldNotHaveBeenEqual, expected[1], actual)
	}

	return success
}

// ShouldHaveSameElementsDeep takes as arguments the actual value and the expected value.
// If the actual value equals the expected value using reflect.DeepEqual, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldHaveSameElementsDeep(actual interface{}, expected ...interface{}) (message string) {
	if message = need(1, expected); message != success {
		return
	}
	return ShouldHaveSameElementsFunc(actual, reflect.DeepEqual, expected[0])
}

// ShouldNotHaveSameElementsDeep takes as arguments the actual value and the expected
// value.
// If the actual value does not equal the expected value using reflect.DeepEqual, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldNotHaveSameElementsDeep(actual interface{}, expected ...interface{}) (message string) {
	if message = need(1, expected); message != success {
		return
	}
	return ShouldNotHaveSameElementsFunc(actual, reflect.DeepEqual, expected[0])
}

// ShouldHaveSameElementsDiff takes as arguments the actual value and the expected value.
// If the actual value equals the expected value using pretty.Diff, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldHaveSameElementsDiff(actual interface{}, expected ...interface{}) (message string) {
	if message = need(1, expected); message != success {
		return
	}
	return ShouldHaveSameElementsFunc(actual, diffEqual, expected[0])
}

// ShouldNotHaveSameElementsDiff takes as arguments the actual value and the expected
// value.
// If the actual value does not equal the expected value using reflect.DeepEqual, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldNotHaveSameElementsDiff(actual interface{}, expected ...interface{}) (message string) {
	if message = need(1, expected); message != success {
		return
	}
	return ShouldNotHaveSameElementsFunc(actual, diffEqual, expected[0])
}

// ShouldHaveSameElementsEvent takes as arguments the actual value and the expected value.
// If the actual value equals the expected value using test.EventEqual, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldHaveSameElementsEvent(actual interface{}, expected ...interface{}) (message string) {
	if message = need(1, expected); message != success {
		return
	}
	return ShouldHaveSameElementsFunc(actual, test.EventEqual, expected[0])
}

// ShouldNotHaveSameElementsEvent takes as arguments the actual value and the expected
// value.
// If the actual value does not equal the expected value using reflect.EventEqual, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldNotHaveSameElementsEvent(actual interface{}, expected ...interface{}) (message string) {
	if message = need(1, expected); message != success {
		return
	}
	return ShouldNotHaveSameElementsFunc(actual, test.EventEqual, expected[0])
}
