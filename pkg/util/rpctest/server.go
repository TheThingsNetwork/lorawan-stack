// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package rpctest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"go.thethings.network/lorawan-stack/pkg/errorcontext"
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
		case foo := <-fooCh:
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
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case <-time.After(100 * time.Millisecond):
			if err := stream.Send(&Bar{Message: foo.Message}); err != nil {
				return err
			}
		}
	}
}

// BidiStream RPC example
func (s *FooBarExampleServer) BidiStream(stream FooBar_BidiStreamServer) error {
	fooCh := make(chan *Foo)
	ctx, cancel := errorcontext.New(stream.Context())

	go func() {
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
		case foo := <-fooCh:
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
