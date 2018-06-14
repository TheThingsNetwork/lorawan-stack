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
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type v3Compat struct {
	Error
}

func (c v3Compat) MessageFormat() string {
	if desc := Descriptor(c.Error); desc != nil {
		return desc.MessageFormat
	}
	return c.Error.Message()
}

func (c v3Compat) PublicAttributes() map[string]interface{} {
	return Safe(c.Error).Attributes()
}

func (c v3Compat) CorrelationID() string {
	return c.Error.ID()
}

// CompatStatus is a compatibility func for converting Errors to gRPC statuses.
func CompatStatus(err Error) *status.Status {
	// Find the right error type/code:
	code := codes.Unknown
	switch err.Type() {
	case InvalidArgument:
		code = codes.InvalidArgument
	case OutOfRange:
		code = codes.OutOfRange
	case NotFound:
		code = codes.NotFound
	case Conflict:
		code = codes.FailedPrecondition
	case AlreadyExists:
		code = codes.AlreadyExists
	case Unauthorized:
		code = codes.Unauthenticated
	case PermissionDenied:
		code = codes.PermissionDenied
	case Timeout:
		code = codes.DeadlineExceeded
	case NotImplemented:
		code = codes.Unimplemented
	case TemporarilyUnavailable:
		code = codes.Unavailable
	case PermanentlyUnavailable:
		code = codes.FailedPrecondition
	case Canceled:
		code = codes.Canceled
	case ResourceExhausted:
		code = codes.ResourceExhausted
	}

	// Build a gRPC status:
	s := status.New(code, err.Message())

	// Convert error details to proto if possible:
	if errors.ErrorDetailsToProto == nil {
		return s
	}
	proto := errors.ErrorDetailsToProto(v3Compat{err})
	if proto == nil {
		return s
	}

	// Set the details on the gRPC status:
	s, sErr := s.WithDetails(proto)
	if sErr != nil {
		panic(sErr)
	}

	return s
}
