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

func (e errorDetails) Error() string {
	return fmt.Sprintf("error:%s:%s (%s)", e.Namespace(), e.Name(), e.MessageFormat())
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
func (e errorDetails) Cause() error {
	cause := e.GetCause()
	if cause == nil {
		return nil
	}
	return errorDetails{cause}
}
func (e errorDetails) Code() uint32 { return e.GetCode() }

func ErrorDetailsToProto(e errors.ErrorDetails) *ErrorDetails {
	proto := &ErrorDetails{
		Namespace:     e.Namespace(),
		Name:          e.Name(),
		MessageFormat: e.MessageFormat(),
		CorrelationID: e.CorrelationID(),
		Code:          e.Code(),
	}
	if attributes := e.PublicAttributes(); len(attributes) > 0 {
		attributesStruct, err := gogoproto.Struct(attributes)
		if err != nil {
			panic(fmt.Sprintf("Failed to encode error attributes: %s", err)) // Likely a bug in ttn (invalid attribute type).
		}
		proto.Attributes = attributesStruct
	}
	if cause := e.Cause(); cause != nil {
		if ttnErr, ok := errors.From(cause); ok {
			proto.Cause = ErrorDetailsToProto(ttnErr)
		}
	}
	return proto
}

func ErrorDetailsFromProto(e *ErrorDetails) errors.ErrorDetails {
	d, ok := errors.From(errorDetails{ErrorDetails: e})
	if !ok {
		panic(fmt.Sprintf("Failed to decode error details")) // Likely a bug in ttn
	}
	return d
}

func init() {
	errors.ErrorDetailsToProto = func(e errors.ErrorDetails) proto.Message {
		return ErrorDetailsToProto(e)
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
		return ErrorDetailsFromProto(detailsMsg), rest
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
	errFieldHasMax        = errors.DefineInvalidArgument("field_with_max", "`{lorawan_field}` should be lower or equal to `{max}`", valueKey)
	errFieldBound         = errors.DefineInvalidArgument("field_bound", "`{lorawan_field}` should be between `{min}` and `{max}`", valueKey)
	errMissingIdentifiers = errors.DefineInvalidArgument("missing_identifiers", "missing identifiers")
	errParse              = errors.DefineInvalidArgument("parse", "could not parse `{lorawan_field}`", valueKey)
	errUnknownField       = errors.DefineInvalidArgument("unknown_field", "unknown `{lorawan_field}`", valueKey)
)

func errExpectedLowerOrEqual(lorawanField string, max interface{}) valueErr {
	return unexpectedValue(errFieldHasMax.WithAttributes("lorawan_field", lorawanField, "max", max))
}

func errExpectedBetween(lorawanField string, min, max interface{}) valueErr {
	return unexpectedValue(errFieldBound.WithAttributes("lorawan_field", lorawanField, "min", min, "max", max))
}

func errCouldNotParse(lorawanField string) valueErr {
	return unexpectedValue(errParse.WithAttributes("lorawan_field", lorawanField))
}

func errMissing(lorawanField string) errors.Error {
	return errUnknownField.WithAttributes("lorawan_field", lorawanField)
}
