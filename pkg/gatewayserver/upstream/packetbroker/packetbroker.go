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
	updateGatewayTimeout = 5 * time.Second

	DefaultUpdateGatewayInterval = 10 * time.Minute
	DefaultUpdateGatewayJitter   = 0.2
	DefaultOnlineTTLMargin       = 10 * time.Second
)

// Config configures the Handler.
type Config struct {
	UpdateInterval  time.Duration
	UpdateJitter    float64
	OnlineTTLMargin time.Duration
	DevAddrPrefixes []types.DevAddrPrefix
	GatewayRegistry GatewayRegistry
	Cluster         Cluster
}

// GatewayRegistry is a store with gateways.
type GatewayRegistry interface {
	Get(ctx context.Context, req *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error)
}

// Cluster represents the interface the cluster.
type Cluster interface {
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	WithClusterAuth() grpc.CallOption
}

// Handler is the upstream handler.
type Handler struct {
	ctx context.Context
	Config
}

// NewHandler returns a new upstream handler.
func NewHandler(ctx context.Context, config Config) *Handler {
	return &Handler{
		ctx:    ctx,
		Config: config,
	}
}

// DevAddrPrefixes implements upstream.Handler.
func (h *Handler) DevAddrPrefixes() []types.DevAddrPrefix {
	return h.Config.DevAddrPrefixes
}

// Setup implements upstream.Handler.
func (h *Handler) Setup(context.Context) error {
	return nil
}

func (h *Handler) nextUpdateGateway(onlineTTL *pbtypes.Duration) <-chan time.Time {
	d := random.Jitter(h.UpdateInterval, h.UpdateJitter)
	if onlineTTL != nil {
		ttl, err := pbtypes.DurationFromProto(onlineTTL)
		if err == nil {
			ttl -= h.OnlineTTLMargin
			if ttl < d {
				d = ttl
			}
		}
	}
	return time.After(d)
}

// ConnectGateway implements upstream.Handler.
func (h *Handler) ConnectGateway(ctx context.Context, ids ttnpb.GatewayIdentifiers, conn *io.Connection) error {
	pbaConn, err := h.Cluster.GetPeerConn(ctx, ttnpb.ClusterRole_PACKET_BROKER_AGENT, nil)
	if err != nil {
		return errPacketBrokerAgentNotFound.WithCause(err)
	}
	pbaClient := ttnpb.NewGsPbaClient(pbaConn)

	gtw := conn.Gateway()
	antennas := make([]*ttnpb.GatewayAntenna, len(gtw.Antennas))
	for i, ant := range gtw.Antennas {
		antennas[i] = ant
	}
	req := &ttnpb.UpdatePacketBrokerGatewayRequest{
		Gateway: &ttnpb.PacketBrokerGateway{
			Ids: &ttnpb.PacketBrokerGateway_GatewayIdentifiers{
				GatewayId: ids.GatewayId,
				Eui:       ids.Eui,
			},
			Antennas:         antennas,
			FrequencyPlanIds: gtw.FrequencyPlanIds,
			StatusPublic:     gtw.StatusPublic,
			LocationPublic:   gtw.LocationPublic,
			Online:           true,
			RxRate: &pbtypes.FloatValue{
				Value: 0,
			},
			TxRate: &pbtypes.FloatValue{
				Value: 0,
			},
		},
		FieldMask: &pbtypes.FieldMask{
			Paths: []string{
				"antennas",
				"frequency_plan_ids",
				"location_public",
				"online",
				"rx_rate",
				"status_public",
				"tx_rate",
			},
		},
	}
	updateCtx, cancel := context.WithTimeout(ctx, updateGatewayTimeout)
	res, err := pbaClient.UpdateGateway(updateCtx, req, h.Cluster.WithClusterAuth())
	cancel()
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
		case <-h.nextUpdateGateway(onlineTTL):
		}

		req := &ttnpb.UpdatePacketBrokerGatewayRequest{
			Gateway: &ttnpb.PacketBrokerGateway{
				Ids: &ttnpb.PacketBrokerGateway_GatewayIdentifiers{
					GatewayId: ids.GatewayId,
					Eui:       ids.Eui,
				},
				Online:       true,
				StatusPublic: gtw.StatusPublic,
			},
			FieldMask: &pbtypes.FieldMask{
				Paths: []string{
					"online",
					"status_public",
				},
			},
		}

		if gtw.LocationPublic {
			// Only update the location when it is public and when it may be updated from status messages.
			// location_public should only be in the field mask if the location is known, so only when a location in the status.
			// This is to avoid that the location gets reset when there is no location in the status.
			if status, _, ok := conn.StatusStats(); ok && gtw.UpdateLocationFromStatus && len(status.GetAntennaLocations()) > 0 && status.AntennaLocations[0] != nil {
				loc := *status.AntennaLocations[0]
				loc.Source = ttnpb.LocationSource_SOURCE_GPS
				req.Gateway.LocationPublic = true
				req.Gateway.Antennas = []*ttnpb.GatewayAntenna{
					{
						Location: &loc,
					},
				}
				req.FieldMask.Paths = append(req.FieldMask.GetPaths(), "antennas", "location_public")
			}
		} else {
			// Explicitly disable location public so that the existing gateway location, if any, gets reset.
			req.FieldMask.Paths = append(req.FieldMask.GetPaths(), "location_public")
		}

		now := time.Now()
		uplinkCount, _, haveUplinkCount := conn.UpStats()
		downlinkCount, _, haveDownlinkCount := conn.DownStats()
		if haveUplinkCount {
			req.Gateway.RxRate = &pbtypes.FloatValue{
				Value: (float32(uplinkCount) - float32(lastUplinkCount)) * float32(time.Hour) / float32(now.Sub(lastCounters)),
			}
			req.FieldMask.Paths = append(req.FieldMask.Paths, "rx_rate")
			lastUplinkCount = uplinkCount
		}
		if haveDownlinkCount {
			req.Gateway.TxRate = &pbtypes.FloatValue{
				Value: (float32(downlinkCount) - float32(lastDownlinkCount)) * float32(time.Hour) / float32(now.Sub(lastCounters)),
			}
			req.FieldMask.Paths = append(req.FieldMask.Paths, "tx_rate")
			lastDownlinkCount = downlinkCount
		}
		lastCounters = now

		pbaConn, err := h.Cluster.GetPeerConn(ctx, ttnpb.ClusterRole_PACKET_BROKER_AGENT, nil)
		if err != nil {
			return errPacketBrokerAgentNotFound.WithCause(err)
		}

		updateCtx, cancel := context.WithTimeout(ctx, updateGatewayTimeout)
		res, err := ttnpb.NewGsPbaClient(pbaConn).UpdateGateway(updateCtx, req, h.Cluster.WithClusterAuth())
		cancel()
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
func (h *Handler) HandleUplink(ctx context.Context, _ ttnpb.GatewayIdentifiers, ids *ttnpb.EndDeviceIdentifiers, msg *ttnpb.GatewayUplinkMessage) error {
	pbaConn, err := h.Cluster.GetPeerConn(ctx, ttnpb.ClusterRole_PACKET_BROKER_AGENT, nil)
	if err != nil {
		return errPacketBrokerAgentNotFound.WithCause(err)
	}
	ctx, cancel := context.WithTimeout(ctx, publishUplinkTimeout)
	defer cancel()
	_, err = ttnpb.NewGsPbaClient(pbaConn).PublishUplink(ctx, msg, h.Cluster.WithClusterAuth())
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
