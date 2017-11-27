// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpcerrors

import (
	"context"
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

// FromGRPC parses a gRPC error and returns an Error
func FromGRPC(in error) (err errors.Error) {
	switch in {
	case io.EOF:
		return errors.ErrEOF.New(nil)
	case context.Canceled:
		return errors.ErrContextCanceled.New(nil)
	case context.DeadlineExceeded:
		return errors.ErrContextDeadlineExceeded.New(nil)
	case grpc.ErrClientConnClosing:
		return ErrClientConnClosing.New(nil)
	case grpc.ErrClientConnTimeout:
		return ErrClientConnTimeout.New(nil)
	case grpc.ErrServerStopped:
		return ErrServerStopped.New(nil)
	}
	if err, ok := in.(errors.Error); ok {
		return err
	}
	if status, ok := status.FromError(in); ok {
		out := &impl{Status: status, code: errors.NoCode}
		switch {
		case status.Code() == codes.FailedPrecondition && status.Message() == ErrClientConnClosing.MessageFormat:
			return ErrClientConnClosing.New(nil)
		case status.Code() == codes.Unavailable && status.Message() == ErrConnClosing.MessageFormat:
			return ErrConnClosing.New(nil)
		case status.Code() == codes.Canceled && status.Message() == errors.ErrContextCanceled.MessageFormat:
			return errors.ErrContextCanceled.New(nil)
		}
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
	switch in {
	case io.EOF:
		in = errors.ErrEOF.New(nil)
	case context.Canceled:
		in = errors.ErrContextCanceled.New(nil)
	case context.DeadlineExceeded:
		in = errors.ErrContextDeadlineExceeded.New(nil)
	case grpc.ErrClientConnClosing:
		in = ErrClientConnClosing.New(nil)
	case grpc.ErrClientConnTimeout:
		in = ErrClientConnTimeout.New(nil)
	case grpc.ErrServerStopped:
		in = ErrServerStopped.New(nil)
	}

	e, ok := in.(errors.Error)
	if ok {
		e = errors.Safe(e)
	} else {
		e = errors.Safe(errors.From(in))
		log.WithField("in_error", in).Error("Sending unknown error over gRPC")
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
