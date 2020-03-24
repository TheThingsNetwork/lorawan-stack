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

package mock

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// GatewayServer is a mock Gateway Server.
type GatewayServer struct {
	*component.Component
	Downlink chan *ttnpb.DownlinkMessage
}

// NewGatewayServer returns a new GatewayServer.
func NewGatewayServer(c *component.Component) (*GatewayServer, error) {
	gs := &GatewayServer{
		Component: c,
		Downlink:  make(chan *ttnpb.DownlinkMessage, 1),
	}
	c.RegisterGRPC(gs)
	return gs, nil
}

// Roles implements rpcserver.Registerer.
func (gs *GatewayServer) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_GATEWAY_SERVER}
}

// RegisterServices implements rpcserver.Registerer.
func (gs *GatewayServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterNsGsServer(s, gs)
}

// RegisterHandlers implements rpcserver.Registerer.
func (gs *GatewayServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
}

// Publish publishes the given message to Packet Broker Agent in the cluster.
func (gs *GatewayServer) Publish(ctx context.Context, up *ttnpb.GatewayUplinkMessage) error {
	client := ttnpb.NewGsPbaClient(gs.LoopbackConn())
	_, err := client.PublishUplink(ctx, up, gs.WithClusterAuth())
	return err
}

// ScheduleDownlink implements ttnpb.NsGsServer.
func (gs *GatewayServer) ScheduleDownlink(ctx context.Context, req *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
	select {
	case gs.Downlink <- req:
	default:
	}
	return &ttnpb.ScheduleDownlinkResponse{}, nil
}
