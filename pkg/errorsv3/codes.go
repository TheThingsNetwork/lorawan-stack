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
	"context"
	"net/http"

	"google.golang.org/grpc/codes"
)

// Code of the error.
// If the code is invalid or unknown, this tries to get the code from the cause of this error.
// This code is consistent with google.golang.org/genproto/googleapis/rpc/code and google.golang.org/grpc/codes.
func (e Error) Code() int32 {
	if e.code != 0 && e.code != int32(codes.Unknown) {
		return e.code
	}
	if e.cause != nil {
		if c, ok := e.cause.(interface{ Code() int32 }); ok {
			return c.Code()
		}
	}
	return int32(codes.Unknown)
}

func code(err error) int32 {
	if c, ok := err.(interface{ Code() int32 }); ok {
		return c.Code()
	}
	if err, ok := From(err); ok {
		return err.Code()
	}
	return 0
}

// Code gets the code of an error.
// If the error doesn't implement codes, Unknown is returned.
func Code(err error) int32 {
	if code := code(err); code != 0 {
		return code
	}
	return int32(codes.Unknown)
}

// HasCode returns whether the given error has the given error code.
// If the error doesn't implement codes, it doesn't have any code.
func HasCode(err error, c int32) bool {
	return code(err) == c
}

// IsCanceled returns whether the givenerror is context.Canceled or of type Canceled.
func IsCanceled(err error) bool {
	return err == context.Canceled || HasCode(err, int32(codes.Canceled))
}

// IsUnknown returns whether the given error is of type Unknown.
func IsUnknown(err error) bool { return HasCode(err, int32(codes.Unknown)) }

// IsInvalidArgument returns whether the given error is of type InvalidArgument.
func IsInvalidArgument(err error) bool { return HasCode(err, int32(codes.InvalidArgument)) }

// IsDeadlineExceeded returns whether the givenerror is context.DeadlineExceeded or of type DeadlineExceeded.
func IsDeadlineExceeded(err error) bool {
	return err == context.DeadlineExceeded || HasCode(err, int32(codes.DeadlineExceeded))
}

// IsNotFound returns whether the given error is of type NotFound.
func IsNotFound(err error) bool { return HasCode(err, int32(codes.NotFound)) }

// IsAlreadyExists returns whether the given error is of type AlreadyExists.
func IsAlreadyExists(err error) bool { return HasCode(err, int32(codes.AlreadyExists)) }

// IsPermissionDenied returns whether the given error is of type PermissionDenied.
func IsPermissionDenied(err error) bool { return HasCode(err, int32(codes.PermissionDenied)) }

// IsResourceExhausted returns whether the given error is of type ResourceExhausted.
func IsResourceExhausted(err error) bool { return HasCode(err, int32(codes.ResourceExhausted)) }

// IsFailedPrecondition returns whether the given error is of type FailedPrecondition.
func IsFailedPrecondition(err error) bool { return HasCode(err, int32(codes.FailedPrecondition)) }

// IsAborted returns whether the given error is of type Aborted.
func IsAborted(err error) bool { return HasCode(err, int32(codes.Aborted)) }

// IsInternal returns whether the given error is of type Internal.
func IsInternal(err error) bool { return HasCode(err, int32(codes.Internal)) }

// IsUnavailable returns whether the given error is of type Unavailable.
func IsUnavailable(err error) bool { return HasCode(err, int32(codes.Unavailable)) }

// IsDataLoss returns whether the given error is of type DataLoss.
func IsDataLoss(err error) bool { return HasCode(err, int32(codes.DataLoss)) }

// IsUnauthenticated returns whether the given error is of type Unauthenticated.
func IsUnauthenticated(err error) bool { return HasCode(err, int32(codes.Unauthenticated)) }

// httpStatuscodes maps status codes to HTTP codes.
// See package google.golang.org/genproto/googleapis/rpc/code and google.golang.org/grpc/codes for details.
var httpStatuscodes = map[int32]int{
	int32(codes.OK):                 http.StatusOK,
	int32(codes.Canceled):           499, // Client Closed Request
	int32(codes.Unknown):            http.StatusInternalServerError,
	int32(codes.InvalidArgument):    http.StatusBadRequest,
	int32(codes.DeadlineExceeded):   http.StatusGatewayTimeout,
	int32(codes.NotFound):           http.StatusNotFound,
	int32(codes.AlreadyExists):      http.StatusConflict,
	int32(codes.PermissionDenied):   http.StatusForbidden,
	int32(codes.Unauthenticated):    http.StatusUnauthorized,
	int32(codes.ResourceExhausted):  http.StatusTooManyRequests,
	int32(codes.FailedPrecondition): http.StatusBadRequest,
	int32(codes.Aborted):            http.StatusConflict,
	int32(codes.OutOfRange):         http.StatusBadRequest,
	int32(codes.Unimplemented):      http.StatusNotImplemented,
	int32(codes.Internal):           http.StatusInternalServerError,
	int32(codes.Unavailable):        http.StatusServiceUnavailable,
	int32(codes.DataLoss):           http.StatusInternalServerError,
}

// HTTPStatusCode maps an error to HTTP response codes.
func HTTPStatusCode(err error) int {
	if status, ok := httpStatuscodes[Code(err)]; ok {
		return status
	}
	return http.StatusInternalServerError
}
