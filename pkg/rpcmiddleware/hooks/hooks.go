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

// Package hooks implements a gRPC middleware that executes service and method hooks.
//
// Register a hook for all methods in a service with filter /package.service or a specific method
// with filter /package.service/method. Registering a hook overwrites existing hooks with the same
// filter and name.
//
// Hooks are executed in order of registration. Service hooks are executed before service method
// hooks.
package hooks

import (
	"context"
	"strings"
	"sync"

	"google.golang.org/grpc"
)

// UnaryHandlerMiddleware wraps grpc.UnaryHandler.
type UnaryHandlerMiddleware func(grpc.UnaryHandler) grpc.UnaryHandler

// StreamHandlerMiddleware wraps grpc.StreamHandler.
type StreamHandlerMiddleware func(grpc.StreamHandler) grpc.StreamHandler

type unaryHook struct {
	name string
	f    UnaryHandlerMiddleware
}
type streamHook struct {
	name string
	f    StreamHandlerMiddleware
}

type Hooks struct {
	mu     sync.RWMutex
	unary  map[string][]*unaryHook
	stream map[string][]*streamHook
}

func (h *Hooks) RegisterUnaryHook(filter, name string, f UnaryHandlerMiddleware) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.unary == nil {
		h.unary = make(map[string][]*unaryHook)
	}
	for _, hook := range h.unary[filter] {
		if hook.name == name {
			hook.f = f
			return
		}
	}
	h.unary[filter] = append(h.unary[filter], &unaryHook{name: name, f: f})
}

func (h *Hooks) RegisterStreamHook(filter, name string, f StreamHandlerMiddleware) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.stream == nil {
		h.stream = make(map[string][]*streamHook)
	}
	for _, hook := range h.stream[filter] {
		if hook.name == name {
			hook.f = f
			return
		}
	}
	h.stream[filter] = append(h.stream[filter], &streamHook{name: name, f: f})
}

func (h *Hooks) UnregisterUnaryHook(filter, name string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, hook := range h.unary[filter] {
		if hook.name == name {
			h.unary[filter] = append(h.unary[filter][:i], h.unary[filter][i+1:]...)
			return true
		}
	}
	return false
}

func (h *Hooks) UnregisterStreamHook(filter, name string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, hook := range h.stream[filter] {
		if hook.name == name {
			h.stream[filter] = append(h.stream[filter][:i], h.stream[filter][i+1:]...)
			return true
		}
	}
	return false
}

// createFilters splits the package.service part from the full method, i.e.,
// /package.service/method and returns the service and method filter.
func createFilters(fullMethod string) []string {
	service := strings.SplitN(fullMethod[1:], "/", 2)[0]
	return []string{"/" + service, fullMethod}
}

// UnaryServerInterceptor returns a new unary server interceptor that executes registered hooks.
func (h *Hooks) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res any, err error) {
		var middleware []UnaryHandlerMiddleware
		h.mu.RLock()
		for _, filter := range createFilters(info.FullMethod) {
			for _, hook := range h.unary[filter] {
				middleware = append(middleware, hook.f)
			}
		}
		h.mu.RUnlock()
		for i := len(middleware) - 1; i >= 0; i-- {
			handler = middleware[i](handler)
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new stream server interceptor that executes registered hooks.
func (h *Hooks) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		var middleware []StreamHandlerMiddleware
		h.mu.RLock()
		for _, filter := range createFilters(info.FullMethod) {
			for _, hook := range h.stream[filter] {
				middleware = append(middleware, hook.f)
			}
		}
		h.mu.RUnlock()
		for i := len(middleware) - 1; i >= 0; i-- {
			handler = middleware[i](handler)
		}
		return handler(srv, stream)
	}
}
