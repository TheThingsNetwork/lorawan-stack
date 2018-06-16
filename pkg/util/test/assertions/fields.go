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
	"go.thethings.network/lorawan-stack/pkg/marshaling"
)

const (
	shouldHaveBeenEqualFieldsErr    = "Expected: '%v'\nActual:   '%v'\n(Should be equal fields but equality check errored with '%v')!"
	shouldHaveBeenEqualFieldsDiff   = "Expected: '%v'\nActual:   '%v'\nDiff:     '%v'\n(Should be equal fields)!"
	shouldNotHaveBeenEqualFields    = "Expected: '%v'\nActual:   '%v'\n(Should not be equal fields, but they were)!"
	shouldNotHaveBeenEqualFieldsErr = "Expected: '%v'\nActual:   '%v'\n(Should not be equal fields but equality check errored with '%v')!"
)

// ShouldEqualFields takes as arguments the actual value and the expected value.
// If the actual value equals the expected value by checking the fields, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldEqualFields(actual interface{}, expected ...interface{}) (message string) {
	return ShouldEqualFieldsWithIgnores()(actual, expected...)
}

// ShouldNotEqualFields takes as arguments the actual value and the expected value.
// If the actual value does not equal the expected value by checking the fields, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldNotEqualFields(actual interface{}, expected ...interface{}) (message string) {
	return ShouldNotEqualFieldsWithIgnores()(actual, expected...)
}

// ShouldEqualFieldsWithIgnores takes fields to ignore and returns a function
// that takes as arguments the actual value and the expected value.
// If the actual value equals the expected value by checking the fields, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldEqualFieldsWithIgnores(ignores ...string) func(actual interface{}, expected ...interface{}) (message string) {
	return func(actual interface{}, expected ...interface{}) (message string) {
		if message = need(1, expected); message != success {
			return
		}

		ma, err := marshaling.MarshalMap(actual)
		if err != nil {
			return fmt.Sprintf(shouldHaveBeenEqualFieldsErr, actual, expected[0], err)
		}

		me, err := marshaling.MarshalMap(expected[0])
		if err != nil {
			return fmt.Sprintf(shouldHaveBeenEqualFieldsErr, actual, expected[0], err)
		}

		for _, ignore := range ignores {
			delete(ma, ignore)
			delete(me, ignore)
		}

		diff := pretty.Diff(ma, me)
		if len(diff) != 0 {
			return fmt.Sprintf(shouldHaveBeenEqualFieldsDiff, actual, expected[0], diff)
		}

		return success
	}
}

// ShouldNotEqualFieldsWithIgnores takes fields to ignore and returns a function
// that takes as arguments the actual value and the expected value.
// If the actual value does not equal the expected value by checking the fields, this
// function returns an empty string. Otherwise, it returns a string describing the error.
func ShouldNotEqualFieldsWithIgnores(ignores ...string) func(actual interface{}, expected ...interface{}) (message string) {
	return func(actual interface{}, expected ...interface{}) (message string) {
		if message = need(1, expected); message != success {
			return
		}

		ma, err := marshaling.MarshalMap(actual)
		if err != nil {
			return fmt.Sprintf(shouldNotHaveBeenEqualFieldsErr, actual, expected[0], err)
		}

		me, err := marshaling.MarshalMap(expected[0])
		if err != nil {
			return fmt.Sprintf(shouldNotHaveBeenEqualFieldsErr, actual, expected[0], err)
		}

		for _, ignore := range ignores {
			delete(ma, ignore)
			delete(me, ignore)
		}

		diff := pretty.Diff(ma, me)
		if len(diff) == 0 {
			return fmt.Sprintf(shouldNotHaveBeenEqualFields, actual, expected[0])
		}

		return success
	}
}
