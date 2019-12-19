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

// Package should wraps assertions in github.com/smartystreets/assertions and util/test/assertions.
package should

import (
	"github.com/smartystreets/assertions"
	testassertions "go.thethings.network/lorawan-stack/pkg/util/test/assertions"
)

var (
	// AlmostEqual makes sure that two parameters are close enough to being equal. The acceptable delta may be specified with a third argument, or a very small default delta will be used.
	AlmostEqual = assertions.ShouldAlmostEqual
	// BeBetween receives exactly three parameters: an actual value, a lower bound, and an upper bound. It ensures that the actual value is between both bounds (but not equal to either of them).
	BeBetween = assertions.ShouldBeBetween
	// BeBetweenOrEqual receives exactly three parameters: an actual value, a lower bound, and an upper bound. It ensures that the actual value is between both bounds or equal to one of them.
	BeBetweenOrEqual = assertions.ShouldBeBetweenOrEqual
	// BeBlank receives exactly 1 string parameter and ensures that it is equal to "".
	BeBlank = assertions.ShouldBeBlank
	// BeChronological receives a []time.Time slice and asserts that the are in chronological order starting with the first time.Time as the earliest.
	BeChronological = assertions.ShouldBeChronological
	// BeEmpty receives a single parameter (actual) and determines whether or not calling len(actual) would return `0`. It obeys the rules specified by the len function for determining length: http:// golang.org/pkg/builtin/#len
	BeEmpty = assertions.ShouldBeEmpty
	// BeError asserts that the first argument implements the error interface. It also compares the first argument against the second argument if provided (which must be an error message string or another error value).
	BeError = assertions.ShouldBeError
	// BeFalse receives a single parameter and ensures that it is false.
	BeFalse = assertions.ShouldBeFalse
	// BeGreaterThan receives exactly two parameters and ensures that the first is greater than the second.
	BeGreaterThan = assertions.ShouldBeGreaterThan
	// BeGreaterThanOrEqualTo receives exactly two parameters and ensures that the first is greater than or equal to the second.
	BeGreaterThanOrEqualTo = assertions.ShouldBeGreaterThanOrEqualTo
	// BeIn receives at least 2 parameters. The first is a proposed member of the collection that is passed in either as the second parameter, or of the collection that is comprised of all the remaining parameters. This assertion ensures that the proposed member is in the collection (usingEqual).
	BeIn = assertions.ShouldBeIn
	// BeLessThan receives exactly two parameters and ensures that the first is less than the second.
	BeLessThan = assertions.ShouldBeLessThan
	// BeLessThanOrEqualTo receives exactly two parameters and ensures that the first is less than or equal to the second.
	BeLessThanOrEqualTo = assertions.ShouldBeLessThanOrEqualTo
	// BeNil receives a single parameter and ensures that it is nil.
	BeNil = assertions.ShouldBeNil
	// BeTrue receives a single parameter and ensures that it is true.
	BeTrue = assertions.ShouldBeTrue
	// BeZeroValue receives a single parameter and ensures that it is the Go equivalent of the default value, or "zero" value.
	BeZeroValue = assertions.ShouldBeZeroValue
	// Contain receives exactly two parameters. The first is a slice and the second is a proposed member. Membership is determined usingEqual.
	Contain = assertions.ShouldContain
	// ContainKey receives exactly two parameters. The first is a map and the second is a proposed key. Keys are compared with a simple '=='.
	ContainKey = assertions.ShouldContainKey
	// ContainSubstring receives exactly 2 string parameters and ensures that the first contains the second as a substring.
	ContainSubstring = assertions.ShouldContainSubstring
	// EndWith receives exactly 2 string parameters and ensures that the first ends with the second.
	EndWith = assertions.ShouldEndWith
	// Equal receives exactly two parameters and does an equality check using the following semantics: 1. If the expected and actual values implement an Equal method in the form `func (this T) Equal(that T) bool` then call the method. If true, they are equal. 2. The expected and actual values are judged equal or not by oglematchers.Equals.
	Equal = assertions.ShouldEqual
	// EqualJSON receives exactly two parameters and does an equality check by marshalling to JSON.
	EqualJSON = assertions.ShouldEqualJSON
	// EqualTrimSpace receives exactly 2 string parameters and ensures that the first is equal to the second after removing all leading and trailing whitespace using strings.TrimSpace(first).
	EqualTrimSpace = assertions.ShouldEqualTrimSpace
	// EqualWithout receives exactly 3 string parameters and ensures that the first is equal to the second after removing all instances of the third from the first using strings.Replace(first, third, "", -1).
	EqualWithout = assertions.ShouldEqualWithout
	// HappenAfter receives exactly 2 time.Time arguments and asserts that the first happens after the second.
	HappenAfter = assertions.ShouldHappenAfter
	// HappenBefore receives exactly 2 time.Time arguments and asserts that the first happens before the second.
	HappenBefore = assertions.ShouldHappenBefore
	// HappenBetween receives exactly 3 time.Time arguments and asserts that the first happens between (not on) the second and third.
	HappenBetween = assertions.ShouldHappenBetween
	// HappenOnOrAfter receives exactly 2 time.Time arguments and asserts that the first happens on or after the second.
	HappenOnOrAfter = assertions.ShouldHappenOnOrAfter
	// HappenOnOrBefore receives exactly 2 time.Time arguments and asserts that the first happens on or before the second.
	HappenOnOrBefore = assertions.ShouldHappenOnOrBefore
	// HappenOnOrBetween receives exactly 3 time.Time arguments and asserts that the first happens between or on the second and third.
	HappenOnOrBetween = assertions.ShouldHappenOnOrBetween
	// HappenWithin receives a time.Time, a time.Duration, and a time.Time (3 arguments) and asserts that the first time.Time happens within or on the duration specified relative to the other time.Time.
	HappenWithin = assertions.ShouldHappenWithin
	// HaveLength receives 2 parameters. The first is a collection to check the length of, the second being the expected length. It obeys the rules specified by the len function for determining length: http:// golang.org/pkg/builtin/#len
	HaveLength = assertions.ShouldHaveLength
	// HaveSameTypeAs receives exactly two parameters and compares their underlying types for equality.
	HaveSameTypeAs = assertions.ShouldHaveSameTypeAs
	// Implement receives exactly two parameters and ensures that the first implements the interface type of the second.
	Implement = assertions.ShouldImplement
	// NotAlmostEqual is the inverse ofAlmostEqual
	NotAlmostEqual = assertions.ShouldNotAlmostEqual
	// NotBeBetween receives exactly three parameters: an actual value, a lower bound, and an upper bound. It ensures that the actual value is NOT between both bounds.
	NotBeBetween = assertions.ShouldNotBeBetween
	// NotBeBetweenOrEqual receives exactly three parameters: an actual value, a lower bound, and an upper bound. It ensures that the actual value is nopt between the bounds nor equal to either of them.
	NotBeBetweenOrEqual = assertions.ShouldNotBeBetweenOrEqual
	// NotBeBlank receives exactly 1 string parameter and ensures that it is equal to "".
	NotBeBlank = assertions.ShouldNotBeBlank
	// NotBeChronological receives a []time.Time slice and asserts that they are
	// NOT in chronological order.
	NotBeChronological = assertions.ShouldNotBeChronological
	// NotBeEmpty receives a single parameter (actual) and determines whether or not calling len(actual) would return a value greater than zero. It obeys the rules specified by the `len` function for determining length: http:// golang.org/pkg/builtin/#len
	NotBeEmpty = assertions.ShouldNotBeEmpty
	// NotBeIn receives at least 2 parameters. The first is a proposed member of the collection that is passed in either as the second parameter, or of the collection that is comprised of all the remaining parameters. This assertion ensures that the proposed member is NOT in the collection (usingEqual).
	NotBeIn = assertions.ShouldNotBeIn
	// NotBeNil receives a single parameter and ensures that it is not nil.
	NotBeNil = assertions.ShouldNotBeNil
	// NotBeZeroValue receives a single parameter and ensures that it is NOT
	// the Go equivalent of the default value, or "zero" value.
	NotBeZeroValue = assertions.ShouldNotBeZeroValue
	// NotContain receives exactly two parameters. The first is a slice and the second is a proposed member. Membership is determinied usingEqual.
	NotContain = assertions.ShouldNotContain
	// NotContainKey receives exactly two parameters. The first is a map and the second is a proposed absent key. Keys are compared with a simple '=='.
	NotContainKey = assertions.ShouldNotContainKey
	// NotContainSubstring receives exactly 2 string parameters and ensures that the first does NOT contain the second as a substring.
	NotContainSubstring = assertions.ShouldNotContainSubstring
	// NotEndWith receives exactly 2 string parameters and ensures that the first does not end with the second.
	NotEndWith = assertions.ShouldNotEndWith
	// NotEqual receives exactly two parameters and does an inequality check. SeeEqual for details on how equality is determined.
	NotEqual = assertions.ShouldNotEqual
	// NotHappenOnOrBetween receives exactly 3 time.Time arguments and asserts that the first does NOT happen between or on the second or third.
	NotHappenOnOrBetween = assertions.ShouldNotHappenOnOrBetween
	// NotHappenWithin receives a time.Time, a time.Duration, and a time.Time (3 arguments) and asserts that the first time.Time does NOT happen within or on the duration specified relative to the other time.Time.
	NotHappenWithin = assertions.ShouldNotHappenWithin
	// NotHaveSameTypeAs receives exactly two parameters and compares their underlying types for inequality.
	NotHaveSameTypeAs = assertions.ShouldNotHaveSameTypeAs
	// NotImplement receives exactly two parameters and ensures that the first does NOT implement the interface type of the second.
	NotImplement = assertions.ShouldNotImplement
	// NotPanic receives a void, niladic function and expects to execute the function without any panic.
	NotPanic = assertions.ShouldNotPanic
	// NotPanicWith receives a void, niladic function and expects to recover a panic whose content differs from the second argument.
	NotPanicWith = assertions.ShouldNotPanicWith
	// NotPointTo receives exactly two parameters and checks to see that they point to different addresess.
	NotPointTo = assertions.ShouldNotPointTo
	// NotResemble receives exactly two parameters and does an inverse deep equal check (see reflect.DeepEqual)
	NotResemble = assertions.ShouldNotResemble
	// NotStartWith receives exactly 2 string parameters and ensures that the first does not start with the second.
	NotStartWith = assertions.ShouldNotStartWith
	// Panic receives a void, niladic function and expects to recover a panic.
	Panic = assertions.ShouldPanic
	// PanicWith receives a void, niladic function and expects to recover a panic with the second argument as the content.
	PanicWith = assertions.ShouldPanicWith
	// PointTo receives exactly two parameters and checks to see that they point to the same address.
	PointTo = assertions.ShouldPointTo
	// Resemble receives exactly two parameters and does a deep equal check (see reflect.DeepEqual)
	Resemble = testassertions.ShouldResemble
	// StartWith receives exactly 2 string parameters and ensures that the first starts with the second.
	StartWith = assertions.ShouldStartWith

	// HaveSameElements asserts that the actual A and expected B elements are equal using an equality function with signature func(A, B) bool.
	HaveSameElements = testassertions.ShouldHaveSameElementsFunc
	// NotHaveSameElements asserts that the actual A and expected B elements are not equal using an equality function with signature func(A, B) bool.
	NotHaveSameElements = testassertions.ShouldNotHaveSameElementsFunc
	// HaveSameElementsDeep asserts that the actual A and expected B elements are equal using reflect.Equal.
	HaveSameElementsDeep = testassertions.ShouldHaveSameElementsDeep
	// NotHaveSameElementsDeep asserts that the actual A and expected B elements are not equal using reflect.Equal.
	NotHaveSameElementsDeep = testassertions.ShouldNotHaveSameElementsDeep
	// HaveSameElementsDiff asserts that the actual A and expected B elements are equal using pretty.Diff.
	HaveSameElementsDiff = testassertions.ShouldHaveSameElementsDiff
	// NotHaveSameElementsDiff asserts that the actual A and expected B elements are not equal using pretty.Diff.
	NotHaveSameElementsDiff = testassertions.ShouldNotHaveSameElementsDiff
	// HaveParentContext asserts that the context.Context is a child of context.Context.
	HaveParentContext = testassertions.ShouldHaveParentContext
	// HaveParentContextOrEqual asserts that the context.Context is a child of context.Context or they're equal.
	HaveParentContextOrEqual = testassertions.ShouldHaveParentContextOrEqual
	// HaveSameErrorDefinitionAs asserts that the error definitions of the actual and expected arguments are the same.
	HaveSameErrorDefinitionAs = testassertions.ShouldHaveSameErrorDefinitionAs
	// EqualErrorOrDefinition asserts that the actual and expected arguments are of the same type (error or definition),
	// and that they have the same underlying definition, as well as arguments if they are both errors.
	EqualErrorOrDefinition = testassertions.ShouldEqualErrorOrDefinition
	// HaveEmptyDiff receives exactly two parameters and does an equality check using pretty.Diff.
	HaveEmptyDiff = testassertions.ShouldHaveEmptyDiff
	// NotHaveEmptyDiff receives exactly two parameters and does an inequality check using pretty.Diff.
	NotHaveEmptyDiff = testassertions.ShouldNotHaveEmptyDiff
	// HaveRoute asserts that the given *echo.Echo server has a route with the given method and path.
	HaveRoute = testassertions.ShouldHaveRoute
	// NotHaveRoute asserts that the given *echo.Echo server does not have a route with the given method and path.
	NotHaveRoute = testassertions.ShouldNotHaveRoute

	// ResembleEvent receives exactly two events.Event and does a resemblance check.
	ResembleEvent = testassertions.ShouldResembleEvent
	// ResembleEventDefinitionDataClosure receives exactly two events.DefinitionDataClosure and does a resemblance check.
	ResembleEventDefinitionDataClosure = testassertions.ShouldResembleEventDefinitionDataClosure
	// ResembleEventDefinitionDataClosures receives exactly two []events.DefinitionDataClosure and does a resemblance check.
	ResembleEventDefinitionDataClosures = testassertions.ShouldResembleEventDefinitionDataClosures
)
