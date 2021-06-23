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

// Package packetbroker abstracts the Packet Broker Agent to the upstream.Handler interface.
package packetbroker

import (
	"context"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc"
)

const (
	publishUplinkTimeout = 3 * time.Second

	updateGatewayInterval  = 5 * time.Minute
	updateGatewayJitter    = 0.2
	updateGatewayTTLMargin = 10 * time.Second
)

// Cluster represents the interface the cluster.
type Cluster interface {
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	WithClusterAuth() grpc.CallOption
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

// DevAddrPrefixes implements upstream.Handler.
func (h *Handler) DevAddrPrefixes() []types.DevAddrPrefix {
	return h.devAddrPrefixes
}

// Setup implements upstream.Handler.
func (h *Handler) Setup(context.Context) error {
	return nil
}

func nextUpdateGateway(onlineTTL *pbtypes.Duration) <-chan time.Time {
	d := random.Jitter(updateGatewayInterval, updateGatewayJitter)
	if onlineTTL != nil {
		ttl, err := pbtypes.DurationFromProto(onlineTTL)
		if err == nil {
			ttl -= updateGatewayTTLMargin
			if ttl < d {
				d = ttl
			}
		}
	}
	return time.After(d)
}

// ConnectGateway implements upstream.Handler.
func (h *Handler) ConnectGateway(ctx context.Context, ids ttnpb.GatewayIdentifiers, conn *io.Connection) error {
	pbaConn, err := h.cluster.GetPeerConn(ctx, ttnpb.ClusterRole_PACKET_BROKER_AGENT, &ids)
	if err != nil {
		return errPacketBrokerAgentNotFound.WithCause(err)
	}
	pbaClient := ttnpb.NewGsPbaClient(pbaConn)

	gtw := conn.Gateway()
	req := &ttnpb.UpdatePacketBrokerGatewayRequest{
		Gateway: &ttnpb.Gateway{
			GatewayIdentifiers: ids,
			Antennas:           gtw.Antennas,
			FrequencyPlanIDs:   gtw.FrequencyPlanIDs,
			StatusPublic:       gtw.StatusPublic,
			LocationPublic:     gtw.LocationPublic,
		},
		Online: true,
		FieldMask: &pbtypes.FieldMask{
			Paths: []string{
				"antennas",
				"frequency_plan_ids",
				"location_public",
				"status_public",
			},
		},
	}
	res, err := pbaClient.UpdateGateway(ctx, req, h.cluster.WithClusterAuth())
	if err != nil {
		return err
	}

	var (
		onlineTTL                = res.OnlineTtl
		lastCounters             = time.Now()
		lastUplinkCount   uint64 = 0
		lastDownlinkCount uint64 = 0
	)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-nextUpdateGateway(onlineTTL):
		}

		req := &ttnpb.UpdatePacketBrokerGatewayRequest{
			Gateway: &ttnpb.Gateway{
				GatewayIdentifiers: ids,
				StatusPublic:       gtw.StatusPublic,
			},
			Online: true,
			FieldMask: &pbtypes.FieldMask{
				Paths: []string{
					"status_public",
				},
			},
		}

		// Only update the location when it is public and when it may be updated from status messages.
		// location_public should only be in the field mask if the location is known, so only when a location in the status.
		// This is to avoid that the location gets reset when there is no location in the status.
		if gtw.LocationPublic && gtw.UpdateLocationFromStatus {
			if status, _, ok := conn.StatusStats(); ok && len(status.GetAntennaLocations()) > 0 && status.AntennaLocations[0] != nil {
				loc := *status.AntennaLocations[0]
				loc.Source = ttnpb.SOURCE_GPS
				req.Gateway.LocationPublic = true
				req.Gateway.Antennas = []ttnpb.GatewayAntenna{
					{
						Location: &loc,
					},
				}
				req.FieldMask.Paths = append(req.FieldMask.Paths, "antennas", "location_public")
			}
		}

		now := time.Now()
		uplinkCount, _, haveUplinkCount := conn.UpStats()
		downlinkCount, _, haveDownlinkCount := conn.DownStats()
		if haveUplinkCount {
			req.RxRate = &pbtypes.FloatValue{
				Value: (float32(uplinkCount) - float32(lastUplinkCount)) * float32(time.Hour) / float32(now.Sub(lastCounters)),
			}
			lastUplinkCount = uplinkCount
		}
		if haveDownlinkCount {
			req.TxRate = &pbtypes.FloatValue{
				Value: (float32(downlinkCount) - float32(lastDownlinkCount)) * float32(time.Hour) / float32(now.Sub(lastCounters)),
			}
			lastDownlinkCount = downlinkCount
		}
		lastCounters = now

		res, err := pbaClient.UpdateGateway(ctx, req, h.cluster.WithClusterAuth())
		if err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to update gateway")
			onlineTTL = nil
		} else {
			onlineTTL = res.OnlineTtl
		}
	}
}

var errPacketBrokerAgentNotFound = errors.DefineNotFound("packet_broker_agent_not_found", "Packet Broker Agent not found")

// HandleUplink implements upstream.Handler.
func (h *Handler) HandleUplink(ctx context.Context, _ ttnpb.GatewayIdentifiers, ids ttnpb.EndDeviceIdentifiers, msg *ttnpb.GatewayUplinkMessage) error {
	pbaConn, err := h.cluster.GetPeerConn(ctx, ttnpb.ClusterRole_PACKET_BROKER_AGENT, &ids)
	if err != nil {
		return errPacketBrokerAgentNotFound.WithCause(err)
	}
	ctx, cancel := context.WithTimeout(ctx, publishUplinkTimeout)
	defer cancel()
	_, err = ttnpb.NewGsPbaClient(pbaConn).PublishUplink(ctx, msg, h.cluster.WithClusterAuth())
	return err
}

// HandleStatus implements upstream.Handler.
func (h *Handler) HandleStatus(context.Context, ttnpb.GatewayIdentifiers, *ttnpb.GatewayStatus) error {
	return nil
}

// HandleTxAck implements upstream.Handler.
func (h *Handler) HandleTxAck(context.Context, ttnpb.GatewayIdentifiers, *ttnpb.TxAcknowledgment) error {
	return nil
}
