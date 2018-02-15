// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package hooks implements a gRPC middleware that executes service and method hooks.
package hooks

import (
	"context"
	"strings"
	"sync"

	"github.com/TheThingsNetwork/ttn/pkg/log"
	"google.golang.org/grpc"
)

// Hook represents a gRPC middleware hook. If a hook returns an error, next hooks will not be
// executed and the error is propagated as the result of the interceptor.
type Hook func(ctx context.Context, req interface{}) (context.Context, error)

var filteredHooks sync.Map

type hooks struct {
	mu   sync.Mutex
	list []struct {
		name string
		f    Hook
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
func RegisterHook(filter, name string, f Hook) {
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
		f    Hook
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

func executeOn(ctx context.Context, filter string, req interface{}) (context.Context, error) {
	val, ok := filteredHooks.Load(filter)
	if !ok {
		return ctx, nil
	}

	var err error
	logger := log.FromContext(ctx)

	hooks := val.(*hooks)
	hooks.mu.Lock()
	for _, hook := range hooks.list {
		ctx, err = hook.f(ctx, req)
		if err != nil {
			logger.WithError(err).WithField("name", hook.name).Debug("Hook interceptor hook failed")
			break
		}
	}
	hooks.mu.Unlock()
	return ctx, err
}

func executeHooks(ctx context.Context, req interface{}, fullMethod string) (context.Context, error) {
	var err error

	// Split the package.service part from the full method, i.e., /package.service/method.
	service := strings.SplitN(fullMethod[1:], "/", 2)[0]

	ctx, err = executeOn(ctx, "/"+service, req)
	if err != nil {
		return nil, err
	}

	ctx, err = executeOn(ctx, fullMethod, req)
	if err != nil {
		return nil, err
	}

	return ctx, nil
}

// UnaryServerInterceptor returns a new unary server interceptor that executes registered hooks.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var err error
		ctx, err = executeHooks(ctx, req, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new stream server interceptor that executes registered hooks.
func StreamServerInterceptor(ctx context.Context) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var err error
		ctx, err = executeHooks(ctx, nil, info.FullMethod)
		if err != nil {
			return err
		}

		return handler(srv, stream)
	}
}
