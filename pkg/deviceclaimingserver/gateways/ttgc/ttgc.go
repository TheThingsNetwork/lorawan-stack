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

var (
	errNoSupportedProtocol   = errors.DefineFailedPrecondition("no_supported_protocol", "no supported gateway protocol found for claiming")
	errNoSupportedAuthMethod = errors.DefineFailedPrecondition("no_supported_auth_method", "no supported authentication method found for gateway")
)

const profileGroup = "tts"

type component interface {
	GetTLSConfig(context.Context) tlsconfig.Config
	GetTLSClientConfig(context.Context, ...tlsconfig.Option) (*tls.Config, error)
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	AllowInsecureForCredentials() bool
}

// Upstream is the client for The Things Gateway Controller.
type Upstream struct {
	component
	client *ttgc.Client

	gatewayAccess ttnpb.GatewayAccessClient
}

// New returns a new upstream client for The Things Gateway Controller.
func New(ctx context.Context, c component, config ttgc.Config) (*Upstream, error) {
	client, err := ttgc.NewClient(ctx, c, config)
	if err != nil {
		return nil, err
	}
	return &Upstream{
		component: c,
		client:    client,
	}, nil
}

// Claim implements gateways.GatewayClaimer.
func (u *Upstream) Claim(
	ctx context.Context, eui types.EUI64, ownerToken, clusterAddress string,
) (*dcstypes.GatewayMetadata, error) {
	// Get the gateway description to verify what protocol it supports.
	gtwClient := northboundv1.NewGatewayServiceClient(u.client)
	desc, err := gtwClient.Describe(ctx, &northboundv1.GatewayServiceDescribeRequest{
		GatewayId: eui.MarshalNumber(),
	})
	if err != nil {
		return nil, err
	}

	if u.supportsProtocol(desc, northboundv1.GatewayProtocolIdentifier_GATEWAY_PROTOCOL_TTI_V1) {
		return u.claimTTIV1Gateway(ctx, eui, ownerToken, clusterAddress)
	}

	if u.supportsProtocol(desc, northboundv1.GatewayProtocolIdentifier_GATEWAY_PROTOCOL_LBS_CUPS) {
		return u.claimLBSCUPSGateway(ctx, eui, ownerToken, clusterAddress)
	}

	return nil, errNoSupportedProtocol.New()
}

func (*Upstream) supportsProtocol(
	desc *northboundv1.GatewayServiceDescribeResponse,
	protocolID northboundv1.GatewayProtocolIdentifier,
) bool {
	for _, p := range desc.SupportedGatewayProtocols {
		if p.GatewayProtocolId == protocolID {
			return true
		}
	}

	return false
}

// Unclaim implements gateways.GatewayClaimer.
func (u *Upstream) Unclaim(ctx context.Context, eui types.EUI64) error {
	// Delete the CUPS and LNS API keys for the gateway.
	if err := u.deleteAPIKeys(ctx, &ttnpb.GatewayIdentifiers{Eui: eui.Bytes()}); err != nil {
		// Don't fail unclaiming if deleting the API keys fails.
		log.FromContext(ctx).WithError(err).Warn("Failed to delete API keys for gateway")
	}

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
	return nil
}

// IsManagedGateway implements gateways.GatewayClaimer.
// This method always returns true.
func (*Upstream) IsManagedGateway(context.Context, types.EUI64) (bool, error) {
	return true, nil
}
