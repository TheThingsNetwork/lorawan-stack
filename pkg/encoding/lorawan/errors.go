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

package lorawan

import (
	"go.thethings.network/lorawan-stack/pkg/errors"
)

var (
	errDecode                       = errors.Define("decode", "could not decode `{lorawan_field}`")
	errEncode                       = errors.Define("encode", "could not encode `{lorawan_field}`")
	errEncodedFieldLengthBound      = errors.DefineInvalidArgument("encoded_field_length_bound", "`{lorawan_field}` encoded length `{value}` should between `{min}` and `{max}`", valueKey)
	errEncodedFieldLengthEqual      = errors.DefineInvalidArgument("encoded_field_length_equal", "`{lorawan_field}` encoded length `{value}` should be `{expected}`", valueKey)
	errEncodedFieldLengthTwoChoices = errors.DefineInvalidArgument("encoded_field_length_multiple_choices", "`{lorawan_field}` encoded length `{value}` should be equal to `{expected_1}` or `{expected_2}`", valueKey)
	errFieldBound                   = errors.DefineInvalidArgument("field_bound", "`{lorawan_field}` should be between `{min}` and `{max}`", valueKey)
	errFieldHasMax                  = errors.DefineInvalidArgument("field_with_max", "`{lorawan_field}` value `{value}` should be lower or equal to `{max}`", valueKey)
	errFieldLengthEqual             = errors.DefineInvalidArgument("field_length_equal", "`{lorawan_field}` length `{value}` should be equal to `{expected}`", valueKey)
	errFieldLengthHasMax            = errors.DefineInvalidArgument("field_length_with_max", "`{lorawan_field}` length `{value}` should be lower or equal to `{max}`", valueKey)
	errFieldLengthHasMin            = errors.DefineInvalidArgument("field_length_with_min", "`{lorawan_field}` length `{value}` should be higher or equal to `{min}`", valueKey)
	errFieldLengthTwoChoices        = errors.DefineInvalidArgument("field_length_with_multiple_choices", "`{lorawan_field}` length `{value}` should be equal to `{expected_1}` or `{expected_2}`", valueKey)
	errUnknownField                 = errors.DefineInvalidArgument("unknown_field", "unknown `{lorawan_field}`", valueKey)
)

const valueKey = "value"

type valueErr func(interface{}) errors.Error

func unexpectedValue(err interface {
	WithAttributes(kv ...interface{}) errors.Error
}) valueErr {
	return func(value interface{}) errors.Error {
		return err.WithAttributes(valueKey, value)
	}
}

func errExpectedLowerOrEqual(lorawanField string, max interface{}) valueErr {
	return unexpectedValue(errFieldHasMax.WithAttributes("lorawan_field", lorawanField, "max", max))
}

func errExpectedLengthEqual(lorawanField string, expected interface{}) valueErr {
	return unexpectedValue(errFieldLengthEqual.WithAttributes("lorawan_field", lorawanField, "expected", expected))
}

func errExpectedLengthLowerOrEqual(lorawanField string, max interface{}) valueErr {
	return unexpectedValue(errFieldLengthHasMax.WithAttributes("lorawan_field", lorawanField, "max", max))
}

func errExpectedLengthHigherOrEqual(lorawanField string, min interface{}) valueErr {
	return unexpectedValue(errFieldLengthHasMin.WithAttributes("lorawan_field", lorawanField, "min", min))
}

func errExpectedLengthTwoChoices(lorawanField string, expected1, expected2 interface{}) valueErr {
	return unexpectedValue(errFieldLengthTwoChoices.WithAttributes("lorawan_field", lorawanField, "expected_1", expected1, "expected_2", expected2))
}

func errExpectedLengthEncodedBound(lorawanField string, min, max interface{}) valueErr {
	return unexpectedValue(errEncodedFieldLengthBound.WithAttributes("lorawan_field", lorawanField, "min", min, "max", max))
}

func errExpectedLengthEncodedEqual(lorawanField string, expected interface{}) valueErr {
	return unexpectedValue(errEncodedFieldLengthEqual.WithAttributes("lorawan_field", lorawanField, "expected", expected))
}

func errExpectedLengthEncodedTwoChoices(lorawanField string, expected1, expected2 interface{}) valueErr {
	return unexpectedValue(errEncodedFieldLengthTwoChoices.WithAttributes("lorawan_field", lorawanField, "expected_1", expected1, "expected_2", expected2))
}

func errFailedEncoding(lorawanField string) errors.Error {
	return errEncode.WithAttributes("lorawan_field", lorawanField)
}

func errFailedDecoding(lorawanField string) errors.Error {
	return errDecode.WithAttributes("lorawan_field", lorawanField)
}

func errUnknown(lorawanField string) valueErr {
	return unexpectedValue(errUnknownField.WithAttributes("lorawan_field", lorawanField))
}

func errMissing(lorawanField string) errors.Error {
	return errUnknownField.WithAttributes("lorawan_field", lorawanField)
}

func errExpectedBetween(lorawanField string, min, max interface{}) valueErr {
	return unexpectedValue(errFieldBound.WithAttributes("lorawan_field", lorawanField, "min", min, "max", max))
}
