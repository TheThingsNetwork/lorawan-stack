// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpcerrors

import (
	"context"
	"fmt"
	"io"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/goproto"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Keys under which error metadata is stored
const (
	CodeKey      = "ttn-error-code"
	AttributeKey = "attributes"
	NamespaceKey = "namespace"
	IDKey        = "ttn-error-id"
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
	case errors.External:
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
	attrs     errors.Attributes
	code      errors.Code
	namespace string
	id        string
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
func (i impl) Namespace() string {
	return i.namespace
}
func (i impl) ID() string {
	return i.id
}

func fromWellKnown(in error) (errors.Error, bool) {
	switch in {
	case io.EOF:
		return errors.ErrEOF.New(nil), true
	case context.Canceled:
		return errors.ErrContextCanceled.New(nil), true
	case context.DeadlineExceeded:
		return errors.ErrContextDeadlineExceeded.New(nil), true
	case grpc.ErrClientConnClosing:
		return ErrClientConnClosing.New(nil), true
	case grpc.ErrClientConnTimeout:
		return ErrClientConnTimeout.New(nil), true
	case grpc.ErrServerStopped:
		return ErrServerStopped.New(nil), true
	}
	return nil, false
}

func fromStatus(status *status.Status) (errors.Error, bool) {
	switch {
	case status.Code() == codes.FailedPrecondition && status.Message() == ErrClientConnClosing.MessageFormat:
		return ErrClientConnClosing.New(nil), true
	case status.Code() == codes.Unavailable && status.Message() == ErrConnClosing.MessageFormat:
		return ErrConnClosing.New(nil), true
	case status.Code() == codes.Canceled && status.Message() == errors.ErrContextCanceled.MessageFormat:
		return errors.ErrContextCanceled.New(nil), true
	}
	return nil, false
}

// FromGRPC parses a gRPC error and returns an Error
func FromGRPC(in error) (err errors.Error) {
	if in == nil {
		return nil
	}
	if err, ok := in.(errors.Error); ok {
		return err
	}
	if wellKnown, ok := fromWellKnown(in); ok {
		return wellKnown
	}
	if status, ok := status.FromError(in); ok {
		if status, ok := fromStatus(status); ok {
			return status
		}
		out := &impl{Status: status, code: errors.NoCode}
		for _, details := range status.Details() {
			if details, ok := details.(*structpb.Struct); ok {
				m, err := goproto.Map(details)
				if err != nil {
					log.WithError(err).WithField("in_error", err).Errorf("Could not decode gRPC error")
					continue
				}
				for k, v := range m {
					switch k {
					case CodeKey:
						if v, ok := v.(float64); ok {
							out.code = errors.Code(v)
						}
					case AttributeKey:
						if v, ok := v.(map[string]interface{}); ok {
							out.attrs = v
						}
					case NamespaceKey:
						if v, ok := v.(string); ok {
							out.namespace = v
						}
					case IDKey:
						if v, ok := v.(string); ok {
							out.id = v
						}
					}
				}
			}
		}

		return errors.ToImpl(out)
	}

	return errors.From(in)
}

// ToGRPC turns an error into a gRPC error
func ToGRPC(in error) (err error) {
	if in == nil {
		return nil
	}
	if _, ok := status.FromError(in); ok {
		return in
	}
	if wellKnown, ok := fromWellKnown(in); ok {
		return wellKnown
	}
	e, ok := in.(errors.Error)
	if ok {
		e = errors.Safe(e)
	} else {
		e = errors.Safe(errors.From(in))
		log.WithFields(log.Fields(
			"in_error", in,
			"in_error_type", fmt.Sprintf("%T", in),
		)).Warn("An unknown error type was sent over gRPC, please use the TTN errors package instead")
	}

	details, err := goproto.Struct(map[string]interface{}{
		CodeKey:      uint32(e.Code()),
		AttributeKey: e.Attributes(),
		NamespaceKey: e.Namespace(),
		IDKey:        e.ID(),
	})

	if err != nil {
		panic(err) // you're trying to encode something you should not be encoding
	}

	status, err := status.New(TypeToGRPCCode(e.Type()), e.Message()).WithDetails(details)
	if err != nil {
		panic(err) // probably means you're trying to send very very bad attributes
	}

	return status.Err()
}
