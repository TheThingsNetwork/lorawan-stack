// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gatewayserver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

type nsErrors map[string]error

func (e nsErrors) Error() string {
	var errors []string
	for nsName, err := range e {
		errors = append(errors, fmt.Sprintf("- %s: %s", nsName, err))
	}
	return strings.Join(errors, "\n")
}

func (g *GatewayServer) getGatewayFrequencyPlan(ctx context.Context, gatewayID *ttnpb.GatewayIdentifier) (ttnpb.FrequencyPlan, error) {
	isInfo := g.GetPeer(ttnpb.PeerInfo_IDENTITY_SERVER, g.nsTags, nil)
	if isInfo == nil {
		return ttnpb.FrequencyPlan{}, errors.New("No identity server to connect to")
	}

	is := ttnpb.NewIsGatewayClient(isInfo.Conn())
	gw, err := is.GetGateway(ctx, gatewayID)
	if err != nil {
		return ttnpb.FrequencyPlan{}, errors.NewWithCause(err, "Could not get gateway information from identity server")
	}

	fp, err := g.frequencyPlans.GetByID(gw.FrequencyPlanID)
	if err != nil {
		return ttnpb.FrequencyPlan{}, errors.NewWithCausef(err, "Could not retrieve frequency plan %s", gw.FrequencyPlanID)
	}

	return fp, nil
}

func (g *GatewayServer) forAllNS(f func(ttnpb.GsNsClient) error) error {
	errors := nsErrors{}
	for _, ns := range g.GetPeers(ttnpb.PeerInfo_NETWORK_SERVER, g.nsTags) {
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

// Link the gateway to the gateway server. The authentication information will
// be used to determine the gateway ID. If no authentication information is present,
// this gateway may not be used for downlink.
func (g *GatewayServer) Link(link ttnpb.GtwGs_LinkServer) (err error) {
	ctx := link.Context()
	md := rpcmetadata.FromIncomingContext(ctx)
	id := ttnpb.GatewayIdentifier{
		GatewayID: md.ID,
	}

	logger := log.FromContext(ctx).WithField("gateway_id", id.GatewayID)
	defer logger.WithError(err).Debug("Link with gateway closed")

	fp, err := g.getGatewayFrequencyPlan(ctx, &id)
	if err != nil {
		return errors.NewWithCause(err, "Could not get frequency plan for this gateway")
	}

	result, err := g.gateways.Subscribe(id, link, fp)
	if err != nil {
		return err
	}

	go func() {
		startServingGatewayFn := func(nsClient ttnpb.GsNsClient) error {
			_, err := nsClient.StartServingGateway(ctx, &id)
			return err
		}
		if err := g.forAllNS(startServingGatewayFn); err != nil {
			logger.WithError(err).Error("Could not signal NS when gateway connected")
		}
	}()

	go func() {
		<-ctx.Done()
		// TODO: Add tenant extraction when #433 is merged
		stopCtx, cancel := context.WithTimeout(g.Context(), time.Minute)
		stopServingGatewayFn := func(nsClient ttnpb.GsNsClient) error {
			_, err := nsClient.StopServingGateway(stopCtx, &id)
			return err
		}
		if err := g.forAllNS(stopServingGatewayFn); err != nil {
			logger.WithError(err).Errorf("Could not signal NS when gateway disconnected")
		}
		cancel()
	}()

	ctx = log.NewContext(ctx, logger)
	for {
		select {
		case <-ctx.Done():
			logger.WithError(ctx.Err()).Warn("Stopped serving Rx packets")
			return ctx.Err()
		case upstream, ok := <-result:
			if !ok {
				logger.Debug("Uplink subscription was closed")
				return nil
			}
			if upstream != nil {
				if upstream.GatewayStatus != nil {
					g.handleStatus(ctx, upstream.GatewayStatus)
				}
				for _, uplink := range upstream.UplinkMessages {
					g.handleUplink(ctx, uplink)
				}
			}
		}
	}
}

func (g *GatewayServer) handleUplink(ctx context.Context, uplink *ttnpb.UplinkMessage) (err error) {
	logger := log.FromContext(ctx)
	defer func() {
		if err != nil {
			logger.WithError(err).Warn("Could not handle uplink")
		} else {
			logger.Debug("Handled uplink")
		}
	}()

	if uplink.DevAddr == nil {
		err = errors.New("No DevAddr specified")
		return
	}
	logger = logger.WithField("devaddr", *uplink.DevAddr)
	devAddr := *uplink.DevAddr
	devAddrBytes, err := devAddr.Marshal()
	if err != nil {
		return
	}

	ns := g.GetPeer(ttnpb.PeerInfo_NETWORK_SERVER, g.nsTags, devAddrBytes)
	if ns == nil {
		err = ErrNoNetworkServerFound.New(errors.Attributes{
			"devaddr": uplink.DevAddr.String(),
		})
		return
	}

	nsClient := ttnpb.NewGsNsClient(ns.Conn())
	_, err = nsClient.HandleUplink(g.Context(), uplink)
	return
}

func (g *GatewayServer) handleStatus(ctx context.Context, status *ttnpb.GatewayStatus) error {
	log.FromContext(ctx).Debug("Received status message")
	return nil
}

// GetFrequencyPlan associated to the gateway. The gateway is ID'd by its authentication token.
func (g *GatewayServer) GetFrequencyPlan(ctx context.Context, r *ttnpb.FrequencyPlanRequest) (*ttnpb.FrequencyPlan, error) {
	fp, err := g.frequencyPlans.GetByID(r.GetFrequencyPlanID())
	if err != nil {
		return nil, errors.NewWithCause(err, "Could not retrieve frequency plan from storage")
	}

	return &fp, nil
}
