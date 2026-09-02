// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

// Package ttgc provides functions to use The Things Gateway Controller.
package ttgc

import (
	"context"
	"crypto/tls"
	"slices"
	"strconv"

	northboundv1 "go.thethings.industries/pkg/api/gen/tti/gateway/controller/northbound/v1"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	dcstypes "go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/types"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttgc"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc"
)

var errNoSupportedClaimOption = errors.DefineFailedPrecondition(
	"no_supported_claim_option",
	"no supported claim option (protocol + auth method) found for gateway",
)

const (
	profileGroup = "tts"

	// Default LoRa Basics Station LNS port of the Gateway Server.
	defaultLNSPort uint16 = 8887
)

type component interface {
	GetTLSConfig(context.Context) tlsconfig.Config
	GetTLSClientConfig(context.Context, ...tlsconfig.Option) (*tls.Config, error)
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	AllowInsecureForCredentials() bool
}

// Upstream is the client for The Things Gateway Controller.
type Upstream struct {
	component
	client        *ttgc.Client
	lnsPort       string
	managedRanges []types.EUI64Range
}

// New returns a new upstream client for The Things Gateway Controller.
func New(ctx context.Context, c component, config ttgc.Config) (*Upstream, error) {
	client, err := ttgc.NewClient(ctx, c, config)
	if err != nil {
		return nil, err
	}
	lnsPort := config.LBSCUPS.LNSPort
	if lnsPort == 0 {
		lnsPort = defaultLNSPort
	}
	return &Upstream{
		component:     c,
		client:        client,
		lnsPort:       strconv.Itoa(int(lnsPort)),
		managedRanges: config.ManagedGatewayEUIs,
	}, nil
}

// claimOption represents the protocol and authentication method for claiming a gateway.
type claimOption struct {
	protocol   northboundv1.GatewayProtocolIdentifier
	authMethod northboundv1.AuthenticationMethod
	handler    func(context.Context, *ttnpb.GatewayIdentifiers, string, string) (*dcstypes.GatewayMetadata, error)
}

// Claim implements gateways.GatewayClaimer.
func (u *Upstream) Claim(
	ctx context.Context, ids *ttnpb.GatewayIdentifiers, ownerToken, clusterAddress string,
) (*dcstypes.GatewayMetadata, error) {
	eui := types.MustEUI64(ids.Eui).OrZero()

	// Get the gateway description to verify what protocol it supports.
	gtwClient := northboundv1.NewGatewayServiceClient(u.client)
	desc, err := gtwClient.Describe(ctx, &northboundv1.GatewayServiceDescribeRequest{
		GatewayId: eui.MarshalNumber(),
	})
	if err != nil {
		return nil, err
	}

	// Defines the preferred claiming options in order.
	claimPreferences := []claimOption{
		{
			protocol:   northboundv1.GatewayProtocolIdentifier_GATEWAY_PROTOCOL_IDENTIFIER_TTI_V1,
			authMethod: northboundv1.AuthenticationMethod_AUTHENTICATION_METHOD_MUTUAL_TLS,
			handler:    u.claimTTIV1Gateway,
		},
		{
			protocol:   northboundv1.GatewayProtocolIdentifier_GATEWAY_PROTOCOL_IDENTIFIER_LBS_LNS,
			authMethod: northboundv1.AuthenticationMethod_AUTHENTICATION_METHOD_GATEWAY_TOKEN,
			handler:    u.claimLBSCUPSGateway,
		},
	}

	// Select the first supported claiming option and use its handler.
	for _, option := range claimPreferences {
		if u.supportsOption(desc, option) {
			return option.handler(ctx, ids, ownerToken, clusterAddress)
		}
	}

	return nil, errNoSupportedClaimOption.New()
}

func (*Upstream) supportsOption(
	desc *northboundv1.GatewayServiceDescribeResponse,
	option claimOption,
) bool {
	for _, p := range desc.SupportedGatewayProtocols {
		if p.GatewayProtocolId != option.protocol {
			continue
		}
		if slices.Contains(p.SupportedAuthenticationMethods, option.authMethod) {
			return true
		}
	}

	return false
}

// Unclaim implements gateways.GatewayClaimer.
func (u *Upstream) Unclaim(ctx context.Context, ids *ttnpb.GatewayIdentifiers) error {
	eui := types.MustEUI64(ids.Eui).OrZero()

	gtwClient := northboundv1.NewGatewayServiceClient(u.client)
	_, err := gtwClient.Unclaim(ctx, &northboundv1.GatewayServiceUnclaimRequest{
		GatewayId: eui.MarshalNumber(),
		Domain:    u.client.Domain(ctx),
	})
	if err != nil {
		if errors.IsNotFound(err) { // The gateway does not exist or is already unclaimed.
			return nil
		}
		return err
	}

	// Delete the CUPS and LNS API keys for the gateway.
	if err := u.deleteAPIKeys(ctx, ids); err != nil {
		// Don't fail unclaiming if deleting the API keys fails.
		log.FromContext(ctx).WithError(err).Warn("Failed to delete API keys for gateway")
	}

	return nil
}

// IsManagedGateway implements gateways.GatewayClaimer.
// This method returns true for gateways in the configured managed Gateway EUI ranges and false otherwise.
func (u *Upstream) IsManagedGateway(_ context.Context, eui types.EUI64) (bool, error) {
	for _, r := range u.managedRanges {
		if r.Contains(eui) {
			return true, nil
		}
	}
	return false, nil
}
