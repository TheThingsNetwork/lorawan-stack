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

	"github.com/gogo/protobuf/types"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

var (
	deviceID = ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: "bar",
		},
		DeviceID: "foo",
	}
	downlinkQueueReq = &ttnpb.DownlinkQueueRequest{
		EndDeviceIdentifiers: deviceID,
	}
	applicationUp = &ttnpb.ApplicationUp{
		EndDeviceIdentifiers: deviceID,
		Up: &ttnpb.ApplicationUp_UplinkMessage{
			UplinkMessage: &ttnpb.ApplicationUplink{
				SessionKeyID: []byte{0x11},
				FPort:        42,
				FCnt:         42,
				FRMPayload:   []byte{0x1, 0x2, 0x3},
			},
		},
	}
)

func TestNewRPCServer(t *testing.T) {
	a := assertions.New(t)
	ctx, cancel := context.WithCancel(test.Context())
	defer cancel()

	logHandler := &mockHandler{}
	logger, err := log.NewLogger(log.WithHandler(logHandler))
	ctx = log.NewContext(ctx, logger)

	server := rpcserver.New(ctx,
		rpcserver.WithContextFiller(
			func(ctx context.Context) context.Context {
				return context.WithValue(ctx, &mockKey{}, "foo")
			}),
		rpcserver.WithFieldExtractor(func(fullMethod string, req interface{}) map[string]interface{} {
			return map[string]interface{}{
				"method": fullMethod,
				"foo":    "bar",
			}
		}),
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
			"peer.address":        "pipe",
			"grpc.request.method": "/ttn.lorawan.v3.AppAs/DownlinkQueuePush",
			"grpc.request.foo":    "bar",
		})
		a.So(mock.pushCtx.Value(&mockKey2{}), should.Resemble, "bar")

		runtime.Gosched()

		a.So(logHandler.entries, should.HaveLength, 1)
	})

	t.Run("Stream", func(t *testing.T) {
		a := assertions.New(t)

		sub, err := cli.Subscribe(ctx, &ttnpb.ApplicationIdentifiers{
			ApplicationID: "bar",
		})
		a.So(sub, should.NotBeNil)
		a.So(err, should.BeNil)

		msg, err := sub.Recv()
		a.So(msg, should.NotBeNil)
		a.So(err, should.BeNil)
		a.So(msg, should.Resemble, applicationUp)

		a.So(mock.subCtx.Value(&mockKey{}), should.Resemble, "foo")
		a.So(grpc_ctxtags.Extract(mock.subCtx).Values(), should.Resemble, map[string]interface{}{
			"peer.address":        "pipe",
			"grpc.request.method": "/ttn.lorawan.v3.AppAs/Subscribe",
			"grpc.request.foo":    "bar",
		})
		a.So(mock.subCtx.Value(&mockKey2{}), should.Resemble, "foo")

		runtime.Gosched()

		a.So(logHandler.entries, should.HaveLength, 2)
	})
}

type mockKey struct{}
type mockKey2 struct{}

type mockServer struct {
	ttnpb.AppAsServer

	pushCtx context.Context
	pushReq *ttnpb.DownlinkQueueRequest

	subCtx context.Context
	subIDs *ttnpb.ApplicationIdentifiers
}

type mockHandler struct {
	entries []log.Entry
}

func (h *mockHandler) HandleLog(entry log.Entry) error {
	h.entries = append(h.entries, entry)
	return nil
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
