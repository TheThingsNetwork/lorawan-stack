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
	"context"

	"google.golang.org/grpc/codes"
)

type coder interface {
	Code() uint32
}

// Code of the error.
// If the code is invalid or unknown, this tries to get the code from the cause of this error.
// This code is consistent with google.golang.org/genproto/googleapis/rpc/code and google.golang.org/grpc/codes.
func (e Error) Code() uint32 {
	if e.code != 0 && e.code != uint32(codes.Unknown) {
		return e.code
	}
	if e.cause != nil {
		return Code(e.cause)
	}
	return uint32(codes.Unknown)
}

func code(err error) uint32 {
	if c, ok := err.(coder); ok {
		return c.Code()
	}
	if ttnErr, ok := From(err); ok {
		return ttnErr.Code()
	}
	return 0
}

// Code gets the code of an error.
// If the error doesn't implement codes, Unknown is returned.
func Code(err error) uint32 {
	if code := code(err); code != 0 {
		return code
	}
	return uint32(codes.Unknown)
}

// HasCode returns whether the given error has the given error code.
// If the error doesn't implement codes, it doesn't have any code.
func HasCode(err error, c uint32) bool {
	return code(err) == c
}

// IsCanceled returns whether the givenerror is context.Canceled or of type Canceled.
func IsCanceled(err error) bool {
	return err == context.Canceled || HasCode(err, uint32(codes.Canceled))
}

// IsUnknown returns whether the given error is of type Unknown.
func IsUnknown(err error) bool { return HasCode(err, uint32(codes.Unknown)) }

// IsInvalidArgument returns whether the given error is of type InvalidArgument.
func IsInvalidArgument(err error) bool { return HasCode(err, uint32(codes.InvalidArgument)) }

// IsDeadlineExceeded returns whether the givenerror is context.DeadlineExceeded or of type DeadlineExceeded.
func IsDeadlineExceeded(err error) bool {
	return err == context.DeadlineExceeded || HasCode(err, uint32(codes.DeadlineExceeded))
}

// IsNotFound returns whether the given error is of type NotFound.
func IsNotFound(err error) bool { return HasCode(err, uint32(codes.NotFound)) }

// IsAlreadyExists returns whether the given error is of type AlreadyExists.
func IsAlreadyExists(err error) bool { return HasCode(err, uint32(codes.AlreadyExists)) }

// IsPermissionDenied returns whether the given error is of type PermissionDenied.
func IsPermissionDenied(err error) bool { return HasCode(err, uint32(codes.PermissionDenied)) }

// IsResourceExhausted returns whether the given error is of type ResourceExhausted.
func IsResourceExhausted(err error) bool { return HasCode(err, uint32(codes.ResourceExhausted)) }

// IsFailedPrecondition returns whether the given error is of type FailedPrecondition.
func IsFailedPrecondition(err error) bool { return HasCode(err, uint32(codes.FailedPrecondition)) }

// IsAborted returns whether the given error is of type Aborted.
func IsAborted(err error) bool { return HasCode(err, uint32(codes.Aborted)) }

// IsInternal returns whether the given error is of type Internal.
func IsInternal(err error) bool { return HasCode(err, uint32(codes.Internal)) }

// IsUnavailable returns whether the given error is of type Unavailable.
func IsUnavailable(err error) bool { return HasCode(err, uint32(codes.Unavailable)) }

// IsDataLoss returns whether the given error is of type DataLoss.
func IsDataLoss(err error) bool { return HasCode(err, uint32(codes.DataLoss)) }

// IsUnauthenticated returns whether the given error is of type Unauthenticated.
func IsUnauthenticated(err error) bool { return HasCode(err, uint32(codes.Unauthenticated)) }
