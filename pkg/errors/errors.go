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

// Package errors implements rich errors that carry contextual information
// such as stack traces, causality and attributes.
package errors

import (
	"fmt"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
)

// New returns an error that formats as the given text.
// This way of creating errors should be avoided if possible.
func New(text string) Error {
	return build(Definition{
		namespace:     namespace(2),
		messageFormat: text,
		code:          uint32(codes.Unknown),
	}, 4)
}

// Error is a rich error implementation.
type Error struct {
	Definition
	*stack
	correlationID string
	cause         error
	details       []proto.Message
	attributes    map[string]interface{}
	grpcStatus    *atomic.Value
}

func (e Error) String() string {
	return fmt.Sprintf("error:%s (%s)", e.FullName(), e.FormatMessage(e.PublicAttributes()))
}

// Error implements the error interface.
func (e Error) Error() string { return e.String() }

// Fields implements the log.Fielder interface.
func (e Error) Fields() map[string]interface{} {
	res := make(map[string]interface{})
	pref := "error_cause"
	for cause := Cause(e); cause != nil && !IsInternal(cause); cause = Cause(cause) {
		res[pref] = cause.Error()
		pref += "_cause"
	}
	return res
}

// Interface is the interface of an error.
type Interface interface {
	DefinitionInterface
	Attributes() map[string]interface{}
	Cause() error
	Details() (details []proto.Message)
}

// build an error from the definition, skipping the first frames of the call stack.
func build(d Definition, skip int) Error {
	e := Error{Definition: d, grpcStatus: new(atomic.Value)}
	if skip > 0 {
		e.stack = callers(skip)
	}
	return e
}
