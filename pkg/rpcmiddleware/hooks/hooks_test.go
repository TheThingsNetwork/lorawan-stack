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

package hooks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func newCallbackUnaryHook(name string, callback func(string)) func(h grpc.UnaryHandler) grpc.UnaryHandler {
	return func(h grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req any) (any, error) {
			callback(name)
			return h(ctx, req)
		}
	}
}

func newCallbackStreamHook(name string, callback func(string)) func(h grpc.StreamHandler) grpc.StreamHandler {
	return func(h grpc.StreamHandler) grpc.StreamHandler {
		return func(srv any, stream grpc.ServerStream) error {
			callback(name)
			return h(srv, stream)
		}
	}
}

func errorHook(h grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, _ any) (any, error) {
		return nil, errors.New("failed")
	}
}

func noopUnaryHandler(_ context.Context, _ any) (any, error) {
	return 42, nil
}

func noopStreamHandler(_ any, _ grpc.ServerStream) error {
	return nil
}

type mockStream struct {
	ctx context.Context
}

func (s *mockStream) Context() context.Context     { return s.ctx }
func (s *mockStream) SendMsg(m any) error          { return nil }
func (s *mockStream) RecvMsg(m any) error          { return nil }
func (s *mockStream) SetHeader(metadata.MD) error  { return nil }
func (s *mockStream) SendHeader(metadata.MD) error { return nil }
func (s *mockStream) SetTrailer(metadata.MD)       {}

func TestRegistration(t *testing.T) {
	a := assertions.New(t)

	h := &Hooks{}

	ctx := test.Context()
	unaryInterceptor := h.UnaryServerInterceptor()
	callUnary := func(fullMethod string) {
		unaryInterceptor(ctx, "test", &grpc.UnaryServerInfo{
			FullMethod: fullMethod,
		}, noopUnaryHandler)
	}
	streamInterceptor := h.StreamServerInterceptor()
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

	h.RegisterUnaryHook("/ttn.lorawan.v3.TestService", "hook1", counterUnaryHook)
	h.RegisterUnaryHook("/ttn.lorawan.v3.TestService", "hook2", counterUnaryHook)
	h.RegisterUnaryHook("/ttn.lorawan.v3.TestService", "hook2", counterUnaryHook) // overwrite hook2
	h.RegisterUnaryHook("/ttn.lorawan.v3.TestService/Foo", "hook3", counterUnaryHook)
	h.RegisterStreamHook("/ttn.lorawan.v3.TestService", "hook4", counterStreamHook)
	h.RegisterStreamHook("/ttn.lorawan.v3.TestService/FooStream", "hook5", counterStreamHook)

	callUnary("/ttn.lorawan.v3.TestService/Foo")
	expected += 3 // hook1, hook2, hook3
	a.So(actual, should.Equal, expected)

	callStream("/ttn.lorawan.v3.TestService/BarStream")
	expected += 1 // hook4
	a.So(actual, should.Equal, expected)

	callStream("/ttn.lorawan.v3.TestService/FooStream")
	expected += 2 // hook4, hook5
	a.So(actual, should.Equal, expected)

	callUnary("/ttn.lorawan.v3.AnotherService/Baz")
	expected += 0
	a.So(actual, should.Equal, expected)

	h.UnregisterUnaryHook("/ttn.lorawan.v3.TestService", "hook1")
	h.UnregisterUnaryHook("/ttn.lorawan.v3.TestService", "hook2")
	h.UnregisterUnaryHook("/ttn.lorawan.v3.TestService/Foo", "hook3")
	h.UnregisterStreamHook("/ttn.lorawan.v3.TestService", "hook4")
	h.UnregisterStreamHook("/ttn.lorawan.v3.TestService/FooStream", "hook5")

	callUnary("/ttn.lorawan.v3.TestService/Foo")
	expected += 0
	a.So(actual, should.Equal, expected)
}

func TestErrorHook(t *testing.T) {
	a := assertions.New(t)

	h := &Hooks{}

	ctx := test.Context()
	interceptor := h.UnaryServerInterceptor()
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

	h.RegisterUnaryHook("/ttn.lorawan.v3.TestService/Foo", "hook1", newCallbackUnaryHook("hook1", callback))
	h.RegisterUnaryHook("/ttn.lorawan.v3.TestService/Foo", "hook2", errorHook)
	h.RegisterUnaryHook("/ttn.lorawan.v3.TestService/Foo", "hook3", newCallbackUnaryHook("hook3", callback))

	err := call("/ttn.lorawan.v3.TestService/Foo")
	a.So(err, should.NotBeNil)
	a.So(actual, should.Resemble, []string{"hook1"})

	h.UnregisterUnaryHook("/ttn.lorawan.v3.TestService/Foo", "hook1")
	h.UnregisterUnaryHook("/ttn.lorawan.v3.TestService/Foo", "hook2")
	h.UnregisterUnaryHook("/ttn.lorawan.v3.TestService/Foo", "hook3")
}

func TestOrder(t *testing.T) {
	a := assertions.New(t)

	h := &Hooks{}

	ctx := test.Context()
	interceptor := h.UnaryServerInterceptor()
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

	h.RegisterUnaryHook("/ttn.lorawan.v3.TestService/Foo", "hook3", newCallbackUnaryHook("hook3", callback))
	h.RegisterUnaryHook("/ttn.lorawan.v3.TestService/Foo", "hook4", newCallbackUnaryHook("hook4", callback))
	h.RegisterUnaryHook("/ttn.lorawan.v3.TestService", "hook1", newCallbackUnaryHook("hook1", callback)) // service hooks go first
	h.RegisterUnaryHook("/ttn.lorawan.v3.TestService", "hook2", newCallbackUnaryHook("hook2", callback))

	call("/ttn.lorawan.v3.TestService/Foo")
	a.So(actual, should.Resemble, []string{"hook1", "hook2", "hook3", "hook4"})

	h.UnregisterUnaryHook("/ttn.lorawan.v3.TestService/Foo", "hook3")
	h.UnregisterUnaryHook("/ttn.lorawan.v3.TestService/Foo", "hook4")
	h.UnregisterUnaryHook("/ttn.lorawan.v3.TestService", "hook1")
	h.UnregisterUnaryHook("/ttn.lorawan.v3.TestService", "hook2")
}

func TestHookContext(t *testing.T) {
	a := assertions.New(t)

	h := &Hooks{}

	ctx := context.WithValue(test.Context(), "global-value", 1337)
	interceptor := h.UnaryServerInterceptor()
	call := func(fullMethod string) error {
		_, err := interceptor(ctx, "test", &grpc.UnaryServerInfo{
			FullMethod: fullMethod,
		}, noopUnaryHandler)
		return err
	}

	h.RegisterUnaryHook("/ttn.lorawan.v3.TestService", "producer", func(h grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req any) (any, error) {
			ctx = context.WithValue(ctx, "produced-value", 42)
			return h(ctx, req)
		}
	})

	h.RegisterUnaryHook("/ttn.lorawan.v3.TestService", "consumer", func(h grpc.UnaryHandler) grpc.UnaryHandler {
		return func(ctx context.Context, req any) (any, error) {
			a.So(ctx.Value("global-value"), should.Equal, 1337)
			a.So(ctx.Value("produced-value"), should.Equal, 42)
			return h(ctx, req)
		}
	})

	call("/ttn.lorawan.v3.TestService/Test")

	h.UnregisterUnaryHook("/ttn.lorawan.v3.TestService", "producer")
	h.UnregisterUnaryHook("/ttn.lorawan.v3.TestService", "consumer")
}
