// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rpctest_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errorcontext"
	"github.com/TheThingsNetwork/ttn/pkg/util/rpctest"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFooBarExampleServer(t *testing.T) {
	a := assertions.New(t)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	rpctest.RegisterFooBarServer(server, &rpctest.FooBarExampleServer{})

	go server.Serve(lis)

	cc, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer cc.Close()

	cli := rpctest.NewFooBarClient(cc)

	{
		bar, err := cli.Unary(context.Background(), &rpctest.Foo{Message: "foo"})
		a.So(err, should.BeNil)
		a.So(bar.Message, should.Equal, "foofoo")
	}

	{
		stream, err := cli.ClientStream(context.Background())
		a.So(err, should.BeNil)
		err = stream.Send(&rpctest.Foo{Message: "foo"})
		a.So(err, should.BeNil)
		bar, err := stream.CloseAndRecv()
		a.So(err, should.BeNil)
		a.So(bar.Message, should.Equal, "Thanks for the 1 Foo")
	}

	{
		ctx, cancel := context.WithCancel(context.Background())
		stream, err := cli.ClientStream(ctx)
		a.So(err, should.BeNil)
		err = stream.Send(&rpctest.Foo{Message: "foo"})
		a.So(err, should.BeNil)
		cancel()
		_, err = stream.CloseAndRecv()
		a.So(grpc.Code(err), should.Equal, codes.Canceled)
	}

	{
		stream, err := cli.ClientStream(context.Background())
		a.So(err, should.BeNil)
		err = stream.Send(&rpctest.Foo{Message: "foo"})
		a.So(err, should.BeNil)
		time.Sleep(150 * time.Millisecond)
		err = stream.RecvMsg(&empty.Empty{})
		a.So(grpc.Code(err), should.Equal, codes.Unknown)
	}

	{
		ctx, cancel := context.WithCancel(context.Background())
		stream, err := cli.ServerStream(ctx, &rpctest.Foo{Message: "foo"})
		a.So(err, should.BeNil)

		bar, err := stream.Recv()
		a.So(err, should.BeNil)
		a.So(bar.Message, should.Equal, "foo")

		bar, err = stream.Recv()
		a.So(err, should.BeNil)
		a.So(bar.Message, should.Equal, "foo")

		cancel()
		bar, err = stream.Recv()
		status, ok := status.FromError(err)
		a.So(ok, should.BeTrue)
		a.So(status.Code(), should.Equal, codes.Canceled)
	}

	{
		ctx, cancel := context.WithCancel(context.Background())
		stream, err := cli.BidiStream(ctx)
		a.So(err, should.BeNil)

		bar, err := stream.Recv()
		a.So(err, should.BeNil)
		a.So(bar.Message, should.Equal, "bar")

		err = stream.Send(&rpctest.Foo{Message: "foo"})
		a.So(err, should.BeNil)

		bar, err = stream.Recv()
		a.So(err, should.BeNil)
		a.So(bar.Message, should.Equal, "foo")

		cancel()
		bar, err = stream.Recv()
		status, ok := status.FromError(err)
		a.So(ok, should.BeTrue)
		a.So(status.Code(), should.Equal, codes.Canceled)
	}
}

func watchClientStream(ctx *errorcontext.ErrorContext, stream rpctest.FooBar_ClientStreamClient) <-chan *rpctest.Bar {
	ch := make(chan *rpctest.Bar)
	bar := new(rpctest.Bar)
	go func() {
		err := stream.RecvMsg(bar)
		if err == nil {
			ch <- bar
		} else {
			ctx.Cancel(err)
		}
		close(ch)
	}()
	return ch
}
