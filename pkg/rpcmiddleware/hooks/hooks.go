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

var unaryHooks sync.Map
var streamHooks sync.Map

type hooks struct {
	mu   sync.Mutex
	list []struct {
		name string
		f    interface{}
	}
}

func registerHook(hooksMap *sync.Map, filter, name string, f interface{}) {
	val, _ := hooksMap.LoadOrStore(filter, &hooks{})
	hooks := val.(*hooks)
	hooks.mu.Lock()
	defer hooks.mu.Unlock()
	for i := 0; i < len(hooks.list); i++ {
		if hooks.list[i].name == name {
			hooks.list[i].f = f
			return
		}
	}
	hooks.list = append(hooks.list, struct {
		name string
		f    interface{}
	}{name, f})
}

func unregisterHook(hooksMap *sync.Map, filter, name string) bool {
	val, ok := hooksMap.Load(filter)
	if !ok {
		return false
	}
	hooks := val.(*hooks)
	hooks.mu.Lock()
	defer hooks.mu.Unlock()
	for i := 0; i < len(hooks.list); i++ {
		if hooks.list[i].name == name {
			hooks.list = append(hooks.list[:i], hooks.list[i+1:]...)
			return true
		}
	}
	return false
}

// RegisterUnaryHook registers a new hook with the specified filter and name.
func RegisterUnaryHook(filter, name string, f UnaryHandlerMiddleware) {
	registerHook(&unaryHooks, filter, name, f)
}

// UnregisterUnaryHook unregisters the unary hook with the specified filter and name and returns
// true on success.
func UnregisterUnaryHook(filter, name string) bool {
	return unregisterHook(&unaryHooks, filter, name)
}

// RegisterStreamHook registers a new hook with the specified filter and name.
func RegisterStreamHook(filter, name string, f StreamHandlerMiddleware) {
	registerHook(&streamHooks, filter, name, f)
}

// UnregisterStreamHook unregisters the Stream hook with the specified filter and name and returns
// true on success.
func UnregisterStreamHook(filter, name string) bool {
	return unregisterHook(&streamHooks, filter, name)
}

// createFilters splits the package.service part from the full method, i.e.,
// /package.service/method and returns the service and method filter.
func createFilters(fullMethod string) []string {
	service := strings.SplitN(fullMethod[1:], "/", 2)[0]
	return []string{"/" + service, fullMethod}
}

// UnaryServerInterceptor returns a new unary server interceptor that executes registered hooks.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
		var middleware []UnaryHandlerMiddleware
		for _, filter := range createFilters(info.FullMethod) {
			val, ok := unaryHooks.Load(filter)
			if !ok {
				continue
			}
			hooks := val.(*hooks)
			hooks.mu.Lock()
			for _, hook := range hooks.list {
				middleware = append(middleware, hook.f.(UnaryHandlerMiddleware))
			}
			hooks.mu.Unlock()
		}
		for i := len(middleware) - 1; i >= 0; i-- {
			handler = middleware[i](handler)
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new stream server interceptor that executes registered hooks.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		var middleware []StreamHandlerMiddleware
		for _, filter := range createFilters(info.FullMethod) {
			val, ok := streamHooks.Load(filter)
			if !ok {
				continue
			}
			hooks := val.(*hooks)
			hooks.mu.Lock()
			for _, hook := range hooks.list {
				middleware = append(middleware, hook.f.(StreamHandlerMiddleware))
			}
			hooks.mu.Unlock()
		}
		for i := len(middleware) - 1; i >= 0; i-- {
			handler = middleware[i](handler)
		}
		return handler(srv, stream)
	}
}
