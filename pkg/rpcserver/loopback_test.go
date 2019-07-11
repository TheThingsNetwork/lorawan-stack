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

package rpcserver_test

import (
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/util/rpctest"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

func TestLoopbackConn(t *testing.T) {
	a := assertions.New(t)
	ctx, cancel := context.WithCancel(test.Context())
	defer cancel()

	server := grpc.NewServer()
	a.So(server, should.NotBeNil)
	rpctest.RegisterFooBarServer(server, &rpctest.FooBarExampleServer{})

	stats := &statsHandler{}
	loopbackConn, err := rpcserver.StartLoopback(ctx, server, grpc.WithStatsHandler(stats))
	a.So(loopbackConn, should.NotBeNil)
	a.So(err, should.BeNil)

	cli := rpctest.NewFooBarClient(loopbackConn)

	bar, err := cli.Unary(ctx, &rpctest.Foo{Message: "foo"})
	a.So(err, should.BeNil)
	a.So(bar.Message, should.Equal, "foofoo")

	a.So(stats.rpcTag, should.NotBeNil)
	a.So(stats.rpcTag.FullMethodName, should.Equal, "/rpctest.FooBar/Unary")
	a.So(stats.rpcTag.FailFast, should.Equal, true)

	a.So(stats.connTag, should.NotBeNil)
	a.So(stats.connTag.RemoteAddr, should.Resemble, stats.connTag.LocalAddr)
}

type statsHandler struct {
	rpcTag  *stats.RPCTagInfo
	connTag *stats.ConnTagInfo
}

func (s *statsHandler) TagRPC(ctx context.Context, tag *stats.RPCTagInfo) context.Context {
	s.rpcTag = tag
	return ctx
}

func (s *statsHandler) HandleRPC(ctx context.Context, rpcstats stats.RPCStats) {}

func (s *statsHandler) TagConn(ctx context.Context, tag *stats.ConnTagInfo) context.Context {
	s.connTag = tag
	return ctx
}

func (s *statsHandler) HandleConn(ctx context.Context, stats stats.ConnStats) {}
