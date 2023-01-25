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

	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

const (
	shouldHaveHadSameElements    = "Expected: '%v'\nActual:   '%v'\n(Should have same elements)!"
	shouldHaveHadSameElementsErr = "Expected: '%v'\nActual:   '%v'\n(Should have same elements, but equality check errored with '%v')!"

	shouldNotHaveHadSameElements    = "Expected: '%v'\nActual:   '%v'\n(Should not have same elements)!"
	shouldNotHaveHadSameElementsErr = "Expected: '%v'\nActual:   '%v'\n(Should not have same elements, but equality check errored with '%v')!"

	shouldHaveBeenProperSubsetOfElements    = "Expected: '%v'\nActual:   '%v'\n(Should represent proper subset of elements)!"
	shouldHaveBeenProperSubsetOfElementsErr = "Expected: '%v'\nActual:   '%v'\n(Should represent proper subset of elements, but equality check errored with '%v')!"
)

// ShouldHaveSameElementsFunc takes as arguments the actual value, a comparison function and the expected value.
// If the actual value equals the expected value using the comparison function, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldHaveSameElementsFunc(actual interface{}, expected ...interface{}) (message string) {
	defer func() {
		if r := recover(); r != nil {
			message = fmt.Sprintf(shouldHaveHadSameElementsErr, expected[0], actual, r)
		}
	}()
	if message = need(2, expected); message != success {
		return message
	}
	if !test.SameElements(expected[0], expected[1], actual) {
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
			message = fmt.Sprintf(shouldNotHaveHadSameElementsErr, expected[0], actual, r)
		}
	}()
	if message = need(2, expected); message != success {
		return message
	}
	if test.SameElements(expected[0], expected[1], actual) {
		return fmt.Sprintf(shouldNotHaveHadSameElements, expected[1], actual)
	}
	return success
}

// ShouldHaveSameElementsDeep takes as arguments the actual value and the expected value.
// If the actual value equals the expected value using test.DiffEqual, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldHaveSameElementsDeep(actual interface{}, expected ...interface{}) (message string) {
	if message = need(1, expected); message != success {
		return
	}
	return ShouldHaveSameElementsFunc(actual, test.DiffEqual, expected[0])
}

// ShouldNotHaveSameElementsDeep takes as arguments the actual value and the expected
// value.
// If the actual value does not equal the expected value using test.DiffEqual, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldNotHaveSameElementsDeep(actual interface{}, expected ...interface{}) (message string) {
	if message = need(1, expected); message != success {
		return
	}
	return ShouldNotHaveSameElementsFunc(actual, test.DiffEqual, expected[0])
}

// ShouldHaveSameElementsDiff takes as arguments the actual value and the expected value.
// If the actual value equals the expected value using test.Diff, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldHaveSameElementsDiff(actual interface{}, expected ...interface{}) (message string) {
	if message = need(1, expected); message != success {
		return
	}
	return ShouldHaveSameElementsFunc(actual, test.DiffEqual, expected[0])
}

// ShouldNotHaveSameElementsDiff takes as arguments the actual value and the expected
// value.
// If the actual value does not equal the expected value using test.Diff, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldNotHaveSameElementsDiff(actual interface{}, expected ...interface{}) (message string) {
	if message = need(1, expected); message != success {
		return
	}
	return ShouldNotHaveSameElementsFunc(actual, test.DiffEqual, expected[0])
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
// If the actual value does not equal the expected value using test.EventEqual, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldNotHaveSameElementsEvent(actual interface{}, expected ...interface{}) (message string) {
	if message = need(1, expected); message != success {
		return
	}
	return ShouldNotHaveSameElementsFunc(actual, test.EventEqual, expected[0])
}

// ShouldBeProperSupersetOfElementsFunc takes as arguments the actual value, a comparison function and the expected value.
// If the actual value represents a proper superset of expected value under equality given by the comparison function, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldBeProperSupersetOfElementsFunc(actual interface{}, expected ...interface{}) (message string) {
	defer func() {
		if r := recover(); r != nil {
			message = fmt.Sprintf(shouldHaveBeenProperSubsetOfElementsErr, expected[0], actual, r)
		}
	}()
	if message = need(2, expected); message != success {
		return message
	}
	if !test.IsProperSubsetOfElements(expected[0], expected[1], actual) {
		return fmt.Sprintf(shouldHaveBeenProperSubsetOfElements, expected[1], actual)
	}
	return success
}
