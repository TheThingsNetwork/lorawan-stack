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

package gatewayserver_test

import (
	"context"
	"net"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// GsNsServer implements ttnpb.GsNsServer
type GsNsServer struct {
	messageReceived chan string
}

func (s GsNsServer) HandleUplink(_ context.Context, up *ttnpb.UplinkMessage) (*types.Empty, error) {
	go func() {
		s.messageReceived <- "HandleUplink"
	}()
	return ttnpb.Empty, nil
}

func StartMockGsNsServer(ctx context.Context) (GsNsServer, string) {
	ns := GsNsServer{messageReceived: make(chan string)}

	serve := func(ctx context.Context, addr string) string {
		srv := rpcserver.New(ctx)
		ttnpb.RegisterGsNsServer(srv.Server, ns)

		for {
			lis, err := net.Listen("tcp", addr)
			if err == nil {
				go srv.Serve(lis)
				return lis.Addr().String()
			}
		}
	}

	addr := serve(ctx, "127.0.0.1:0")
	return ns, addr
}
