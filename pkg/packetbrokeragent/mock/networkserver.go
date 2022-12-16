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

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

// NetworkServer is a mock Network Server.
type NetworkServer struct {
	ttnpb.UnimplementedGsNsServer

	*component.Component
	Uplink chan *ttnpb.UplinkMessage
	TxAck  chan *ttnpb.GatewayTxAcknowledgment
}

// NewNetworkServer returns a new NetworkServer.
func NewNetworkServer(c *component.Component) (*NetworkServer, error) {
	ns := &NetworkServer{
		Component: c,
		Uplink:    make(chan *ttnpb.UplinkMessage, 1),
	}
	c.RegisterGRPC(ns)
	return ns, nil
}

// Roles implements rpcserver.Registerer.
func (ns *NetworkServer) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_NETWORK_SERVER}
}

// RegisterServices implements rpcserver.Registerer.
func (ns *NetworkServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterGsNsServer(s, ns)
}

// RegisterHandlers implements rpcserver.Registerer.
func (ns *NetworkServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
}

// Publish publishes the given message to Packet Broker Agent in the cluster.
func (ns *NetworkServer) Publish(ctx context.Context, down *ttnpb.DownlinkMessage) error {
	client := ttnpb.NewNsPbaClient(ns.LoopbackConn())
	_, err := client.PublishDownlink(ctx, down, ns.WithClusterAuth())
	return err
}

// HandleUplink implements ttnpb.GsNsServer.
func (ns *NetworkServer) HandleUplink(ctx context.Context, req *ttnpb.UplinkMessage) (*pbtypes.Empty, error) {
	select {
	case ns.Uplink <- req:
	default:
	}
	return ttnpb.Empty, nil
}

// ReportTxAcknowledgment implements ttnpb.GsNsServer.
func (ns *NetworkServer) ReportTxAcknowledgment(_ context.Context, req *ttnpb.GatewayTxAcknowledgment) (*pbtypes.Empty, error) {
	select {
	case ns.TxAck <- req:
	default:
	}
	return ttnpb.Empty, nil
}
