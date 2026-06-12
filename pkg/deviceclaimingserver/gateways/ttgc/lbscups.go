// Copyright © 2026 The Things Network Foundation, The Things Industries B.V.
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

package ttgc

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	northboundv1 "go.thethings.industries/pkg/api/gen/tti/gateway/controller/northbound/v1"
	dcstypes "go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/types"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errCreateAPIKey = errors.DefineFailedPrecondition("create_api_key", "failed to create API key for gateway")
	errDeleteAPIKey = errors.DefineAborted("delete_api_key", "delete API key")
)

func (u *Upstream) claimLBSCUPSGateway(
	ctx context.Context, ids *ttnpb.GatewayIdentifiers, ownerToken, clusterAddress string,
) (*dcstypes.GatewayMetadata, error) {
	logger := log.FromContext(ctx)
	eui := types.MustEUI64(ids.Eui).OrZero()

	// Create CUPS and LNS API keys for the gateway. The CUPS key will be used as gateway token when claiming on TTGC and
	// the LNS key will be returned in the metadata. The caller is responsible for updating the LNS key in the gateway.
	cupsKey, lnsKey, err := u.createAPIKeys(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Claim the gateway on TTGC with the CUPS key as the gateway token.
	gtwClient := northboundv1.NewGatewayServiceClient(u.client)
	_, err = gtwClient.Claim(ctx, &northboundv1.GatewayServiceClaimRequest{
		GatewayId:    eui.MarshalNumber(),
		Domain:       u.client.Domain(ctx),
		OwnerToken:   ownerToken,
		GatewayToken: []byte(cupsKey.Key),
	})
	if err != nil {
		logger.WithError(err).Warn("Failed to claim gateway on TTGC")
		return nil, err
	}

	// Get the Root CA from the Gateway Server.
	host, _, err := net.SplitHostPort(clusterAddress)
	if err != nil {
		host = clusterAddress
	}
	clusterAddress = net.JoinHostPort(host, "8889")
	rootCA, err := u.getRootCA(ctx, clusterAddress)
	if err != nil {
		return nil, err
	}

	var (
		loraPFProfileID []byte
		loraPFProfile   = &northboundv1.LoraPacketForwarderProfile{
			ProfileName: clusterAddress,
			Shared:      false,
			Protocol:    northboundv1.LoraPacketForwarderProtocol_LORA_PACKET_FORWARDER_PROTOCOL_BASIC_STATION,
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
			return nil, err
		}
		res, err := loraPFProfileClient.Create(ctx, &northboundv1.LoraPacketForwarderProfileServiceCreateRequest{
			Domain:                     u.client.Domain(ctx),
			Group:                      profileGroup,
			LoraPacketForwarderProfile: loraPFProfile,
		})
		if err != nil {
			logger.WithError(err).Warn("Failed to create LoRa Packet Forwarder profile")
			return nil, err
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
				return nil, err
			}
		}
		loraPFProfileID = loraPFGetRes.ProfileId
	}

	// Update the gateway with the Lora Packet Forwarder profile.
	_, err = gtwClient.Update(ctx, &northboundv1.GatewayServiceUpdateRequest{
		GatewayId: eui.MarshalNumber(),
		Domain:    u.client.Domain(ctx),
		LoraPacketForwarderProfileId: &northboundv1.ProfileIDValue{
			Value: loraPFProfileID,
		},
	})
	if err != nil {
		logger.WithError(err).Warn("Failed to update gateway with profiles")
		return nil, err
	}

	return &dcstypes.GatewayMetadata{
		LBSLNSKey: lnsKey,
	}, nil
}

// createAPIKeys creates the CUPS and LNS API keys for the gateway.
func (u *Upstream) createAPIKeys(
	ctx context.Context, ids *ttnpb.GatewayIdentifiers,
) (cupsKey, lnsKey *ttnpb.APIKey, err error) {
	logger := log.FromContext(ctx)

	gatewayAccess, err := u.getGatewayAccess(ctx)
	if err != nil {
		return nil, nil, err
	}

	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, u.AllowInsecureForCredentials())
	if err != nil {
		return nil, nil, err
	}

	cupsKey, err = gatewayAccess.CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
		GatewayIds: ids,
		Name:       fmt.Sprintf("LBS CUPS Key (TTGC), %s", time.Now().UTC().Format(time.RFC3339)),
		Rights: []ttnpb.Right{
			ttnpb.Right_RIGHT_GATEWAY_INFO,
			ttnpb.Right_RIGHT_GATEWAY_SETTINGS_BASIC,
			ttnpb.Right_RIGHT_GATEWAY_READ_SECRETS,
		},
	}, callOpt)
	if err != nil {
		logger.WithError(err).Warn("Failed to create CUPS API key")
		return nil, nil, errCreateAPIKey.WithCause(err)
	}

	lnsKey, err = gatewayAccess.CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
		GatewayIds: ids,
		Name:       fmt.Sprintf("LBS LNS Key (TTGC), %s", time.Now().UTC().Format(time.RFC3339)),
		Rights: []ttnpb.Right{
			ttnpb.Right_RIGHT_GATEWAY_LINK,
		},
	}, callOpt)
	if err != nil {
		logger.WithError(err).Warn("Failed to create LNS API key")
		return nil, nil, errCreateAPIKey.WithCause(err)
	}

	return cupsKey, lnsKey, nil
}

func (u *Upstream) getGatewayAccess(ctx context.Context) (ttnpb.GatewayAccessClient, error) {
	if u.gatewayAccess != nil {
		return u.gatewayAccess, nil
	}
	conn, err := u.GetPeerConn(ctx, ttnpb.ClusterRole_ACCESS, nil)
	if err != nil {
		return nil, err
	}
	return ttnpb.NewGatewayAccessClient(conn), nil
}

// deleteAPIKeys deletes the CUPS and LNS API keys for the gateway.
func (u *Upstream) deleteAPIKeys(ctx context.Context, ids *ttnpb.GatewayIdentifiers) error {
	logger := log.FromContext(ctx)

	gatewayAccess, err := u.getGatewayAccess(ctx)
	if err != nil {
		return err
	}

	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, u.AllowInsecureForCredentials())
	if err != nil {
		return err
	}

	apiKeys, err := gatewayAccess.ListAPIKeys(ctx, &ttnpb.ListGatewayAPIKeysRequest{
		GatewayIds: ids,
	}, callOpt)
	if err != nil {
		logger.WithError(err).Warn("Failed to list API keys")
		return errDeleteAPIKey.WithCause(err)
	}

	// Delete the LBS CUPS and LBS LNS keys.
	for _, key := range apiKeys.ApiKeys {
		if key.Name == "" {
			continue
		}
		// Match keys created by this claimer.
		if len(key.Name) > 8 && (key.Name[:8] == "LBS CUPS" || key.Name[:7] == "LBS LNS") {
			_, err := gatewayAccess.DeleteAPIKey(ctx, &ttnpb.DeleteGatewayAPIKeyRequest{
				GatewayIds: ids,
				KeyId:      key.Id,
			}, callOpt)
			if err != nil {
				logger.WithError(err).WithField("key_id", key.Id).Warn("Failed to delete API key")
				// Continue deleting other keys.
			}
		}
	}

	return nil
}
