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
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"google.golang.org/grpc"
)

const NamespaceHook = "namespace"

// UnaryNamespaceHook adds the component namespace to the context of the unary call.
func UnaryNamespaceHook(namespace string) hooks.UnaryHandlerMiddleware {
	return func(h grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			ctx = log.NewContextWithField(ctx, "namespace", namespace)
			return h(ctx, req)
		}
	}
}

// StreamNamespaceHook adds the component namespace to the context of the stream.
func StreamNamespaceHook(namespace string) hooks.StreamHandlerMiddleware {
	return func(h grpc.StreamHandler) grpc.StreamHandler {
		return func(srv interface{}, stream grpc.ServerStream) error {
			wrapped := grpc_middleware.WrapServerStream(stream)
			wrapped.WrappedContext = log.NewContextWithField(stream.Context(), "namespace", namespace)
			return h(srv, wrapped)
		}
	}
}
