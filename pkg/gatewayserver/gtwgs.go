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
	"fmt"
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/validate"
)

type nsErrors map[string]error

func (e nsErrors) Error() string {
	var errors []string
	for nsName, err := range e {
		errors = append(errors, fmt.Sprintf("- %s: %s", nsName, err))
	}
	return strings.Join(errors, "\n")
}

func (g *GatewayServer) getGatewayFrequencyPlan(ctx context.Context, gatewayID *ttnpb.GatewayIdentifiers) (ttnpb.FrequencyPlan, error) {
	isInfo := g.GetPeer(ttnpb.PeerInfo_IDENTITY_SERVER, g.config.NSTags, nil)
	if isInfo == nil {
		return ttnpb.FrequencyPlan{}, errNoIdentityServerFound
	}

	is := ttnpb.NewIsGatewayClient(isInfo.Conn())
	gw, err := is.GetGateway(ctx, gatewayID)
	if err != nil {
		return ttnpb.FrequencyPlan{}, errRetrieveGatewayInformation.WithCause(err)
	}

	fp, err := g.FrequencyPlans.GetByID(gw.FrequencyPlanID)
	if err != nil {
		return ttnpb.FrequencyPlan{}, errRetrieveFrequencyPlanForGateway.WithAttributes("frequency_plan_id", gw.FrequencyPlanID)
	}

	return fp, nil
}

