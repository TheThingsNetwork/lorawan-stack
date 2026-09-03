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
	"strings"
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

const (
	// Prefix of the LNS API key name created by this claimer.
	lnsKeyNamePrefix = "LNS Key (TTGC)"
)

var (
	errCreateAPIKey = errors.DefineFailedPrecondition("create_api_key", "failed to create API key for gateway")
	errDeleteAPIKey = errors.DefineAborted("delete_api_key", "delete API key")
)

func (u *Upstream) claimLBSCUPSGateway(
	ctx context.Context, ids *ttnpb.GatewayIdentifiers, ownerToken, clusterAddress string,
) (*dcstypes.GatewayMetadata, error) {
	// Create the LNS API key for the gateway. It is sent as the gateway token when claiming on TTGC and returned in
	// the metadata; the caller is responsible for updating the LNS key in the gateway.
	lnsKey, err := u.createLNSAPIKey(ctx, ids)
	if err != nil {
		return nil, err
	}

	host, _, err := net.SplitHostPort(clusterAddress)
	if err != nil {
		host = clusterAddress
	}
	lnsAddress := net.JoinHostPort(host, u.lnsPort)

	if err := u.ClaimLBSGateway(ctx, ids, ownerToken, lnsKey.Key, lnsAddress); err != nil {
		return nil, err
	}

	return &dcstypes.GatewayMetadata{
		LBSLNSKey: lnsKey,
	}, nil
}

// ClaimLBSGateway claims a LoRa Basics Station gateway on TTGC and configures its LNS settings.
// The owner token is verified by TTGC against the owner token stored for the gateway.
// The LNS key is sent as the gateway token: TTGC serves it as the LNS authentication token in the CUPS response.
// It is a raw API key, without an authentication scheme prefix.
// The LNS address and its root CA are stored in a shared LoRa Packet Forwarder profile that is attached
// to the gateway.
func (u *Upstream) ClaimLBSGateway(
	ctx context.Context, ids *ttnpb.GatewayIdentifiers, ownerToken, lnsKey, lnsAddress string,
) error {
	logger := log.FromContext(ctx)
	eui := types.MustEUI64(ids.Eui).OrZero()

	// Claim the gateway on TTGC.
	gtwClient := northboundv1.NewGatewayServiceClient(u.client)
	_, err := gtwClient.Claim(ctx, &northboundv1.GatewayServiceClaimRequest{
		GatewayId:    eui.MarshalNumber(),
		Domain:       u.client.Domain(ctx),
		OwnerToken:   ownerToken,
		GatewayToken: []byte(lnsKey),
	})
	if err != nil {
		logger.WithError(err).Warn("Failed to claim gateway on TTGC")
		return err
	}

	// Get the Root CA from the Gateway Server.
	rootCA, err := u.getRootCA(ctx, lnsAddress)
	if err != nil {
		return err
	}

	var (
		loraPFProfileID []byte
		loraPFProfile   = &northboundv1.LoraPacketForwarderProfile{
			ProfileName: lnsAddress,
			Shared:      true,
			Protocol:    northboundv1.LoraPacketForwarderProtocol_LORA_PACKET_FORWARDER_PROTOCOL_BASIC_STATION,
			Address:     lnsAddress,
			RootCa:      rootCA.Raw,
		}
		loraPFProfileClient = northboundv1.NewLoraPacketForwarderProfileServiceClient(u.client)
	)
	loraPFGetRes, err := loraPFProfileClient.GetByName(
		ctx,
		&northboundv1.LoraPacketForwarderProfileServiceGetByNameRequest{
			Domain:      u.client.Domain(ctx),
			Group:       profileGroup,
			ProfileName: lnsAddress,
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
		return err
	}

	return nil
}

// createLNSAPIKey creates the LNS API key for the gateway.
func (u *Upstream) createLNSAPIKey(
	ctx context.Context, ids *ttnpb.GatewayIdentifiers,
) (*ttnpb.APIKey, error) {
	logger := log.FromContext(ctx)

	gatewayAccess, err := u.getGatewayAccess(ctx)
	if err != nil {
		return nil, err
	}

	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, u.AllowInsecureForCredentials())
	if err != nil {
		return nil, err
	}

	lnsKey, err := gatewayAccess.CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
		GatewayIds: ids,
		Name:       fmt.Sprintf("%s, %s", lnsKeyNamePrefix, time.Now().UTC().Format(time.RFC3339)),
		Rights: []ttnpb.Right{
			ttnpb.Right_RIGHT_GATEWAY_LINK,
		},
	}, callOpt)
	if err != nil {
		logger.WithError(err).Warn("Failed to create LNS API key")
		return nil, errCreateAPIKey.WithCause(err)
	}

	return lnsKey, nil
}

func (u *Upstream) getGatewayAccess(ctx context.Context) (ttnpb.GatewayAccessClient, error) {
	conn, err := u.GetPeerConn(ctx, ttnpb.ClusterRole_ACCESS, nil)
	if err != nil {
		return nil, err
	}
	return ttnpb.NewGatewayAccessClient(conn), nil
}

// deleteAPIKeys deletes the LNS API keys for the gateway.
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

	// Delete the LBS LNS keys.
	for _, key := range apiKeys.ApiKeys {
		// Match keys created by this claimer.
		if strings.HasPrefix(key.Name, lnsKeyNamePrefix) {
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
