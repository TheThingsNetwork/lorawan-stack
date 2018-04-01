// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"reflect"

	"github.com/pkg/errors"
)

// SameElementsFunc is like SameElementsFunc, but uses reflect.DeepEqual as eq.
func SameElements(xs, ys interface{}) bool {
	return SameElementsFunc(reflect.DeepEqual, xs, ys)
}

// SameElementsFunc reports whether xs and ys represent the same multiset of elements
// under equality given by eq.
// Signature of eq must be func(A, B) bool, where A, B are types, which
// elements of xs and ys can be assigned to respectively.
// It panics if reflect.Kind of xs or ys is not a slice.
func SameElementsFunc(eq interface{}, xs, ys interface{}) bool {
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

	yn := yv.Len()
	bm := make([]bool, yn)

outer:
	for i := 0; i < xv.Len(); i++ {
		for j := 0; j < yn; j++ {
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