func (g *GatewayServer) forAllNS(f func(ttnpb.GsNsClient) error) error {
	errors := nsErrors{}
	for _, ns := range g.GetPeers(ttnpb.PeerInfo_NETWORK_SERVER, g.config.NSTags) {
		nsClient := ttnpb.NewGsNsClient(ns.Conn())
		err := f(nsClient)
		if err != nil {
			errors[ns.Name()] = err
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

func (g *GatewayServer) signalStartServingGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) {
	if err := g.forAllNS(func(nsClient ttnpb.GsNsClient) error {
		_, err := nsClient.StartServingGateway(ctx, id, g.Component.WithClusterAuth())
		return err
	}); err != nil {
		log.FromContext(ctx).WithError(err).Error("Could not signal Network Server when gateway connected")
	}
}

func (g *GatewayServer) signalStopServingGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) {
	if err := g.forAllNS(func(nsClient ttnpb.GsNsClient) error {
		_, err := nsClient.StopServingGateway(ctx, id, g.Component.WithClusterAuth())
		return err
	}); err != nil {
		log.FromContext(ctx).WithError(err).Error("Could not signal Network Server when gateway disconnected")
	}
}

func (g *GatewayServer) setupConnection(uid string, connectionInfo connection) {
	g.connectionsMu.Lock()
	if conn, ok := g.connections[uid]; ok {
		conn.Close()
		delete(g.connections, uid)
	}
	g.connections[uid] = connectionInfo
	g.connectionsMu.Unlock()
}

// Link the gateway to the Gateway Server. The authentication information will
// be used to determine the gateway ID. If no authentication information is present,
// this gateway may not be used for downlink.
func (g *GatewayServer) Link(link ttnpb.GtwGs_LinkServer) (err error) {
	ctx := link.Context()
	id := ttnpb.GatewayIdentifiers{
		GatewayID: rpcmetadata.FromIncomingContext(ctx).ID,
	}
	if err = validate.ID(id.GetGatewayID()); err != nil {
		return
	}
	if err = rights.RequireGateway(ctx, id, ttnpb.RIGHT_GATEWAY_LINK); err != nil {
		return
	}

	uid := unique.ID(ctx, id)
	logger := log.FromContext(ctx).WithField("gateway_uid", uid)
	ctx = log.NewContext(ctx, logger)

	registerStartGatewayLink(ctx, id)
	defer registerEndGatewayLink(ctx, id)

	logger.Info("Link with gateway opened")
	defer logger.WithError(err).Debug("Link with gateway closed")

	isInfo := g.GetPeer(ttnpb.PeerInfo_IDENTITY_SERVER, g.config.NSTags, nil)
	if isInfo == nil {
		return errNoIdentityServerFound
	}

	gtw, err := g.getGateway(ctx, &id)
	if err != nil {
		return errRetrieveGatewayInformation.WithCause(err)
	}

	fp, err := g.FrequencyPlans.GetByID(gtw.GetFrequencyPlanID())
	if err != nil {
		return errRetrieveFrequencyPlanForGateway.WithAttributes("frequency_plan_id", gtw.GetFrequencyPlanID()).WithCause(err)
	}

	scheduler, err := scheduling.FrequencyPlanScheduler(ctx, fp)
	if err != nil {
		return err
	}

	connectionInfo := &gRPCConnection{
		link: link,
		gtw:  gtw,
		connectionData: connectionData{
			scheduler: scheduler,
		},
	}

	ctx, connectionInfo.cancel = context.WithCancel(ctx)
	defer connectionInfo.cancel()

	g.setupConnection(uid, connectionInfo)

	go g.signalStartServingGateway(ctx, &id)

	go func() {
		<-ctx.Done()
		// TODO: Add tenant extraction when #433 is merged
		stopCtx, cancel := context.WithTimeout(g.Context(), time.Minute)
		g.signalStopServingGateway(stopCtx, &id)
		cancel()

		g.connectionsMu.Lock()
		if oldConnection := g.connections[uid]; oldConnection == connectionInfo {
			delete(g.connections, uid)
		}
		g.connectionsMu.Unlock()
	}()

	logger.Info("Uplink subscription was opened")
	// Uplink receiving routine
	for {
		upstreamMessage, err := link.Recv()
		if err != nil {
			return err
		}
		logger.Debug("Received message from gateway")

		g.handleUpstreamMessage(ctx, connectionInfo, upstreamMessage)
	}
}

func (g *GatewayServer) handleUpstreamMessage(ctx context.Context, conn connection, upstreamMessage *ttnpb.GatewayUp) {
	if upstreamMessage.GatewayStatus != nil {
		g.handleStatus(ctx, upstreamMessage.GatewayStatus, conn)
	}
	for _, uplink := range upstreamMessage.UplinkMessages {
		g.handleUplink(ctx, uplink, conn)
	}

	return
}

func (g *GatewayServer) handleUplink(ctx context.Context, uplink *ttnpb.UplinkMessage, conn connection) (err error) {
	ctx = events.ContextWithCorrelationID(ctx, append(
		uplink.CorrelationIDs,
		fmt.Sprintf("gs:uplink:%s", events.NewCorrelationID()),
	)...)
	uplink.CorrelationIDs = events.CorrelationIDsFromContext(ctx)
	registerReceiveUplink(ctx, conn.gatewayIdentifiers(), uplink)
	conn.addUplinkObservation()

	logger := log.FromContext(ctx)
	defer func() {
		if err != nil {
			logger.WithError(err).Warn("Could not handle uplink")
		} else {
			logger.Debug("Handled uplink")
		}
	}()

	gwMetadata := conn.gateway()
	useLocationFromMetadata := gwMetadata != nil && len(gwMetadata.GetAntennas()) == 0

	for _, antenna := range uplink.RxMetadata {
		antenna.GatewayIdentifiers = conn.gatewayIdentifiers()

		index := int(antenna.GetAntennaIndex())
		if !gwMetadata.GetPrivacySettings().LocationPublic {
			antenna.Location = nil
			continue
		}

		if useLocationFromMetadata {
			continue
		}

		if gwMetadata != nil && len(gwMetadata.GetAntennas()) >= index {
			antenna.Location = &gwMetadata.GetAntennas()[index].Location
		} else {
			antenna.Location = nil
		}
	}

	pld := uplink.GetPayload()

	var ns cluster.Peer
	switch pld.GetMType() {
	case ttnpb.MType_CONFIRMED_UP, ttnpb.MType_UNCONFIRMED_UP:
		if uplink.DevAddr == nil {
			err = errNoDevAddr
			return
		}
		logger = logger.WithField("dev_addr", *uplink.DevAddr)
		var devAddrBytes []byte
		devAddrBytes, err = uplink.DevAddr.Marshal()
		if err != nil {
			return
		}
		ns = g.GetPeer(ttnpb.PeerInfo_NETWORK_SERVER, g.config.NSTags, devAddrBytes)
	case ttnpb.MType_JOIN_REQUEST, ttnpb.MType_REJOIN_REQUEST:
		if uplink.DevEUI == nil {
			err = errNoDevEUI
			return
		}
		logger = logger.WithField("dev_eui", uplink.DevEUI.String())
		var devEUIBytes []byte
		devEUIBytes, err = uplink.DevEUI.Marshal()
		if err != nil {
			return
		}
		ns = g.GetPeer(ttnpb.PeerInfo_NETWORK_SERVER, g.config.NSTags, devEUIBytes)
	}

	if ns == nil {
		err = errNoNetworkServerFound
		return
	}

	nsClient := ttnpb.NewGsNsClient(ns.Conn())
	_, err = nsClient.HandleUplink(g.Context(), uplink, g.Component.WithClusterAuth())
	return
}

func (g *GatewayServer) handleStatus(ctx context.Context, status *ttnpb.GatewayStatus, conn connection) error {
	registerReceiveStatus(ctx, conn.gatewayIdentifiers(), status)
	conn.addStatusObservation(status)
	log.FromContext(ctx).Debug("Received status message")
	return nil
}

// GetFrequencyPlan associated to the gateway.
func (g *GatewayServer) GetFrequencyPlan(ctx context.Context, r *ttnpb.GetFrequencyPlanRequest) (*ttnpb.FrequencyPlan, error) {
	fp, err := g.FrequencyPlans.GetByID(r.GetFrequencyPlanID())
	if err != nil {
		return nil, errRetrieveFrequencyPlanForGateway.WithAttributes("frequency_plan_id", r.GetFrequencyPlanID()).WithCause(err)
	}

	return &fp, nil
}
