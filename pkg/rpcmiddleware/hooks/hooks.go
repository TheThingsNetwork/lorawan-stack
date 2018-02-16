// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package hooks implements a gRPC middleware that executes service and method hooks.
package hooks

import (
	"context"
	"strings"
	"sync"

	"google.golang.org/grpc"
)

// HookFunc represents a gRPC middleware hook.
type HookFunc func(ctx context.Context, req interface{}) (context.Context, error)

// MiddlewareFunc wraps HookFunc.
type MiddlewareFunc func(HookFunc) HookFunc

var filteredHooks sync.Map

type hooks struct {
	mu   sync.Mutex
	list []struct {
		name string
		f    MiddlewareFunc
	}
}

// RegisterHook registers a new hook with the specified filter and name.
//
// Register a hook for all methods in a service with filter /package.service or a specific method
// with filter /package.service/method. If there is already a hook registered with the specified
// filter and name, the registered hook gets overwritten.
//
// Hooks are executed in order of registration. Service hooks are executed before service method
// hooks.
func RegisterHook(filter, name string, f MiddlewareFunc) {
	val, _ := filteredHooks.LoadOrStore(filter, &hooks{})
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
		f    MiddlewareFunc
	}{name, f})
}

// UnregisterHook unregisters the hook with the specified filter and name and returns true on
// success.
func UnregisterHook(filter, name string) bool {
	val, ok := filteredHooks.Load(filter)
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

func buildChain(f HookFunc, m ...MiddlewareFunc) HookFunc {
	if len(m) == 0 {
		return f
	}
	return m[0](buildChain(f, m[1:]...))
}

func filterMiddleware(fullMethod string) []MiddlewareFunc {
	// Split the package.service part from the full method, i.e., /package.service/method.
	service := strings.SplitN(fullMethod[1:], "/", 2)[0]
	// Place service filter before method filter to preserve order.
	filters := []string{"/" + service, fullMethod}

	var middleware []MiddlewareFunc
	for _, filter := range filters {
		val, ok := filteredHooks.Load(filter)
		if !ok {
			continue
		}
		hooks := val.(*hooks)
		hooks.mu.Lock()
		for _, hook := range hooks.list {
			middleware = append(middleware, hook.f)
		}
		hooks.mu.Unlock()
	}
	return middleware
}

// UnaryServerInterceptor returns a new unary server interceptor that executes registered hooks.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
		inner := func(ctx context.Context, req interface{}) (context.Context, error) {
			res, err = handler(ctx, req)
			if err != nil {
				return nil, err
			}
			return ctx, nil
		}

		chain := buildChain(inner, filterMiddleware(info.FullMethod)...)
		_, err = chain(ctx, req)
		return
	}
}

// StreamServerInterceptor returns a new stream server interceptor that executes registered hooks.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		inner := func(ctx context.Context, _ interface{}) (context.Context, error) {
			err = handler(srv, stream)
			if err != nil {
				return nil, err
			}
			return ctx, nil
		}

		chain := buildChain(inner, filterMiddleware(info.FullMethod)...)
		_, err = chain(stream.Context(), nil)
		return
	}
}
