// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package hooks_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/hooks"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func newCallbackHook(name string, callback func(string)) func(HookFunc) HookFunc {
	return func(h HookFunc) HookFunc {
		return func(ctx context.Context, req interface{}) (context.Context, error) {
			callback(name)
			return h(ctx, req)
		}
	}
}

func errorHook(h HookFunc) HookFunc {
	return func(ctx context.Context, _ interface{}) (context.Context, error) {
		return nil, errors.New("failed")
	}
}

func ExampleRegisterHook_service() {
	RegisterHook("/ttn.v3.ExampleService", "add-magic", func(h HookFunc) HookFunc {
		return func(ctx context.Context, req interface{}) (context.Context, error) {
			ctx = context.WithValue(ctx, "magic", 42)
			return h(ctx, req)
		}
	})
}

func ExampleRegisterHook_method() {
	RegisterHook("/ttn.v3.ExampleService/Get", "add-magic", func(h HookFunc) HookFunc {
		return func(ctx context.Context, req interface{}) (context.Context, error) {
			ctx = context.WithValue(ctx, "magic", 42)
			return h(ctx, req)
		}
	})
}

func noopUnaryHandler(_ context.Context, _ interface{}) (interface{}, error) {
	return 42, nil
}

func noopStreamHandler(_ interface{}, _ grpc.ServerStream) error {
	return nil
}

type mockStream struct {
	ctx context.Context
}

func (s *mockStream) Context() context.Context     { return s.ctx }
func (s *mockStream) SendMsg(m interface{}) error  { return nil }
func (s *mockStream) RecvMsg(m interface{}) error  { return nil }
func (s *mockStream) SetHeader(metadata.MD) error  { return nil }
func (s *mockStream) SendHeader(metadata.MD) error { return nil }
func (s *mockStream) SetTrailer(metadata.MD)       {}

func TestRegistration(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()
	unaryInterceptor := UnaryServerInterceptor()
	callUnary := func(fullMethod string) {
		unaryInterceptor(ctx, "test", &grpc.UnaryServerInfo{
			FullMethod: fullMethod,
		}, noopUnaryHandler)
	}
	streamInterceptor := StreamServerInterceptor()
	callStream := func(fullMethod string) {
		streamInterceptor(nil, &mockStream{ctx}, &grpc.StreamServerInfo{
			FullMethod: fullMethod,
		}, noopStreamHandler)
	}

	actual := 0
	count := func(_ string) {
		actual++
	}
	counterHook := newCallbackHook("", count)
	expected := 0

	RegisterHook("/ttn.v3.TestService", "hook1", counterHook)
	RegisterHook("/ttn.v3.TestService", "hook2", counterHook)
	RegisterHook("/ttn.v3.TestService", "hook2", counterHook) // overwrite hook2
	RegisterHook("/ttn.v3.TestService/Foo", "hook3", counterHook)
	RegisterHook("/ttn.v3.TestService/FooStream", "hook4", counterHook)

	callUnary("/ttn.v3.TestService/Foo")
	expected += 3 // hook1, hook2, hook3
	a.So(actual, should.Equal, expected)

	callStream("/ttn.v3.TestService/BarStream")
	expected += 2 // hook1, hook2
	a.So(actual, should.Equal, expected)

	callStream("/ttn.v3.TestService/FooStream")
	expected += 3 // hook1, hook2, hook4
	a.So(actual, should.Equal, expected)

	callUnary("/ttn.v3.AnotherService/Baz")
	expected += 0
	a.So(actual, should.Equal, expected)

	UnregisterHook("/ttn.v3.TestService", "hook1")
	UnregisterHook("/ttn.v3.TestService", "hook2")
	UnregisterHook("/ttn.v3.TestService/Foo", "hook3")

	callUnary("/ttn.v3.TestService/Foo")
	expected += 0
	a.So(actual, should.Equal, expected)
}

func TestErrorHook(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()
	interceptor := UnaryServerInterceptor()
	call := func(fullMethod string) error {
		_, err := interceptor(ctx, "test", &grpc.UnaryServerInfo{
			FullMethod: fullMethod,
		}, noopUnaryHandler)
		return err
	}

	var actual []string
	callback := func(id string) {
		actual = append(actual, id)
	}

	RegisterHook("/ttn.v3.TestService/Foo", "hook1", newCallbackHook("hook1", callback))
	RegisterHook("/ttn.v3.TestService/Foo", "hook2", errorHook)
	RegisterHook("/ttn.v3.TestService/Foo", "hook3", newCallbackHook("hook3", callback))

	err := call("/ttn.v3.TestService/Foo")
	a.So(err, should.NotBeNil)
	a.So(actual, should.Resemble, []string{"hook1"})

	UnregisterHook("/ttn.v3.TestService/Foo", "hook1")
	UnregisterHook("/ttn.v3.TestService/Foo", "hook2")
	UnregisterHook("/ttn.v3.TestService/Foo", "hook3")
}

func TestOrder(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()
	interceptor := UnaryServerInterceptor()
	call := func(fullMethod string) error {
		_, err := interceptor(ctx, "test", &grpc.UnaryServerInfo{
			FullMethod: fullMethod,
		}, noopUnaryHandler)
		return err
	}

	var actual []string
	callback := func(id string) {
		actual = append(actual, id)
	}

	RegisterHook("/ttn.v3.TestService/Foo", "hook3", newCallbackHook("hook3", callback))
	RegisterHook("/ttn.v3.TestService/Foo", "hook4", newCallbackHook("hook4", callback))
	RegisterHook("/ttn.v3.TestService", "hook1", newCallbackHook("hook1", callback)) // service hooks go first
	RegisterHook("/ttn.v3.TestService", "hook2", newCallbackHook("hook2", callback))

	call("/ttn.v3.TestService/Foo")
	a.So(actual, should.Resemble, []string{"hook1", "hook2", "hook3", "hook4"})

	UnregisterHook("/ttn.v3.TestService/Foo", "hook3")
	UnregisterHook("/ttn.v3.TestService/Foo", "hook4")
	UnregisterHook("/ttn.v3.TestService", "hook1")
	UnregisterHook("/ttn.v3.TestService", "hook2")
}

func TestHookContext(t *testing.T) {
	a := assertions.New(t)

	ctx := context.WithValue(context.Background(), "global-value", 1337)
	interceptor := UnaryServerInterceptor()
	call := func(fullMethod string) error {
		_, err := interceptor(ctx, "test", &grpc.UnaryServerInfo{
			FullMethod: fullMethod,
		}, noopUnaryHandler)
		return err
	}

	RegisterHook("/ttn.v3.TestService", "producer", func(h HookFunc) HookFunc {
		return func(ctx context.Context, req interface{}) (context.Context, error) {
			ctx = context.WithValue(ctx, "produced-value", 42)
			return h(ctx, req)
		}
	})

	RegisterHook("/ttn.v3.TestService", "consumer", func(h HookFunc) HookFunc {
		return func(ctx context.Context, req interface{}) (context.Context, error) {
			a.So(ctx.Value("global-value"), should.Equal, 1337)
			a.So(ctx.Value("produced-value"), should.Equal, 42)
			return h(ctx, req)
		}
	})

	call("/ttn.v3.TestService/Test")

	UnregisterHook("/ttn.v3.TestService", "producer")
	UnregisterHook("/ttn.v3.TestService", "consumer")
}
