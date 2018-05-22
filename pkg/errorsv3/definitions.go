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
	"regexp"

	"google.golang.org/grpc/codes"
)

// Definition of a registered error.
type Definition struct {
	namespace              string
	name                   string
	messageFormat          string
	messageFormatArguments []string
	publicAttributes       []string
	code                   int32 // 0 is invalid; so implies Unknown (code 2)
}

// Namespace of the error.
func (d Definition) Namespace() string { return d.namespace }

// Name of the error.
func (d Definition) Name() string { return d.name }

// MessageFormat of the error.
func (d Definition) MessageFormat() string { return d.messageFormat }

// Code of the error.
// This code is consistent with google.golang.org/genproto/googleapis/rpc/code and google.golang.org/grpc/codes.
func (d Definition) Code() int32 { return d.code }

func (d Definition) String() string {
	if d.namespace == "" || d.name == "" {
		return d.messageFormat
	}
	return fmt.Sprintf("error:%s:%s", d.namespace, d.name)
}

// Error implements the error interface.
func (d Definition) Error() string { return d.String() }

var messageFormatArgument = regexp.MustCompile(`\{[\s]*([a-z0-9_]+)`)

func messageFormatArguments(messageFormat string) (args []string) {
	for _, matches := range messageFormatArgument.FindAllStringSubmatch(messageFormat, -1) {
		if len(matches) == 2 {
			args = append(args, matches[1])
		}
	}
	m := make(map[string]struct{}, len(args))
	for _, arg := range args {
		m[arg] = struct{}{}
	}
	args = make([]string, 0, len(m))
	for arg := range m {
		args = append(args, arg)
	}
	return
}

func define(name, messageFormat string, publicAttributes ...string) Definition {
	ns := pkg(3)
	fullName := fmt.Sprintf("%s:%s", ns, name)

	if Definitions[fullName] != nil {
		panic(fmt.Errorf("Error %s already defined", fullName))
	}

	def := Definition{
		namespace:              ns,
		name:                   name,
		messageFormat:          messageFormat,
		messageFormatArguments: messageFormatArguments(messageFormat),
		publicAttributes:       publicAttributes,
		code:                   int32(codes.Unknown),
	}

	// All message format arguments must be public:
nextArg:
	for _, arg := range def.messageFormatArguments {
		for _, attr := range def.publicAttributes {
			if arg == attr {
				continue nextArg
			}
		}
		def.publicAttributes = append(def.publicAttributes, arg)
	}

	Definitions[fullName] = &def
	return def
}

// Definitions of registered errors.
// Errors that are defined in init() funcs will be collected for translation.
var Definitions = make(map[string]*Definition)

// Canceled - not used for now; should be created by canceling context.

// Define defines a registered error of type Unknown.
func Define(name, messageFormat string, publicAttributes ...string) Definition {
	return define(name, messageFormat, publicAttributes...)
}

// DefineInvalidArgument defines a registered error of type InvalidArgument.
func DefineInvalidArgument(name, messageFormat string, publicAttributes ...string) Definition {
	def := define(name, messageFormat, publicAttributes...)
	def.code = int32(codes.InvalidArgument)
	return def
}

// DeadlineExceeded - not used for now; should be created by expiring context.

// DefineNotFound defines a registered error of type NotFound.
func DefineNotFound(name, messageFormat string, publicAttributes ...string) Definition {
	def := define(name, messageFormat, publicAttributes...)
	def.code = int32(codes.NotFound)
	return def
}

// DefineAlreadyExists defines a registered error of type AlreadyExists.
func DefineAlreadyExists(name, messageFormat string, publicAttributes ...string) Definition {
	def := define(name, messageFormat, publicAttributes...)
	def.code = int32(codes.AlreadyExists)
	return def
}

// DefinePermissionDenied defines a registered error of type PermissionDenied.
// It should be used when a client attempts to perform an authorized action using incorrect credentials or credentials with insufficient rights.
// If the client attempts to perform the action without providing any form of authentication, Unauthenticated should be used instead.
func DefinePermissionDenied(name, messageFormat string, publicAttributes ...string) Definition {
	def := define(name, messageFormat, publicAttributes...)
	def.code = int32(codes.PermissionDenied)
	return def
}

// DefineResourceExhausted defines a registered error of type ResourceExhausted.
func DefineResourceExhausted(name, messageFormat string, publicAttributes ...string) Definition {
	def := define(name, messageFormat, publicAttributes...)
	def.code = int32(codes.ResourceExhausted)
	return def
}

// DefineFailedPrecondition defines a registered error of type FailedPrecondition.
// Use Unavailable if the client can retry just the failing call.
// Use Aborted if the client should retry at a higher-level.
func DefineFailedPrecondition(name, messageFormat string, publicAttributes ...string) Definition {
	def := define(name, messageFormat, publicAttributes...)
	def.code = int32(codes.FailedPrecondition)
	return def
}

// DefineAborted defines a registered error of type Aborted.
func DefineAborted(name, messageFormat string, publicAttributes ...string) Definition {
	def := define(name, messageFormat, publicAttributes...)
	def.code = int32(codes.Aborted)
	return def
}

// OutOfRange - not used for now

// Unimplemented - not used for now

// DefineInternal defines a registered error of type Internal.
func DefineInternal(name, messageFormat string, publicAttributes ...string) Definition {
	def := define(name, messageFormat, publicAttributes...)
	def.code = int32(codes.Internal)
	return def
}

// DefineUnavailable defines a registered error of type Unavailable.
func DefineUnavailable(name, messageFormat string, publicAttributes ...string) Definition {
	def := define(name, messageFormat, publicAttributes...)
	def.code = int32(codes.Unavailable)
	return def
}

// DefineDataLoss defines a registered error of type DataLoss.
func DefineDataLoss(name, messageFormat string, publicAttributes ...string) Definition {
	def := define(name, messageFormat, publicAttributes...)
	def.code = int32(codes.DataLoss)
	return def
}

// DefineCorruption is the same as DefineDataLoss.
func DefineCorruption(name, messageFormat string, publicAttributes ...string) Definition {
	def := define(name, messageFormat, publicAttributes...)
	def.code = int32(codes.DataLoss)
	return def
}

// DefineUnauthenticated defines a registered error of type Unauthenticated.
// It should be used when a client attempts to perform an authenticated action without providing any form of authentication.
// If the client attempts to perform the action using incorrect credentials or credentials with insufficient rights,
// PermissionDenied should be used instead.
func DefineUnauthenticated(name, messageFormat string, publicAttributes ...string) Definition {
	def := define(name, messageFormat, publicAttributes...)
	def.code = int32(codes.Unauthenticated)
	return def
}
