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
	"bytes"
	"context"
	"crypto/tls"
	"net"

	northboundv1 "go.thethings.industries/pkg/api/gen/tti/gateway/controller/northbound/v1"
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttgc"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

const profileGroup = "tts"

type component interface {
	GetTLSClientConfig(context.Context, ...tlsconfig.Option) (*tls.Config, error)
}

// Upstream is the client for The Things Gateway Controller.
type Upstream struct {
	component
	client *ttgc.Client
}

// New returns a new upstream client for The Things Gateway Controller.
func New(ctx context.Context, c ttgc.Component, config ttgc.Config) (*Upstream, error) {
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
// Claim does four things:
//  1. Claim the gateway
//  2. Upsert a LoRa Packet Forwarder profile with the root CA presented by the given Gateway Server
//  3. Upsert a Geolocation profile
//  4. Update the gateway with the profiles
func (u *Upstream) Claim(ctx context.Context, eui types.EUI64, ownerToken, clusterAddress string) error {
	logger := log.FromContext(ctx)

	// Claim the gateway.
	gtwClient := northboundv1.NewGatewayServiceClient(u.client)
	_, err := gtwClient.Claim(ctx, &northboundv1.GatewayServiceClaimRequest{
		GatewayId:  eui.MarshalNumber(),
		Domain:     u.client.Domain(ctx),
		OwnerToken: ownerToken,
	})
	if err != nil {
		return err
	}

	// Get the root CA from the Gateway Server and upsert the LoRa Packet Forwarder profile.
	host, _, err := net.SplitHostPort(clusterAddress)
	if err != nil {
		host = clusterAddress
	}
	clusterAddress = net.JoinHostPort(host, "8889")
	rootCA, err := u.getRootCA(ctx, clusterAddress)
	if err != nil {
		return err
	}
	var (
		loraPFProfileID []byte
		loraPFProfile   = &northboundv1.LoraPacketForwarderProfile{
			ProfileName: clusterAddress,
			Shared:      true,
			Protocol:    northboundv1.LoraPacketForwarderProtocol_LORA_PACKET_FORWARDER_PROTOCOL_TTI_V1,
			Address:     clusterAddress,
			RootCa:      rootCA.Raw,
		}
		loraPFProfileClient = northboundv1.NewLoraPacketForwarderProfileServiceClient(u.client)
	)
	loraPFGetRes, err := loraPFProfileClient.GetByName(
		ctx,
		&northboundv1.LoraPacketForwarderProfileServiceGetByNameRequest{
			Domain:      u.client.Domain(ctx),
			Group:       profileGroup,
			ProfileName: clusterAddress,
		},
	)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			logger.WithError(err).Warn("Failed to get LoRa Packet Forwarder profile")
			return err
		}
		res, err := loraPFProfileClient.Create(ctx, &northboundv1.LoraPacketForwarderProfileServiceCreateRequest{
			Domain:                     u.client.Domain(ctx),
			Group:                      profileGroup,
			LoraPacketForwarderProfile: loraPFProfile,
		})
		if err != nil {
			logger.WithError(err).Warn("Failed to create LoRa Packet Forwarder profile")
			return err
		}
		loraPFProfileID = res.ProfileId
	} else {
		if profile := loraPFGetRes.LoraPacketForwarderProfile; profile.Shared != loraPFProfile.Shared ||
			profile.Protocol != loraPFProfile.Protocol ||
			!bytes.Equal(profile.RootCa, loraPFProfile.RootCa) {
			_, err := loraPFProfileClient.Update(ctx, &northboundv1.LoraPacketForwarderProfileServiceUpdateRequest{
				Domain:                     u.client.Domain(ctx),
				Group:                      profileGroup,
				ProfileId:                  loraPFGetRes.ProfileId,
				LoraPacketForwarderProfile: loraPFProfile,
			})
			if err != nil {
				logger.WithError(err).Warn("Failed to update LoRa Packet Forwarder profile")
				return err
			}
		}
		loraPFProfileID = loraPFGetRes.ProfileId
	}

	// Upsert the Geolocation profile.
	var (
		geolocationProfileID []byte
		geolocationProfile   = &northboundv1.GeolocationProfile{
			ProfileName:     "on connect",
			Shared:          true,
			DisconnectedFor: durationpb.New(0),
		}
		geolocationProfileClient = northboundv1.NewGeolocationProfileServiceClient(u.client)
	)
	geolocationGetRes, err := geolocationProfileClient.GetByName(
		ctx,
		&northboundv1.GeolocationProfileServiceGetByNameRequest{
			Domain:      u.client.Domain(ctx),
			Group:       profileGroup,
			ProfileName: geolocationProfile.ProfileName,
		},
	)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			logger.WithError(err).Warn("Failed to get geolocation profile")
			return err
		}
		res, err := geolocationProfileClient.Create(ctx, &northboundv1.GeolocationProfileServiceCreateRequest{
			Domain:             u.client.Domain(ctx),
			Group:              profileGroup,
			GeolocationProfile: geolocationProfile,
		})
		if err != nil {
			logger.WithError(err).Warn("Failed to create geolocation profile")
			return err
		}
		geolocationProfileID = res.ProfileId
	} else {
		geolocationProfileID = geolocationGetRes.ProfileId
	}

	// Update the gateway with the profiles.
	_, err = gtwClient.Update(ctx, &northboundv1.GatewayServiceUpdateRequest{
		GatewayId: eui.MarshalNumber(),
		Domain:    u.client.Domain(ctx),
		LoraPacketForwarderProfileId: &northboundv1.ProfileIDValue{
			Value: loraPFProfileID,
		},
		GeolocationProfileId: &northboundv1.ProfileIDValue{
			Value: geolocationProfileID,
		},
	})
	if err != nil {
		logger.WithError(err).Warn("Failed to update gateway with profiles")
		return err
	}

	return nil
}

// Unclaim implements gateways.GatewayClaimer.
func (u *Upstream) Unclaim(ctx context.Context, eui types.EUI64) error {
	gtwClient := northboundv1.NewGatewayServiceClient(u.client)
	_, err := gtwClient.Unclaim(ctx, &northboundv1.GatewayServiceUnclaimRequest{
		GatewayId: eui.MarshalNumber(),
		Domain:    u.client.Domain(ctx),
	})
	if err != nil {
		return err
	}
	return nil
}

// IsManagedGateway implements gateways.GatewayClaimer.
// This method always returns true.
func (*Upstream) IsManagedGateway(context.Context, types.EUI64) (bool, error) {
	return true, nil
}
