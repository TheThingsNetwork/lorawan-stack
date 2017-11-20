// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rpctest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errorcontext"
)

// The FooBarExampleServer is an example/test server
type FooBarExampleServer struct{}

// Unary RPC example
func (s *FooBarExampleServer) Unary(ctx context.Context, foo *Foo) (*Bar, error) {
	return &Bar{Message: foo.Message + foo.Message}, nil
}

// ClientStream RPC example
func (s *FooBarExampleServer) ClientStream(stream FooBar_ClientStreamServer) error {
	fooCh := make(chan *Foo)
	ctx, cancel := errorcontext.New(stream.Context())

	defer cancel(context.Canceled)

	go func() {
		defer close(fooCh)
		for {
			foo, err := stream.Recv()
			if err != nil {
				cancel(err)
				return
			}
			fooCh <- foo
		}
	}()

	var received uint64

	for {
		select {
		case <-ctx.Done():
			switch err := ctx.Err(); err {
			case io.EOF:
				return stream.SendAndClose(&Bar{Message: fmt.Sprintf("Thanks for the %d Foo", received)})
			default:
				return err
			}
		case foo, ok := <-fooCh:
			if !ok {
				continue // will select ctx.Done() in next loop
			}
			if foo.Message == "reset" {
				received = 0
			}
			received++
		case <-time.After(100 * time.Millisecond):
			cancel(errors.New("stream expired")) // will select ctx.Done() in next loop
		}
	}
}

// ServerStream RPC example
func (s *FooBarExampleServer) ServerStream(foo *Foo, stream FooBar_ServerStreamServer) error {
	for i := 0; i < 10; i++ {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case <-time.After(100 * time.Millisecond):
			if err := stream.Send(&Bar{Message: foo.Message}); err != nil {
				return err
			}
		}
	}
	return nil
}

// BidiStream RPC example
func (s *FooBarExampleServer) BidiStream(stream FooBar_BidiStreamServer) error {
	fooCh := make(chan *Foo)
	ctx, cancel := errorcontext.New(stream.Context())

	go func() {
		defer close(fooCh)
		for {
			foo, err := stream.Recv()
			if err != nil {
				cancel(err)
				return
			}
			fooCh <- foo
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case foo, ok := <-fooCh:
			if !ok {
				continue // will select ctx.Done() in next loop
			}
			if err := stream.Send(&Bar{Message: foo.Message}); err != nil {
				return err
			}
		case <-time.After(100 * time.Millisecond):
			if err := stream.Send(&Bar{Message: "bar"}); err != nil {
				return err
			}
		}
	}
}
