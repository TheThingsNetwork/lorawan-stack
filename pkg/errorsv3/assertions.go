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

package errors

import (
	"fmt"
	"reflect"

	"github.com/smartystreets/assertions"
)

const (
	success                       = ""
	needExactValues               = "This assertion requires exactly %d comparison values (you provided %d)."
	needDefinitionCompatible      = "This assertion requires a Definition-compatible comparison type (you provided %T)."
	needErrorDefinitionCompatible = "This assertion requires an Error-compatible or Definition-compatible comparison type (you provided %T)."
	shouldBeDefinitionCompatible  = "Expected a Definition-compatible value (but was of type %T instead)!"
	shouldHaveNamespace           = "Expected error to have namespace '%v' (but it was '%v' instead)!"
	shouldHaveName                = "Expected error to have name '%v' (but it was '%v' instead)!"
	shouldHaveMessageFormat       = "Expected error to have message format '%v' (but it was '%v' instead)!"
	shouldHaveCode                = "Expected error to have code '%v' (but it was '%v' instead)!"
	shouldHaveAttributes          = "Expected error to have attributes '%v' (but it was '%v' instead)!"
	shouldHaveCause               = "Expected error to have cause '%v' (but it was '%v' instead)!"
	shouldHaveDetails             = "Expected error to have details '%v' (but it was '%v' instead)!"
)

type assertDefinitionCompatible interface {
	Namespace() string
	Name() string
	MessageFormat() string
	Code() int32
}

type assertErrorCompatible interface {
	assertDefinitionCompatible
	Attributes() map[string]interface{}
	Cause() error
	Details() (details []interface{})
}

// ShouldHaveSameDefinitionAs is used to assert that an error resembles the given Error or Definition.
func ShouldHaveSameDefinitionAs(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf(needExactValues, 1, len(expected))
	}
	expectedErr, ok := expected[0].(assertDefinitionCompatible)
	if !ok {
		return fmt.Sprintf(needDefinitionCompatible, actual)
	}
	actualErr, ok := expected[0].(assertDefinitionCompatible)
	if !ok {
		return fmt.Sprintf(shouldBeDefinitionCompatible, actual)
	}
	return assertDefinitionCompatibleEquals(actualErr, expectedErr)
}

func assertDefinitionCompatibleEquals(actual, expected assertDefinitionCompatible) string {
	if actual.Namespace() != expected.Namespace() {
		return fmt.Sprintf(shouldHaveNamespace, expected.Namespace(), actual.Namespace())
	}
	if actual.Name() != expected.Name() {
		return fmt.Sprintf(shouldHaveName, expected.Name(), actual.Name())
	}
	if actual.MessageFormat() != expected.MessageFormat() {
		return fmt.Sprintf(shouldHaveMessageFormat, expected.MessageFormat(), actual.MessageFormat())
	}
	if actual.Code() != expected.Code() {
		return fmt.Sprintf(shouldHaveCode, expected.Code(), actual.Code())
	}
	return success
}

func assertErrorCompatibleEquals(actual, expected assertErrorCompatible) string {
	if assertDefinition := assertDefinitionCompatibleEquals(actual, expected); assertDefinition != success {
		return assertDefinition
	}
	if !reflect.DeepEqual(actual.Attributes(), expected.Attributes()) {
		return fmt.Sprintf(shouldHaveAttributes, expected.Attributes(), actual.Attributes())
	}
	if actual.Cause() != expected.Cause() {
		return fmt.Sprintf(shouldHaveCause, expected.Cause(), actual.Cause())
	}
	if !reflect.DeepEqual(actual.Details(), expected.Details()) {
		return fmt.Sprintf(shouldHaveDetails, expected.Details(), actual.Details())
	}
	return success
}

// ShouldEqual is used to assert that an error equals the given Error or Definition.
func ShouldEqual(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf(needExactValues, 1, len(expected))
	}
	if assertType := assertions.ShouldHaveSameTypeAs(actual, expected...); assertType != success {
		return assertType
	}
	if actual == nil && expected[0] == nil {
		return success
	}
	switch expected := expected[0].(type) {
	case Error:
		return assertErrorCompatibleEquals(actual.(Error), expected)
	case *Error:
		return assertErrorCompatibleEquals(actual.(*Error), expected)
	case Definition:
		return assertDefinitionCompatibleEquals(actual.(Definition), expected)
	case *Definition:
		return assertDefinitionCompatibleEquals(actual.(*Definition), expected)
	default:
		return fmt.Sprintf(needErrorDefinitionCompatible, actual)
	}
}
