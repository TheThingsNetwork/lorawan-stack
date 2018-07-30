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

package gatewayserver

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/udp"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

const gatewayInMemory = time.Hour

func (g *GatewayServer) runUDPEndpoint(ctx context.Context, rawConn *net.UDPConn) {
	gwStore := udp.NewGatewayStore(gatewayInMemory)

	udpConn := udp.Handle(rawConn, gwStore, gwStore)

	logger := log.FromContext(g.Context())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-ctx.Done()
		if err := udpConn.Close(); err != nil {
			logger.WithError(err).Debug("Could not close UDP connection")
		}
	}()

	udpGateways := &sync.Map{}

	for {
		packet, err := udpConn.Read()
		if err != nil && err != io.EOF {
			logger.WithError(err).Error("Could not read incoming UDP packets")
		}
		if err != nil {
			return
		}

		if packet.GatewayEUI == nil {
			logger.Error("No gateway EUI in the packet, dropping the packet")
			continue
		}
		logger := logger.WithField("gateway_eui", packet.GatewayEUI.String())
		ctx := log.NewContext(ctx, logger)
		if err := packet.Ack(); err != nil {
			logger.WithError(err).Error("Could not acknowledge incoming packet")
		}

		v, loaded := udpGateways.LoadOrStore(*packet.GatewayEUI, &udpConnState{eui: packet.GatewayEUI})
		conn := v.(*udpConnState)

		if !loaded {
			go func() {
				if err := g.setupUDPConnection(ctx, conn); err != nil {
					udpGateways.Delete(*packet.GatewayEUI)
				}
			}()
		}

		go g.handleUpstreamUDPMessage(ctx, packet, conn)
	}
}

// CustomUDPContextFiller allows for filling the context for the UDP connection.
var CustomUDPContextFiller func(ctx context.Context, id ttnpb.GatewayIdentifiers) (context.Context, error)

func (g *GatewayServer) setupUDPConnection(ctx context.Context, conn *udpConnState) error {
	// TODO: Add frequency plan refresh on a regular basis: https://github.com/TheThingsIndustries/ttn/issues/727

	ctx = g.Component.FillContext(ctx)
	if filler := CustomUDPContextFiller; filler != nil {
		var err error
		ctx, err = filler(ctx, ttnpb.GatewayIdentifiers{EUI: conn.eui})
		if err != nil {
			return err
		}
	}

	logger := log.FromContext(ctx)
	logger.Info("Fetching gateway information and frequency plan...")

	gtw, err := g.getGateway(ctx, &ttnpb.GatewayIdentifiers{EUI: conn.eui})
	if err != nil {
		logger.WithError(err).Error("Could not retrieve gateway information from the Gateway Server")
		return err
	}
	uid := unique.ID(ctx, gtw.GatewayIdentifiers)
	logger = logger.WithField("gateway_uid", uid)
	conn.gtw.Store(gtw)

	fpID := gtw.GetFrequencyPlanID()
	logger = logger.WithField("frequency_plan_id", fpID)
	fp, err := g.FrequencyPlans.GetByID(fpID)
	if err != nil {
		logger.WithError(err).Error("Could not retrieve frequency plan")
		return err
	}
	scheduler, err := scheduling.FrequencyPlanScheduler(ctx, fp)
	if err != nil {
		logger.WithError(err).Error("Could not build a scheduler from the frequency plan")
		return err
	}
	conn.scheduler = scheduler

	logger.Info("Gateway information and frequency plan fetched")

	g.setupConnection(uid, conn)

	// TODO: Claim identifiers here (https://github.com/TheThingsIndustries/lorawan-stack/issues/941)
	go func() {
		g.signalStartServingGateway(ctx, &gtw.GatewayIdentifiers)
	}()

	conn.ctx = ctx
	return nil
}

func (g *GatewayServer) handleUpstreamUDPMessage(ctx context.Context, packet *udp.Packet, conn *udpConnState) {
	logger := log.FromContext(ctx).WithField("packet_type", packet.PacketType.String())
	logger.Debug("Received packet")

	switch packet.PacketType {
	case udp.PullData:
		g.processPullData(log.NewContext(ctx, logger), packet, conn)
	case udp.PushData:
		g.processPushData(log.NewContext(ctx, logger), packet, conn)
	case udp.TxAck:
		logger.Debug("Received downlink reception confirmation")
		conn.hasSentTxAck.Store(true)
	}
}

func (g *GatewayServer) processPullData(ctx context.Context, packet *udp.Packet, conn *udpConnState) {
	conn.lastPullDataStorage.Store(packet)
	conn.lastPullDataTime.Store(time.Now())
}

func (g *GatewayServer) processPushData(ctx context.Context, packet *udp.Packet, conn *udpConnState) {
	if packet.Data == nil {
		return
	}

	gtwIDs := ttnpb.GatewayIdentifiers{EUI: packet.GatewayEUI}
	gtw := conn.gateway()
	if gtw != nil {
		gtwIDs.GatewayID = gtw.GetGatewayID()
	}

	logger := log.FromContext(ctx)
	upstream, err := udp.TranslateUpstream(*packet.Data, udp.UpstreamMetadata{
		ID: gtwIDs,
		IP: packet.GatewayAddr.IP.String(),
	})
	if err != nil {
		logger.WithError(err).Error("Could not translate incoming packet")
		return
	}

	if len(packet.Data.RxPacket) > 0 {
		var maxTmst uint32
		for _, rxMetadata := range packet.Data.RxPacket {
			if rxMetadata.Tmst > maxTmst {
				maxTmst = rxMetadata.Tmst
			}
		}
		conn.syncClock(maxTmst)
	}
	g.handleUpstreamMessage(ctx, conn, upstream)
}
