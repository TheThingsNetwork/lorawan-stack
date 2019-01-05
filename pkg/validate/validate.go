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

// Package validate implements validation functions, usually used RPC message validation.
package validate

import "go.thethings.network/lorawan-stack/pkg/errors"

type validateFn func(v interface{}) error

var (
	errInvalidField = errors.DefineInvalidArgument("field", "invalid field `{name}`")
	errNotString    = errors.DefineInvalidArgument("not_string", "got `{type}` instead of string")
	errRequired     = errors.DefineInvalidArgument("required", "a value is required")
)

// All returns an error if one of the passed fields is invalid.
func All(fields ...error) error {
	for _, verified := range fields {
		if verified != nil {
			return verified
		}
	}
	return nil
}

// ValidationField implements error, and represents whether a field is valid or invalid.
type ValidationField struct {
	error
}

// Field verifies whether a field is valid, and returns an error if it is invalid.
func Field(v interface{}, verifiers ...func(interface{}) error) (vf *ValidationField) {
	for _, verifier := range verifiers {
		err := verifier(v)
		switch err {
		case errZeroValue:
			return nil
		case nil:
			continue
		default:
			vf = &ValidationField{error: err}
		}
	}
	return
}

// DescribeFieldName attaches the name of a field to the validation of a field, and returns it in an error format.
func (vf *ValidationField) DescribeFieldName(name string) error {
	if vf != nil && vf.error != nil {
		return errInvalidField.WithAttributes("name", name).WithCause(vf.error)
	}
	return nil
}
