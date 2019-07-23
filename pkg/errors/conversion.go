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

package errors

import (
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/status"
)

// From returns an *Error if it can be derived from the given input.
// For a nil error, false will be returned.
func From(err error) (out *Error, ok bool) {
	if err == nil {
		return nil, false
	}
	defer func() {
		if out != nil {
			var copy = *out
			out = &copy
		}
	}()
	switch err := err.(type) {
	case Error:
		return &err, true
	case *Error:
		if err == nil {
			return nil, false // This is invalid.
		}
		return err, true
	case Definition:
		e := build(err, 0)
		return &e, true
	case *Definition:
		if err == nil {
			return nil, false // This is invalid.
		}
		e := build(*err, 0)
		return &e, true
	case ErrorDetails: // Received over an API.
		var e Error
		setErrorDetails(&e, err)
		return &e, true
	case interface{ GRPCStatus() *status.Status }:
		e := FromGRPCStatus(err.GRPCStatus())
		return &e, true
	case validationError:
		e := build(errValidation, 0).WithAttributes(
			"field", err.Field(),
			"reason", err.Reason(),
			"name", err.ErrorName(),
		)
		if cause := err.Cause(); cause != nil {
			e = e.WithCause(cause)
		}
		return &e, true
	}
	return nil, false
}

// ErrorDetails that can be carried over API.
type ErrorDetails interface {
	Error() string
	Namespace() string
	Name() string
	MessageFormat() string
	PublicAttributes() map[string]interface{}
	CorrelationID() string
	Cause() error
	Code() uint32
	Details() []proto.Message
}

func setErrorDetails(err *Error, details ErrorDetails) {
	if namespace := details.Namespace(); namespace != "" {
		err.namespace = namespace
	}
	if name := details.Name(); name != "" {
		err.name = name
	}
	if messageFormat := details.MessageFormat(); messageFormat != "" {
		err.messageFormat = messageFormat
		err.messageFormatArguments = messageFormatArguments(messageFormat)
		err.parsedMessageFormat, _ = formatter.Parse(messageFormat)
	}
	if attributes := details.PublicAttributes(); len(attributes) != 0 {
		err.attributes = attributes
		for attr := range attributes {
			err.publicAttributes = append(err.publicAttributes, attr)
		}
	}
	if correlationID := details.CorrelationID(); correlationID != "" {
		err.correlationID = correlationID
	}
	if cause := details.Cause(); cause != nil {
		err.cause, _ = From(cause)
	}
	if code := details.Code(); code != 0 {
		err.code = code
	}
	err.details = append(err.details, details.Details()...)
}
