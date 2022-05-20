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

	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc"
)

// Cluster provides cluster operations.
type Cluster interface {
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	WithClusterAuth() grpc.CallOption
	ClaimIDs(ctx context.Context, ids cluster.EntityIdentifiers) error
	UnclaimIDs(ctx context.Context, ids cluster.EntityIdentifiers) error
}

// ContextDecoupler decouples the request context from its values.
type ContextDecoupler interface {
	FromRequestContext(ctx context.Context) context.Context
}

// Handler is the upstream handler.
type Handler struct {
	ctx              context.Context
	cluster          Cluster
	contextDecoupler ContextDecoupler
	devAddrPrefixes  []types.DevAddrPrefix
}

// NewHandler returns a new upstream handler.
func NewHandler(ctx context.Context, cluster Cluster, contextDecoupler ContextDecoupler, devAddrPrefixes []types.DevAddrPrefix) *Handler {
	return &Handler{
		ctx:              ctx,
		cluster:          cluster,
		contextDecoupler: contextDecoupler,
		devAddrPrefixes:  devAddrPrefixes,
	}
}

// DevAddrPrefixes implements upstream.Handler.
func (h *Handler) DevAddrPrefixes() []types.DevAddrPrefix {
	return h.devAddrPrefixes
}

// Setup implements upstream.Handler.
func (h *Handler) Setup(context.Context) error {
	return nil
}

// ConnectGateway implements upstream.Handler.
func (h *Handler) ConnectGateway(ctx context.Context, ids *ttnpb.GatewayIdentifiers, conn *io.Connection) error {
	// If the frontend can claim downlinks, don't claim automatically on connection.
	if conn.Frontend().SupportsDownlinkClaim() {
		return nil
	}
	decoupledCtx := h.contextDecoupler.FromRequestContext(ctx)
	logger := log.FromContext(ctx)
	if err := h.cluster.ClaimIDs(decoupledCtx, ids); err != nil {
		logger.WithError(err).Error("Failed to claim downlink path")
		return err
	}
	logger.Info("Downlink path claimed")
	defer func() {
		if err := h.cluster.UnclaimIDs(decoupledCtx, ids); err != nil {
			logger.WithError(err).Error("Failed to unclaim downlink path")
			return
		}
		logger.Info("Downlink path unclaimed")
	}()
	<-ctx.Done()
	return ctx.Err()
}

var errNetworkServerNotFound = errors.DefineNotFound("network_server_not_found", "Network Server not found")

// HandleUplink implements upstream.Handler.
func (h *Handler) HandleUplink(ctx context.Context, _ *ttnpb.GatewayIdentifiers, ids *ttnpb.EndDeviceIdentifiers, msg *ttnpb.GatewayUplinkMessage) error {
	nsConn, err := h.cluster.GetPeerConn(ctx, ttnpb.ClusterRole_NETWORK_SERVER, nil)
	if err != nil {
		return errNetworkServerNotFound.WithCause(err)
	}
	_, err = ttnpb.NewGsNsClient(nsConn).HandleUplink(ctx, msg.Message, h.cluster.WithClusterAuth())
	return err
}

// HandleStatus implements upstream.Handler.
func (h *Handler) HandleStatus(context.Context, *ttnpb.GatewayIdentifiers, *ttnpb.GatewayStatus) error {
	return nil
}

// HandleTxAck implements upstream.Handler.
func (h *Handler) HandleTxAck(ctx context.Context, ids *ttnpb.GatewayIdentifiers, msg *ttnpb.TxAcknowledgment) error {
	nsConn, err := h.cluster.GetPeerConn(ctx, ttnpb.ClusterRole_NETWORK_SERVER, nil)
	if err != nil {
		return errNetworkServerNotFound.WithCause(err)
	}
	_, err = ttnpb.NewGsNsClient(nsConn).ReportTxAcknowledgment(ctx, &ttnpb.GatewayTxAcknowledgment{
		TxAck:      msg,
		GatewayIds: ids,
	}, h.cluster.WithClusterAuth())
	return err
}
