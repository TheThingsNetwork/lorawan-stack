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
		errors = append(errors, fmt.Sprintf("%s: %s", nsName, err))
	}
	return strings.Join(errors, " ; ")
}

func (g *GatewayServer) getGatewayFrequencyPlan(ctx context.Context, gatewayID *ttnpb.GatewayIdentifier) (ttnpb.FrequencyPlan, error) {
	isInfo := g.GetPeer(ttnpb.PeerInfo_IDENTITY_SERVER, nil, nil)
	if isInfo == nil {
		return ttnpb.FrequencyPlan{}, errors.New("No identity server to connect to")
	}

	is := ttnpb.NewIsGatewayClient(isInfo.Conn())
	gw, err := is.GetGateway(g.Context(), gatewayID)
	if err != nil {
		return ttnpb.FrequencyPlan{}, errors.NewWithCause("Could not get gateway information", err)
	}

	fp, err := g.frequencyPlans.GetByID(gw.FrequencyPlanID)
	if err != nil {
		return ttnpb.FrequencyPlan{}, errors.NewWithCause(fmt.Sprintf("Could not retrieve frequency plan %s", gw.FrequencyPlanID), err)
	}

	return fp, nil
}

func (g *GatewayServer) forAllNS(f func(ttnpb.GsNsClient) error) error {
	errors := nsErrors{}
	for _, ns := range g.GetPeers(ttnpb.PeerInfo_NETWORK_SERVER, nil) {
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
func (g *GatewayServer) Link(link ttnpb.GtwGs_LinkServer) error {
	md := rpcmetadata.FromIncomingContext(link.Context())
	id := ttnpb.GatewayIdentifier{
		GatewayID: md.ID,
	}

	fp, err := g.getGatewayFrequencyPlan(link.Context(), &id)
	if err != nil {
		return errors.NewWithCause("Could not get frequency plan for this gateway", err)
	}

	result, err := g.gateways.Subscribe(id, link, fp)
	if err != nil {
		return err
	}

	logger := g.Logger().WithField("gateway_id", id.GatewayID)

	ctx := link.Context()

	go func() {
		startServingGatewayFn := func(nsClient ttnpb.GsNsClient) error {
			_, err := nsClient.StartServingGateway(ctx, &id)
			return err
		}
		if err := g.forAllNS(startServingGatewayFn); err != nil {
			logger.WithError(err).Errorf("An error occurred when signaling to the network servers "+
				" that the gateway %s is being served", id)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			logger.WithError(ctx.Err()).Warn("Stopped serving Rx packets")
			go func() {
				stopCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
				stopServingGatewayFn := func(nsClient ttnpb.GsNsClient) error {
					_, err := nsClient.StopServingGateway(stopCtx, &id)
					return err
				}
				if err := g.forAllNS(stopServingGatewayFn); err != nil {
					logger.WithError(err).Errorf("An error occurred when signaling to the network servers "+
						" that the gateway %s is not being served anymore", id)
				}
				cancel()
			}()
			return ctx.Err()
		case upstream := <-result:
			if upstream != nil {
				g.handleStatus(logger, upstream.GatewayStatus)
				for _, uplink := range upstream.UplinkMessages {
					g.handleUplink(logger, uplink)
				}
			}
		}
	}
}

func (g *GatewayServer) handleUplink(logger log.Interface, uplink *ttnpb.UplinkMessage) {
	var err error
	defer func() {
		if err != nil {
			logger.WithError(err).Warn("Could not handle uplink")
		} else {
			logger.Debug("Uplink handled")
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

	ns := g.GetPeer(ttnpb.PeerInfo_NETWORK_SERVER, nil, devAddrBytes)
	if ns == nil {
		return
	}

	nsClient := ttnpb.NewGsNsClient(ns.Conn())
	_, err = nsClient.HandleUplink(g.Context(), uplink)
}

func (g *GatewayServer) handleStatus(logger log.Interface, status *ttnpb.GatewayStatus) {
	logger.Debug("Received status message")
	return
}

// GetFrequencyPlan associated to the gateway. The gateway is ID'd by its authentication token.
func (g *GatewayServer) GetFrequencyPlan(ctx context.Context, r *ttnpb.FrequencyPlanRequest) (*ttnpb.FrequencyPlan, error) {
	fp, err := g.frequencyPlans.GetByID(r.GetFrequencyPlanID())
	if err != nil {
		return nil, errors.NewWithCause("Could not retrieve frequency plan from storage", err)
	}

	return &fp, nil
}
