// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpcerrors

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/transport"
)

// ErrClientConnClosing is the descriptor for the grpc.ErrClientConnClosing error.
var ErrClientConnClosing = &errors.ErrDescriptor{
	MessageFormat: grpc.ErrClientConnClosing.Error(),
	Code:          errors.Code(1),
	Namespace:     "grpc",
	Type:          errors.TemporarilyUnavailable,
}

// ErrClientConnTimeout is the descriptor for the grpc.ErrClientConnTimeout error
var ErrClientConnTimeout = &errors.ErrDescriptor{
	MessageFormat: grpc.ErrClientConnTimeout.Error(),
	Code:          errors.Code(2),
	Namespace:     "grpc",
	Type:          errors.TemporarilyUnavailable,
}

// ErrServerStopped is the descriptor for the grpc.ErrServerStopped error
var ErrServerStopped = &errors.ErrDescriptor{
	MessageFormat: grpc.ErrServerStopped.Error(),
	Code:          errors.Code(3),
	Namespace:     "grpc",
	Type:          errors.TemporarilyUnavailable,
}

// ErrConnClosing is the descriptor for the grpc.ErrConnClosing error
var ErrConnClosing = &errors.ErrDescriptor{
	MessageFormat: transport.ErrConnClosing.Desc,
	Code:          errors.Code(4),
	Namespace:     "grpc",
	Type:          errors.TemporarilyUnavailable,
}

func init() {
	ErrClientConnClosing.Register()
	ErrClientConnTimeout.Register()
	ErrServerStopped.Register()
	ErrConnClosing.Register()
}
