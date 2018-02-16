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

func ExampleRegisterUnaryHook_service() {
	RegisterUnaryHook("/ttn.v3.ExampleService", "add-magic", func(h grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			ctx = context.WithValue(ctx, "magic", 42)
			return h(ctx, req)
		}
	})
}

func ExampleRegisterUnaryHook_method() {
	RegisterUnaryHook("/ttn.v3.ExampleService/Get", "add-magic", func(h grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			ctx = context.WithValue(ctx, "magic", 42)
			return h(ctx, req)
		}
	})
}

func newCallbackUnaryHook(name string, callback func(string)) func(h grpc.UnaryHandler) grpc.UnaryHandler {
	return func(h grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			callback(name)
			return h(ctx, req)
		}
	}
}

func newCallbackStreamHook(name string, callback func(string)) func(h grpc.StreamHandler) grpc.StreamHandler {
	return func(h grpc.StreamHandler) grpc.StreamHandler {
		return func(srv interface{}, stream grpc.ServerStream) error {
			callback(name)
			return h(srv, stream)
		}
	}
}

func errorHook(h grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		return nil, errors.New("failed")
	}
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
	counterUnaryHook := newCallbackUnaryHook("", count)
	counterStreamHook := newCallbackStreamHook("", count)
	expected := 0

	RegisterUnaryHook("/ttn.v3.TestService", "hook1", counterUnaryHook)
	RegisterUnaryHook("/ttn.v3.TestService", "hook2", counterUnaryHook)
	RegisterUnaryHook("/ttn.v3.TestService", "hook2", counterUnaryHook) // overwrite hook2
	RegisterUnaryHook("/ttn.v3.TestService/Foo", "hook3", counterUnaryHook)
	RegisterStreamHook("/ttn.v3.TestService", "hook4", counterStreamHook)
	RegisterStreamHook("/ttn.v3.TestService/FooStream", "hook5", counterStreamHook)

	callUnary("/ttn.v3.TestService/Foo")
	expected += 3 // hook1, hook2, hook3
	a.So(actual, should.Equal, expected)

	callStream("/ttn.v3.TestService/BarStream")
	expected += 1 // hook4
	a.So(actual, should.Equal, expected)

	callStream("/ttn.v3.TestService/FooStream")
	expected += 2 // hook4, hook5
	a.So(actual, should.Equal, expected)

	callUnary("/ttn.v3.AnotherService/Baz")
	expected += 0
	a.So(actual, should.Equal, expected)

	UnregisterUnaryHook("/ttn.v3.TestService", "hook1")
	UnregisterUnaryHook("/ttn.v3.TestService", "hook2")
	UnregisterUnaryHook("/ttn.v3.TestService/Foo", "hook3")
	UnregisterStreamHook("/ttn.v3.TestService", "hook4")
	UnregisterStreamHook("/ttn.v3.TestService/FooStream", "hook5")

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

	RegisterUnaryHook("/ttn.v3.TestService/Foo", "hook1", newCallbackUnaryHook("hook1", callback))
	RegisterUnaryHook("/ttn.v3.TestService/Foo", "hook2", errorHook)
	RegisterUnaryHook("/ttn.v3.TestService/Foo", "hook3", newCallbackUnaryHook("hook3", callback))

	err := call("/ttn.v3.TestService/Foo")
	a.So(err, should.NotBeNil)
	a.So(actual, should.Resemble, []string{"hook1"})

	UnregisterUnaryHook("/ttn.v3.TestService/Foo", "hook1")
	UnregisterUnaryHook("/ttn.v3.TestService/Foo", "hook2")
	UnregisterUnaryHook("/ttn.v3.TestService/Foo", "hook3")
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

	RegisterUnaryHook("/ttn.v3.TestService/Foo", "hook3", newCallbackUnaryHook("hook3", callback))
	RegisterUnaryHook("/ttn.v3.TestService/Foo", "hook4", newCallbackUnaryHook("hook4", callback))
	RegisterUnaryHook("/ttn.v3.TestService", "hook1", newCallbackUnaryHook("hook1", callback)) // service hooks go first
	RegisterUnaryHook("/ttn.v3.TestService", "hook2", newCallbackUnaryHook("hook2", callback))

	call("/ttn.v3.TestService/Foo")
	a.So(actual, should.Resemble, []string{"hook1", "hook2", "hook3", "hook4"})

	UnregisterUnaryHook("/ttn.v3.TestService/Foo", "hook3")
	UnregisterUnaryHook("/ttn.v3.TestService/Foo", "hook4")
	UnregisterUnaryHook("/ttn.v3.TestService", "hook1")
	UnregisterUnaryHook("/ttn.v3.TestService", "hook2")
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

	RegisterUnaryHook("/ttn.v3.TestService", "producer", func(h grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			ctx = context.WithValue(ctx, "produced-value", 42)
			return h(ctx, req)
		}
	})

	RegisterUnaryHook("/ttn.v3.TestService", "consumer", func(h grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			a.So(ctx.Value("global-value"), should.Equal, 1337)
			a.So(ctx.Value("produced-value"), should.Equal, 42)
			return h(ctx, req)
		}
	})

	call("/ttn.v3.TestService/Test")

	UnregisterUnaryHook("/ttn.v3.TestService", "producer")
	UnregisterUnaryHook("/ttn.v3.TestService", "consumer")
}
