// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package hooks_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/hooks"
	. "github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"google.golang.org/grpc"
)

func ExampleRegisterHook_service() {
	RegisterHook("/ttn.v3.ExampleService", "add-magic", func(ctx context.Context, req interface{}) (context.Context, error) {
		ctx = context.WithValue(ctx, "magic", 42)
		return ctx, nil
	})
}

func ExampleRegisterHook_method() {
	RegisterHook("/ttn.v3.ExampleService/Get", "add-magic", func(ctx context.Context, req interface{}) (context.Context, error) {
		ctx = context.WithValue(ctx, "magic", 42)
		return ctx, nil
	})
}

func noopUnaryHandler(_ context.Context, _ interface{}) (interface{}, error) {
	return 42, nil
}

func noopStreamHandler(_ interface{}, _ grpc.ServerStream) error {
	return nil
}

func TestRegistration(t *testing.T) {
	a := New(t)

	calls := 0
	hook := func(ctx context.Context, _ interface{}) (context.Context, error) {
		calls++
		return ctx, nil
	}

	ctx := context.Background()
	unaryInterceptor := UnaryServerInterceptor()
	callUnary := func(fullMethod string) {
		unaryInterceptor(ctx, "test", &grpc.UnaryServerInfo{
			FullMethod: fullMethod,
		}, noopUnaryHandler)
	}
	streamInterceptor := StreamServerInterceptor(ctx)
	callStream := func(fullMethod string) {
		streamInterceptor(nil, nil, &grpc.StreamServerInfo{
			FullMethod: fullMethod,
		}, noopStreamHandler)
	}

	expected := 0

	RegisterHook("/ttn.v3.TestService", "hook1", hook)
	RegisterHook("/ttn.v3.TestService", "hook2", hook)
	RegisterHook("/ttn.v3.TestService", "hook2", hook) // overwrite hook2
	RegisterHook("/ttn.v3.TestService/Foo", "hook3", hook)

	callUnary("/ttn.v3.TestService/Foo")
	expected += 3
	a.So(calls, should.Equal, expected)

	callStream("/ttn.v3.TestService/Bar")
	expected += 2
	a.So(calls, should.Equal, expected)

	callUnary("/ttn.v3.AnotherService/Baz")
	expected += 0
	a.So(calls, should.Equal, expected)

	UnregisterHook("/ttn.v3.TestService", "hook1")
	UnregisterHook("/ttn.v3.TestService", "hook2")
	UnregisterHook("/ttn.v3.TestService/Foo", "hook3")

	callUnary("/ttn.v3.TestService/Foo")
	expected += 0
	a.So(calls, should.Equal, expected)
}

func TestErrorHook(t *testing.T) {
	a := New(t)

	ctx := context.Background()
	interceptor := UnaryServerInterceptor()
	call := func(fullMethod string) error {
		_, err := interceptor(ctx, "test", &grpc.UnaryServerInfo{
			FullMethod: fullMethod,
		}, noopUnaryHandler)
		return err
	}

	var calls []int
	newHook := func(order int, err error) Hook {
		return func(ctx context.Context, _ interface{}) (context.Context, error) {
			if err != nil {
				return nil, err
			}
			calls = append(calls, order)
			return ctx, nil
		}
	}

	RegisterHook("/ttn.v3.TestService/Foo", "1", newHook(1, nil))
	RegisterHook("/ttn.v3.TestService/Foo", "2", newHook(2, errors.New("failed")))
	RegisterHook("/ttn.v3.TestService/Foo", "3", newHook(3, nil))

	err := call("/ttn.v3.TestService/Foo")
	a.So(err, should.NotBeNil)
	a.So(calls, should.Resemble, []int{1})
}

func TestOrder(t *testing.T) {
	a := New(t)

	ctx := context.Background()
	interceptor := UnaryServerInterceptor()
	call := func(fullMethod string) error {
		_, err := interceptor(ctx, "test", &grpc.UnaryServerInfo{
			FullMethod: fullMethod,
		}, noopUnaryHandler)
		return err
	}

	var calls []int
	newHook := func(order int) Hook {
		return func(ctx context.Context, _ interface{}) (context.Context, error) {
			calls = append(calls, order)
			return ctx, nil
		}
	}

	RegisterHook("/ttn.v3.TestService/Foo", "2", newHook(2))
	RegisterHook("/ttn.v3.TestService/Foo", "3", newHook(3))
	RegisterHook("/ttn.v3.TestService", "0", newHook(0)) // service hooks go first
	RegisterHook("/ttn.v3.TestService", "1", newHook(1))

	call("/ttn.v3.TestService/Foo")
	a.So(calls, should.Resemble, []int{0, 1, 2, 3})
}
