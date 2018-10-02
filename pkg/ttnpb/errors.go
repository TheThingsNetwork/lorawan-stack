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

package ttnpb

import (
	"fmt"

	proto "github.com/golang/protobuf/proto"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gogoproto"
)

const valueKey = "value"

type errorDetails struct {
	*ErrorDetails
}

func (e errorDetails) Namespace() string     { return e.GetNamespace() }
func (e errorDetails) Name() string          { return e.GetName() }
func (e errorDetails) MessageFormat() string { return e.GetMessageFormat() }
func (e errorDetails) PublicAttributes() map[string]interface{} {
	attributes, err := gogoproto.Map(e.GetAttributes())
	if err != nil {
		panic(fmt.Sprintf("Failed to decode error attributes: %s", err)) // Likely a bug in gogoproto.
	}
	return attributes
}
func (e errorDetails) CorrelationID() string { return e.GetCorrelationID() }

func init() {
	errors.ErrorDetailsToProto = func(e errors.ErrorDetails) proto.Message {
		attributes, err := gogoproto.Struct(e.PublicAttributes())
		if err != nil {
			panic(fmt.Sprintf("Failed to encode error attributes: %s", err)) // Likely a bug in ttn (invalid attribute type).
		}
		return &ErrorDetails{
			Namespace:     e.Namespace(),
			Name:          e.Name(),
			MessageFormat: e.MessageFormat(),
			Attributes:    attributes,
			CorrelationID: e.CorrelationID(),
		}
	}
	errors.ErrorDetailsFromProto = func(msg ...proto.Message) (details errors.ErrorDetails, rest []proto.Message) {
		var detailsMsg *ErrorDetails
		for _, msg := range msg {
			switch msg := msg.(type) {
			case *ErrorDetails:
				detailsMsg = msg
			default:
				rest = append(rest, msg)
			}
		}
		details = errorDetails{detailsMsg}
		return
	}
}

type valueErr func(interface{}) errors.Error

func unexpectedValue(err interface {
	WithAttributes(kv ...interface{}) errors.Error
}) valueErr {
	return func(value interface{}) errors.Error {
		return err.WithAttributes(valueKey, value)
	}
}

var (
	errMissingIdentifiers = errors.DefineInvalidArgument("identifiers", "missing identifiers")

	errFieldBound                   = errors.DefineInvalidArgument("field_bound", "`{lorawan_field}` should be between `{min}` and `{max}`", valueKey)
	errFieldHasMax                  = errors.DefineInvalidArgument("field_with_max", "`{lorawan_field}` should be lower or equal to `{max}`", valueKey)
	errFieldLengthEqual             = errors.DefineInvalidArgument("field_length_equal", "`{lorawan_field}` length should be equal to `{expected}`", valueKey)
	errFieldLengthHasMax            = errors.DefineInvalidArgument("field_length_with_max", "`{lorawan_field}` length should be lower or equal to `{expected}`", valueKey)
	errFieldLengthHasMin            = errors.DefineInvalidArgument("field_length_with_min", "`{lorawan_field}` length should be higher or equal to `{expected}`", valueKey)
	errFieldLengthTwoChoices        = errors.DefineInvalidArgument("field_length_with_multiple_choices", "`{lorawan_field}` length should be equal to `{expected_1}` or `{expected_2}`", valueKey)
	errEncodedFieldLengthBound      = errors.DefineInvalidArgument("encoded_field_length_bound", "`{lorawan_field}` encoded length should between `{min}` and `{max}`", valueKey)
	errEncodedFieldLengthTwoChoices = errors.DefineInvalidArgument("encoded_field_length_multiple_choices", "`{lorawan_field}` encoded length should be equal to `{expected_1}` or `{expected_2}`", valueKey)

	errEncode       = errors.Define("encode", "could not encode `{lorawan_field}`")
	errDecode       = errors.Define("decode", "could not decode `{lorawan_field}`")
	errMissingField = errors.DefineInvalidArgument("missing_field", "missing `{lorawan_field}`")
	errParse        = errors.DefineInvalidArgument("parse", "could not parse `{lorawan_field}`", valueKey)
	errUnknownField = errors.DefineInvalidArgument("unknown_field", "unknown `{lorawan_field}`", valueKey)
)

func errExpectedBetween(lorawanField string, min, max interface{}) valueErr {
	return unexpectedValue(errFieldBound.WithAttributes("lorawan_field", lorawanField, "min", min, "max", max))
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
	return unexpectedValue(errEncodedFieldLengthBound.WithAttributes("lorawan_field", lorawanField, "expected", expected))
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

func errMissing(lorawanField string) errors.Error {
	return errUnknownField.WithAttributes("lorawan_field", lorawanField)
}

func errUnknown(lorawanField string) valueErr {
	return unexpectedValue(errUnknownField.WithAttributes("lorawan_field", lorawanField))
}

func errCouldNotParse(lorawanField string) valueErr {
	return unexpectedValue(errParse.WithAttributes("lorawan_field", lorawanField))
}
