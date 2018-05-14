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
	"time"

	"go.thethings.network/lorawan-stack/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/udp"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/version"
)

const (
	gatewayInMemory = time.Hour
	ttnGSIndex      = "ttn-gateway-server"
)

var ttnVersions = map[string]string{ttnGSIndex: version.TTN}

func (g *GatewayServer) runUDPBridge(ctx context.Context, udpConn *net.UDPConn) {
	gwStore := udp.NewGatewayStore(gatewayInMemory)

	conn := udp.Handle(udpConn, gwStore, gwStore)

	logger := log.FromContext(g.Context())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-ctx.Done()
		if err := conn.Close(); err != nil {
			logger.WithError(err).Debug("Could not close UDP connection")
		}
	}()

	udpGateways := map[types.EUI64]*udpConnection{}

	for {
		packet, err := conn.Read()
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

		gatewayConnection, ok := udpGateways[*packet.GatewayEUI]
		if !ok {
			gatewayConnection = &udpConnection{}
			udpGateways[*packet.GatewayEUI] = gatewayConnection
		}

		go g.handleUpstreamUDPMessage(ctx, packet, gatewayConnection)
	}
}

func (g *GatewayServer) handleUpstreamUDPMessage(ctx context.Context, packet *udp.Packet, gateway *udpConnection) {
	logger := log.FromContext(ctx)
	logger.WithField("packet_type", packet.PacketType.String()).Debug("Received packet")

	gtwIDs := ttnpb.GatewayIdentifiers{EUI: packet.GatewayEUI}

	switch packet.PacketType {
	case udp.PullData:
		g.processPullData(ctx, packet, gateway)
	case udp.PushData:
		if packet.Data == nil {
			return
		}

		gtw := gateway.gateway()
		if gtw != nil {
			gtwIDs.GatewayID = gtw.GetGatewayID()
		}

		upstream, err := udp.TranslateUpstream(*packet.Data, udp.UpstreamMetadata{
			ID: gtwIDs,
			IP: packet.GatewayAddr.IP.String(),
		})
		if err != nil {
			logger.WithError(err).Error("Could not translate incoming packet")
			return
		}

		g.handleUpstreamMessage(ctx, gateway, upstream)
	case udp.TxAck:
		logger.Debug("Received downlink reception confirmation")
	}
}

func (g *GatewayServer) processPullData(ctx context.Context, firstPacket *udp.Packet, connection *udpConnection) {
	connection.lastPullDataStorage.Store(firstPacket)
	connection.lastPullDataTime.Store(time.Now())

	if _, ok := connection.gtw.Load().(*ttnpb.Gateway); ok {
		// TODO: Add frequency plan refresh on a regular basis: https://github.com/TheThingsIndustries/ttn/issues/727
		return
	}

	logger := log.FromContext(ctx)
	logger.Info("Fetching gateway information and frequency plan...")

	isPeer := g.GetPeer(ttnpb.PeerInfo_IDENTITY_SERVER, nil, nil)
	if isPeer == nil {
		logger.Error("No identity server to connect to found")
		return
	}
	isConn := isPeer.Conn()
	if isConn == nil {
		logger.Error("No ready connection to the identity server")
		return
	}

	is := ttnpb.NewIsGatewayClient(isConn)
	gtw, err := is.GetGateway(ctx, &ttnpb.GatewayIdentifiers{EUI: firstPacket.GatewayEUI})
	if err != nil {
		logger.WithError(err).Error("Could not retrieve gateway information from the gateway server")
		return
	}

	fpID := gtw.GetFrequencyPlanID()
	logger = logger.WithField("frequency_plan_id", fpID)

	fp, err := g.FrequencyPlans.GetByID(fpID)
	if err != nil {
		logger.WithError(err).Error("Could not retrieve frequency plan")
		return
	}

	scheduler, err := scheduling.FrequencyPlanScheduler(ctx, fp)
	if err != nil {
		logger.WithError(err).Error("Could not build a scheduler from the frequency plan")
		return
	}

	connection.scheduler = scheduler
	connection.gtw.Store(gtw)

	logger.Info("Gateway information and frequency plan fetched")

	g.setupConnection(gtw.GatewayIdentifiers.UniqueID(ctx), connection)
	ctx, cancel := context.WithTimeout(g.Context(), time.Minute)
	g.signalStartServingGateway(ctx, &gtw.GatewayIdentifiers)
	cancel()
}
