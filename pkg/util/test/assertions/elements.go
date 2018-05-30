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

package assertions

import (
	"fmt"

	"github.com/kr/pretty"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

const (
	shouldHaveBeenEqual       = "Expected: '%v'\nActual:   '%v'\n(Should be equal)!"
	shouldHaveBeenEqualErr    = "Expected: '%v'\nActual:   '%v'\n(Should be equal but equality check errored with '%v')!"
	shouldHaveBeenEqualDiff   = "Expected: '%v'\nActual:   '%v'\nDiff:     '%v'\n(Should be equal)!"
	shouldNotHaveBeenEqual    = "Expected: '%v'\nActual:   '%v'\n(Should not be equal, but they were)!"
	shouldNotHaveBeenEqualErr = "Expected: '%v'\nActual:   '%v'\n(Should not be equal but equality check errored with '%v')!"
)

// ShouldHaveSameElements takes as arguments the actual value, the expected value and a
// comparison function.
// If the actual value equals the expected value using the comparison function, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldHaveSameElements(actual interface{}, expected ...interface{}) (message string) {
	defer func() {
		if r := recover(); r != nil {
			message = fmt.Sprintf(shouldHaveBeenEqualErr, expected, actual, r)
		}
	}()

	if message = need(2, expected); message != success {
		return
	}

	if !test.SameElements(expected[1], actual, expected[0]) {
		message = fmt.Sprintf(shouldHaveBeenEqual, expected[0], actual)
		return
	}

	message = success
	return
}

// ShouldNotHaveSameElements takes as arguments the actual value, the expected value and
// a comparison function.
// If the actual value does not equal the expected value using the comparison function,
// this function returns an empty string. Otherwise, it returns a string describing the
// error.
func ShouldNotHaveSameElements(actual interface{}, expected ...interface{}) (message string) {
	defer func() {
		if r := recover(); r != nil {
			message = fmt.Sprintf(shouldNotHaveBeenEqualErr, expected, actual, r)
		}
	}()

	if message = need(2, expected); message != success {
		return
	}

	if test.SameElements(expected[1], actual, expected[0]) {
		message = fmt.Sprintf(shouldNotHaveBeenEqual, expected[0], actual)
		return
	}

	message = success
	return
}

// ShouldHaveSameElementsDeep takes as arguments the actual value and the expected value.
// If the actual value equals the expected value using reflect.DeepEqual, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldHaveSameElementsDeep(actual interface{}, expected ...interface{}) (message string) {
	defer func() {
		if r := recover(); r != nil {
			message = fmt.Sprintf(shouldHaveBeenEqualErr, expected, actual, r)
		}
	}()

	if message = need(1, expected); message != success {
		return
	}

	if !test.SameElementsDeep(actual, expected[0]) {
		message = fmt.Sprintf(shouldHaveBeenEqual, expected, actual)
		return
	}

	message = success
	return
}

// ShouldNotHaveSameElementsDeep takes as arguments the actual value and the expected
// value.
// If the actual value does not equal the expected value using reflect.DeepEqual, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldNotHaveSameElementsDeep(actual interface{}, expected ...interface{}) (message string) {
	defer func() {
		if r := recover(); r != nil {
			message = fmt.Sprintf(shouldNotHaveBeenEqualErr, expected, actual, r)
		}
	}()

	if message = need(1, expected); message != success {
		return
	}

	if test.SameElementsDeep(actual, expected[0]) {
		message = fmt.Sprintf(shouldNotHaveBeenEqual, expected, actual)
		return
	}

	message = success
	return
}

// ShouldHaveSameElementsDiff takes as arguments the actual value and the expected value.
// If the actual value equals the expected value using pretty.Diff, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldHaveSameElementsDiff(actual interface{}, expected ...interface{}) (message string) {
	defer func() {
		if r := recover(); r != nil {
			message = fmt.Sprintf(shouldHaveBeenEqualErr, expected, actual, r)
		}
	}()

	if message = need(1, expected); message != success {
		return
	}

	if !test.SameElementsDiff(actual, expected[0]) {
		message = fmt.Sprintf(shouldHaveBeenEqualDiff, expected, actual, pretty.Diff(actual, expected[0]))
		return
	}

	message = success
	return
}

// ShouldNotHaveSameElementsDiff takes as arguments the actual value and the expected
// value.
// If the actual value does not equal the expected value using reflect.DeepEqual, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldNotHaveSameElementsDiff(actual interface{}, expected ...interface{}) (message string) {
	defer func() {
		if r := recover(); r != nil {
			message = fmt.Sprintf(shouldNotHaveBeenEqualErr, expected, actual, r)
		}
	}()

	if message = need(1, expected); message != success {
		return
	}

	if test.SameElementsDiff(actual, expected[0]) {
		message = fmt.Sprintf(shouldNotHaveBeenEqual, expected, actual)
		return
	}

	message = success
	return
}
