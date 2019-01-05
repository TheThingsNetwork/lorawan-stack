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

package rpclog

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/logging"
	"go.thethings.network/lorawan-stack/pkg/log"
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
