// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/goproto"
	"github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TypeToGRPCCode returns the corresponding http status code from an error type
func TypeToGRPCCode(t errors.Type) codes.Code {
	switch t {
	case errors.InvalidArgument:
		return codes.InvalidArgument
	case errors.OutOfRange:
		return codes.OutOfRange
	case errors.NotFound:
		return codes.NotFound
	case errors.Conflict:
	case errors.AlreadyExists:
		return codes.AlreadyExists
	case errors.Unauthorized:
		return codes.Unauthenticated
	case errors.PermissionDenied:
		return codes.PermissionDenied
	case errors.Timeout:
		return codes.DeadlineExceeded
	case errors.NotImplemented:
		return codes.Unimplemented
	case errors.TemporarilyUnavailable:
		return codes.Unavailable
	case errors.PermanentlyUnavailable:
		return codes.FailedPrecondition
	case errors.Canceled:
		return codes.Canceled
	case errors.ResourceExhausted:
		return codes.ResourceExhausted
	case errors.Internal:
	case errors.Unknown:
	}
	return codes.Unknown
}

// GRPCCodeToType converts the gRPC error code to an error type or returns the
// Unknown type if not possible.
func GRPCCodeToType(code codes.Code) errors.Type {
	switch code {
	case codes.InvalidArgument:
		return errors.InvalidArgument
	case codes.OutOfRange:
		return errors.OutOfRange
	case codes.NotFound:
		return errors.NotFound
	case codes.AlreadyExists:
		return errors.AlreadyExists
	case codes.Unauthenticated:
		return errors.Unauthorized
	case codes.PermissionDenied:
		return errors.PermissionDenied
	case codes.DeadlineExceeded:
		return errors.Timeout
	case codes.Unimplemented:
		return errors.NotImplemented
	case codes.Unavailable:
		return errors.TemporarilyUnavailable
	case codes.FailedPrecondition:
		return errors.PermanentlyUnavailable
	case codes.Canceled:
		return errors.Canceled
	case codes.ResourceExhausted:
		return errors.ResourceExhausted
	case codes.Unknown:
		return errors.Unknown
	}
	return errors.Unknown
}

// GRPCCode returns the corresponding http status code from an error
func GRPCCode(err error) codes.Code {
	e, ok := err.(errors.Error)
	if ok {
		return TypeToGRPCCode(e.Type())
	}
	return grpc.Code(err)
}

type impl struct {
	*status.Status
	attrs errors.Attributes
	code  errors.Code
}

func (i impl) Error() string {
	return i.Status.Message()
}
func (i impl) Code() errors.Code {
	return i.code
}
func (i impl) Type() errors.Type {
	return GRPCCodeToType(i.Status.Code())
}
func (i impl) Attributes() errors.Attributes {
	return i.attrs
}

// FromGRPC parses a gRPC error and returns an Error
func FromGRPC(in error) errors.Error {
	status, ok := status.FromError(in)
	if ok {
		out := &impl{Status: status, code: errors.NoCode}
		for _, details := range status.Details() {
			if details, ok := details.(*structpb.Struct); ok {
				for k, v := range goproto.MapFromProto(details) {
					switch k {
					case "ttn-error-code":
						if v, ok := v.(float64); ok {
							out.code = errors.Code(v)
						}
					case "attributes":
						if v, ok := v.(map[string]interface{}); ok {
							out.attrs = v
						}
					}
				}
			}
		}
		return errors.ToImpl(out)
	}
	return errors.ToImpl(errors.New(in.Error()))
}

// ToGRPC turns an error into a gRPC error
func ToGRPC(in error) error {
	if err, ok := in.(errors.Error); ok {
		s, dErr := status.New(TypeToGRPCCode(err.Type()), err.Error()).
			WithDetails(goproto.MapProto(map[string]interface{}{
				"ttn-error-code": uint32(err.Code()),
				"attributes":     err.Attributes(),
			}))
		if dErr != nil {
			panic(err) // probably means you're trying to send very very bad attributes
		}
		return s.Err()
	}
	return status.Errorf(codes.Unknown, in.Error())
}
