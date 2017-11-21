// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rpclog

import (
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging"
	"google.golang.org/grpc/codes"
)

var (
	defaultOptions = &options{
		levelFunc: DefaultCodeToLevel,
		codeFunc:  grpc_logging.DefaultErrorToCode,
	}
)

type options struct {
	levelFunc CodeToLevel
	codeFunc  grpc_logging.ErrorToCode
}

func evaluateServerOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = DefaultCodeToLevel
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

func evaluateClientOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = DefaultClientCodeToLevel
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

// Option for the grpc logging interceptors
type Option func(*options)

// CodeToLevel function defines the mapping between gRPC return codes and interceptor log level.
type CodeToLevel func(code codes.Code) log.Level

// WithLevels customizes the function for mapping gRPC return codes and interceptor log level statements.
func WithLevels(f CodeToLevel) Option {
	return func(o *options) {
		o.levelFunc = f
	}
}

// WithCodes customizes the function for mapping errors to error codes.
func WithCodes(f grpc_logging.ErrorToCode) Option {
	return func(o *options) {
		o.codeFunc = f
	}
}

// DefaultCodeToLevel is the default implementation of gRPC return codes and interceptor log level for server side.
func DefaultCodeToLevel(code codes.Code) log.Level {
	switch code {
	case codes.OK:
		return log.InfoLevel
	case codes.Canceled:
		return log.InfoLevel
	case codes.Unknown:
		return log.WarnLevel
	case codes.InvalidArgument:
		return log.InfoLevel
	case codes.DeadlineExceeded:
		return log.WarnLevel
	case codes.NotFound:
		return log.InfoLevel
	case codes.AlreadyExists:
		return log.InfoLevel
	case codes.PermissionDenied:
		return log.WarnLevel
	case codes.Unauthenticated:
		return log.InfoLevel
	case codes.ResourceExhausted:
		return log.WarnLevel
	case codes.FailedPrecondition:
		return log.WarnLevel
	case codes.Aborted:
		return log.WarnLevel
	case codes.OutOfRange:
		return log.InfoLevel
	case codes.Unimplemented:
		return log.ErrorLevel
	case codes.Internal:
		return log.ErrorLevel
	case codes.Unavailable:
		return log.WarnLevel
	case codes.DataLoss:
		return log.ErrorLevel
	default:
		return log.ErrorLevel
	}
}

// DefaultClientCodeToLevel is the default implementation of gRPC return codes to log levels for client side.
func DefaultClientCodeToLevel(code codes.Code) log.Level {
	switch code {
	case codes.OK:
		return log.DebugLevel
	case codes.Canceled:
		return log.DebugLevel
	case codes.Unknown:
		return log.DebugLevel
	case codes.InvalidArgument:
		return log.DebugLevel
	case codes.DeadlineExceeded:
		return log.DebugLevel
	case codes.NotFound:
		return log.DebugLevel
	case codes.AlreadyExists:
		return log.DebugLevel
	case codes.PermissionDenied:
		return log.WarnLevel
	case codes.Unauthenticated:
		return log.WarnLevel
	case codes.ResourceExhausted:
		return log.WarnLevel
	case codes.FailedPrecondition:
		return log.WarnLevel
	case codes.Aborted:
		return log.WarnLevel
	case codes.OutOfRange:
		return log.WarnLevel
	case codes.Unimplemented:
		return log.ErrorLevel // Probably client/server version mismatch
	case codes.Internal:
		return log.WarnLevel
	case codes.Unavailable:
		return log.WarnLevel
	case codes.DataLoss:
		return log.WarnLevel
	default:
		return log.WarnLevel
	}
}
