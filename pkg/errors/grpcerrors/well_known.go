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

package grpcerrors

import (
	"go.thethings.network/lorawan-stack/pkg/errors"
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
