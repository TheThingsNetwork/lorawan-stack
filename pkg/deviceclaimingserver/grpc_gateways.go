// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package deviceclaimingserver

import (
	"context"
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/gateways"
	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/observability"
	gtwregistry "go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/registry/gateways"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

type peerAccess interface {
	AllowInsecureForCredentials() bool
}

// gatewayClaimingServer is the front facing entity for gRPC requests.
type gatewayClaimingServer struct {
	ttnpb.UnimplementedGatewayClaimingServerServer

	peerAccess

	upstream *gateways.Upstream
	registry gtwregistry.GatewayRegistry
}

var (
	errGatewayClaimingWithQRCode = errors.DefineUnimplemented(
		"gateway_claiming_with_qrcodes_not_implemented",
		"gateway claiming with QR codes not implemented",
	)
	errGatewayAlreadyExists = errors.DefineAlreadyExists(
		"gateway_already_exists",
		"gateway with EUI `{eui}` already exists",
	)
	errGatewayClaimingNotSupported = errors.DefineAborted(
		"gateway_claiming_not_supported",
		"claiming not supported for gateway with EUI `{eui}`",
	)
	errClaim = errors.DefineAborted(
		"claim gateway",
		"claim gateway",
	)
	errCreateGateway = errors.DefineAborted(
		"create_gateway",
		"create gateway",
	)
	errNoEUI = errors.DefineInvalidArgument(
		"no_eui",
		"no EUI found for gateway",
	)
	errNoGatewayServerAddress = errors.DefineInvalidArgument(
		"no_gateway_server_address",
		"no gateway server address set for gateway",
	)
)

// Claim implements GatewayClaimingServer.
func (gcls *gatewayClaimingServer) Claim(
	ctx context.Context,
	req *ttnpb.ClaimGatewayRequest,
) (ids *ttnpb.GatewayIdentifiers, retErr error) {
	logger := log.FromContext(ctx)

	// Extract the EUI and the owner token (claim authentication code) from the request.
	var (
		authCode   []byte
		gatewayEUI types.EUI64
	)
	switch claim := req.SourceGateway.(type) {
	case *ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_:
		authIDs := claim.AuthenticatedIdentifiers
		gatewayEUI, authCode = types.MustEUI64(authIDs.GatewayEui).OrZero(), authIDs.AuthenticationCode
	case *ttnpb.ClaimGatewayRequest_QrCode:
		return nil, errGatewayClaimingWithQRCode.New()
	default:
		panic(fmt.Sprintf("proto: unexpected type %T", claim))
	}
	logger = logger.WithFields(log.Fields(
		"gateway_eui", gatewayEUI,
	))
	ids = &ttnpb.GatewayIdentifiers{
		Eui:       gatewayEUI.Bytes(),
		GatewayId: req.TargetGatewayId,
	}

	// Check if the gateway already exists.
	_, err := gcls.registry.GetIdentifiersForEUI(ctx, gatewayEUI)
	if err == nil {
		return nil, errGatewayAlreadyExists.WithAttributes("eui", gatewayEUI)
	} else if !errors.IsNotFound(err) {
		return nil, err
	}

	// Support clients that only set a single frequency plan.
	if len(req.TargetFrequencyPlanIds) == 0 && req.TargetFrequencyPlanId != "" { // nolint:staticcheck
		req.TargetFrequencyPlanIds = []string{req.TargetFrequencyPlanId} // nolint:staticcheck
	}

	// Check if the gateway is configured for claiming.
	claimer := gcls.upstream.Claimer(gatewayEUI)
	if claimer == nil {
		return nil, errGatewayClaimingNotSupported.WithAttributes("eui", gatewayEUI)
	}

	// Claim the gateway on the upstream.
	if err := claimer.Claim(ctx, gatewayEUI, string(authCode), req.TargetGatewayServerAddress); err != nil {
		observability.RegisterFailClaim(ctx, ids.GetEntityIdentifiers(), err)
		return nil, errClaim.WithCause(err)
	}

	// Unclaim if creation fails.
	defer func(ids *ttnpb.GatewayIdentifiers) {
		if retErr != nil {
			observability.RegisterAbortClaim(ctx, ids.GetEntityIdentifiers(), retErr)
			if err := claimer.Unclaim(ctx, gatewayEUI); err != nil {
				logger.WithError(err).Warn("Failed to unclaim gateway")
			}
			return
		}
		observability.RegisterSuccessClaim(ctx, ids.GetEntityIdentifiers())
	}(ids)

	// Create the gateway in the IS.
	gateway := &ttnpb.Gateway{
		Ids:                            ids,
		GatewayServerAddress:           req.TargetGatewayServerAddress,
		EnforceDutyCycle:               true,
		RequireAuthenticatedConnection: true,
		FrequencyPlanIds:               req.TargetFrequencyPlanIds,
	}

	_, err = gcls.registry.Create(ctx, &ttnpb.CreateGatewayRequest{
		Gateway:      gateway,
		Collaborator: req.GetCollaborator(),
	})
	if err != nil {
		return nil, errCreateGateway.WithCause(err)
	}

	return ids, nil
}

// GetInfoByGatewayEUI implements GatewayClaimingServer.
func (gcls gatewayClaimingServer) GetInfoByGatewayEUI(
	ctx context.Context, in *ttnpb.GetInfoByGatewayEUIRequest,
) (*ttnpb.GetInfoByGatewayEUIResponse, error) {
	// Check that there's any auth token on the request context.
	_, err := rpcmetadata.WithForwardedAuth(ctx, gcls.AllowInsecureForCredentials())
	if err != nil {
		return nil, err
	}
	eui := types.MustEUI64(in.Eui).OrZero()

	return &ttnpb.GetInfoByGatewayEUIResponse{
		Eui:              in.Eui,
		SupportsClaiming: gcls.upstream.Claimer(eui) != nil,
	}, nil
}

// Unclaim implements GatewayClaimingServer.
func (gcls gatewayClaimingServer) Unclaim(ctx context.Context, req *ttnpb.GatewayIdentifiers) (*emptypb.Empty, error) {
	// Check for the necessary rights.
	if err := gcls.registry.AssertGatewayRights(
		ctx,
		&ttnpb.GatewayIdentifiers{
			GatewayId: req.GatewayId,
		},
		ttnpb.Right_RIGHT_GATEWAY_INFO,
		ttnpb.Right_RIGHT_GATEWAY_DELETE,
	); err != nil {
		return nil, err
	}

	// Get the gateway.
	gtw, err := gcls.registry.Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIds: req,
	})
	if err != nil {
		return nil, err
	}
	gatewayEUI := types.MustEUI64(gtw.Ids.Eui).OrZero()
	if gatewayEUI.IsZero() {
		return nil, errNoEUI.New()
	}
	if gtw.GatewayServerAddress == "" {
		return nil, errNoGatewayServerAddress.New()
	}
	claimer := gcls.upstream.Claimer(gatewayEUI)
	if claimer == nil {
		return nil, errGatewayClaimingNotSupported.WithAttributes("eui", gatewayEUI)
	}

	if err := claimer.Unclaim(ctx, gatewayEUI); err != nil {
		observability.RegisterFailUnclaim(ctx, gtw.GetEntityIdentifiers(), err)
		return nil, err
	}
	observability.RegisterSuccessUnclaim(ctx, gtw.GetEntityIdentifiers())

	return ttnpb.Empty, nil
}
