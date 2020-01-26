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

// Package ns abstracts the V3 Network Server to the upstream.Handler interface.
package ns

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"google.golang.org/grpc"
)

// Cluster provides cluster operations.
type Cluster interface {
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) (*grpc.ClientConn, error)
	WithClusterAuth() grpc.CallOption
	ClaimIDs(ctx context.Context, ids ttnpb.Identifiers) error
	UnclaimIDs(ctx context.Context, ids ttnpb.Identifiers) error
}

// Handler is the upstream handler.
type Handler struct {
	ctx             context.Context
	cluster         Cluster
	devAddrPrefixes []types.DevAddrPrefix
}

// NewHandler returns a new upstream handler.
func NewHandler(ctx context.Context, cluster Cluster, devAddrPrefixes []types.DevAddrPrefix) *Handler {
	return &Handler{
		ctx:             ctx,
		cluster:         cluster,
		devAddrPrefixes: devAddrPrefixes,
	}
}

// GetDevAddrPrefixes implements upstream.Handler.
func (h *Handler) GetDevAddrPrefixes() []types.DevAddrPrefix {
	return h.devAddrPrefixes
}

// Setup implements upstream.Handler.
func (h *Handler) Setup(context.Context) error {
	return nil
}

// ConnectGateway implements upstream.Handler.
func (h *Handler) ConnectGateway(ctx context.Context, ids ttnpb.GatewayIdentifiers, conn *io.Connection) error {
	// If the frontend can claim downlinks, don't claim automatically on connection.
	if conn.Frontend().SupportsDownlinkClaim() {
		return nil
	}
	h.cluster.ClaimIDs(ctx, ids)
	select {
	case <-ctx.Done():
		h.cluster.UnclaimIDs(ctx, ids)
		return ctx.Err()
	default:
		return nil
	}
}

var errNetworkServerNotFound = errors.DefineNotFound("network_server_not_found", "Network Server not found")

// HandleUplink implements upstream.Handler.
func (h *Handler) HandleUplink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, msg *ttnpb.GatewayUplinkMessage) error {
	nsConn, err := h.cluster.GetPeerConn(ctx, ttnpb.ClusterRole_NETWORK_SERVER, ids)
	if err != nil {
		return errNetworkServerNotFound.WithCause(err)
	}
	_, err = ttnpb.NewGsNsClient(nsConn).HandleUplink(ctx, msg.UplinkMessage, h.cluster.WithClusterAuth())
	return err
}

// HandleStatus implements upstream.Handler.
func (h *Handler) HandleStatus(context.Context, ttnpb.GatewayIdentifiers, *ttnpb.GatewayStatus) error {
	return nil
}
