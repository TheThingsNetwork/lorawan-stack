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
	"runtime"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/log/handler/memory"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

var (
	deviceID = ttnpb.EndDeviceIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{
			ApplicationId: "bar",
		},
		DeviceId: "foo",
	}
	downlinkQueueReq = &ttnpb.DownlinkQueueRequest{
		EndDeviceIds: &deviceID,
	}
	applicationUp = &ttnpb.ApplicationUp{
		EndDeviceIds: &deviceID,
		Up: &ttnpb.ApplicationUp_UplinkMessage{
			UplinkMessage: &ttnpb.ApplicationUplink{
				SessionKeyId: []byte{0x11},
				FPort:        42,
				FCnt:         42,
				FrmPayload:   []byte{0x1, 0x2, 0x3},
			},
		},
	}
)

func TestNewRPCServer(t *testing.T) {
	a := assertions.New(t)
	ctx, cancel := context.WithCancel(test.Context())
	defer cancel()

	logHandler := memory.New()
	logger := log.NewLogger(logHandler)
	ctx = log.NewContext(ctx, logger)

	server := rpcserver.New(ctx,
		rpcserver.WithContextFiller(
			func(ctx context.Context) context.Context {
				return context.WithValue(ctx, &mockKey{}, "foo")
			},
		),
		rpcserver.WithUnaryInterceptors(UnaryServerInterceptor),
		rpcserver.WithStreamInterceptors(StreamServerInterceptor),
	)
	a.So(server, should.NotBeNil)
	mock := &mockServer{}
	ttnpb.RegisterAppAsServer(server.Server, mock)

	loopbackConn, err := rpcserver.StartLoopback(ctx, server.Server, rpcclient.DefaultDialOptions(ctx)...)
	a.So(loopbackConn, should.NotBeNil)
	a.So(err, should.BeNil)

	cli := ttnpb.NewAppAsClient(loopbackConn)

	t.Run("Unary", func(t *testing.T) {
		a := assertions.New(t)

		_, err = cli.DownlinkQueuePush(ctx, downlinkQueueReq)
		a.So(err, should.BeNil)

		a.So(mock.pushReq, should.NotBeNil)
		a.So(mock.pushReq, should.Resemble, downlinkQueueReq)

		a.So(mock.pushCtx, should.NotBeNil)
		a.So(mock.pushCtx.Value(&mockKey{}), should.Resemble, "foo")
		a.So(grpc_ctxtags.Extract(mock.pushCtx).Values(), should.Resemble, map[string]interface{}{
			"peer.address":                "pipe",
			"grpc.request.device_id":      "foo",
			"grpc.request.application_id": "bar",
		})
		a.So(mock.pushCtx.Value(&mockKey2{}), should.Resemble, "bar")

		runtime.Gosched()
		time.Sleep(test.Delay)

		a.So(logHandler.Entries, should.HaveLength, 1)
	})

	t.Run("Stream", func(t *testing.T) {
		a := assertions.New(t)

		sub, err := cli.Subscribe(ctx, &ttnpb.ApplicationIdentifiers{
			ApplicationId: "bar",
		})
		a.So(sub, should.NotBeNil)
		a.So(err, should.BeNil)

		msg, err := sub.Recv()
		a.So(msg, should.NotBeNil)
		a.So(err, should.BeNil)
		a.So(msg, should.Resemble, applicationUp)

		a.So(mock.subCtx.Value(&mockKey{}), should.Resemble, "foo")
		a.So(grpc_ctxtags.Extract(mock.subCtx).Values(), should.Resemble, map[string]interface{}{
			"peer.address":                "pipe",
			"grpc.request.application_id": "bar",
		})
		a.So(mock.subCtx.Value(&mockKey2{}), should.Resemble, "foo")

		runtime.Gosched()
		time.Sleep(test.Delay)

		a.So(logHandler.Entries, should.HaveLength, 2)
	})
}

type (
	mockKey  struct{}
	mockKey2 struct{}
)

type mockServer struct {
	ttnpb.UnimplementedAppAsServer

	pushCtx context.Context
	pushReq *ttnpb.DownlinkQueueRequest

	subCtx context.Context
	subIDs *ttnpb.ApplicationIdentifiers

	decCtx context.Context
	decReq *ttnpb.DecodeDownlinkRequest
}

func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx = context.WithValue(ctx, &mockKey2{}, "bar")
	return handler(ctx, req)
}

func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	wrapped := grpc_middleware.WrapServerStream(ss)
	wrapped.WrappedContext = context.WithValue(ss.Context(), &mockKey2{}, "foo")
	return handler(srv, wrapped)
}

func (s *mockServer) Subscribe(ids *ttnpb.ApplicationIdentifiers, srv ttnpb.AppAs_SubscribeServer) error {
	s.subCtx, s.subIDs = srv.Context(), ids
	srv.Send(applicationUp)
	return nil
}

func (s *mockServer) DownlinkQueuePush(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*types.Empty, error) {
	s.pushCtx, s.pushReq = ctx, req
	return &types.Empty{}, nil
}
