// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"fmt"
)

const (
	success             = ""
	needExactValues     = "This assertion requires exactly %d comparison values (you provided %d)."
	needDescriptor      = "This assertion requires ErrDescriptor as comparison type (you provided %T)."
	shouldBeErrorType   = "Expected a known error value (but was of type %T instead)!"
	shouldHaveNamespace = "Expected error to have namespace '%v' (but it was '%v' instead)!"
	shouldHaveCode      = "Expected error to have code '%v' (but it was '%v' instead)!"
	shouldNotDescribe   = "Expected error to not describe '%v' (but it does)!"
)

func ShouldDescribe(actual interface{}, expected ...interface{}) string {
	err, ok := actual.(Error)
	if !ok {
		return fmt.Sprintf(shouldBeErrorType, actual)
	}
	if len(expected) != 1 {
		return fmt.Sprintf(needExactValues, 1, len(expected))
	}
	descriptor, ok := expected[0].(*ErrDescriptor)
	if !ok {
		return fmt.Sprintf(needDescriptor, expected[0])
	}

	if err.Namespace() != descriptor.Namespace {
		return fmt.Sprintf(shouldHaveNamespace, descriptor.Namespace, err.Namespace())
	}
	if err.Code() != descriptor.Code {
		return fmt.Sprintf(shouldHaveCode, descriptor.Code, err.Code())
	}

	return success
}

func ShouldNotDescribe(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf(needExactValues, 1, len(expected))
	}
	if ShouldDescribe(actual, expected...) == success {
		return fmt.Sprintf(shouldNotDescribe, expected[0])
	}
	return success
}
