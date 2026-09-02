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
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/gateways"
	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/observability"
	gtwregistry "go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/registry/gateways"
	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/retry"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
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
	errUpdateGateway = errors.DefineAborted(
		"update_gateway",
		"update gateway",
	)
	errNoEUI = errors.DefineInvalidArgument(
		"no_eui",
		"no EUI found for gateway",
	)
	errFetchCreatedGateway = errors.DefineDeadlineExceeded(
		"fetch_created_gateway",
		"fetch gateway after creation",
	)
)

// claimCleanupTimeout bounds the compensating operations that revert a partially claimed gateway.
const claimCleanupTimeout = 10 * time.Second

// parseClaimRequest extracts the EUI and the owner token (claim authentication code) from the request.
func parseClaimRequest(req *ttnpb.ClaimGatewayRequest) (types.EUI64, []byte, error) {
	switch claim := req.SourceGateway.(type) {
	case *ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_:
		authIDs := claim.AuthenticatedIdentifiers
		return types.MustEUI64(authIDs.GatewayEui).OrZero(), authIDs.AuthenticationCode, nil
	case *ttnpb.ClaimGatewayRequest_QrCode:
		return types.EUI64{}, nil, errGatewayClaimingWithQRCode.New()
	default:
		panic(fmt.Sprintf("proto: unexpected type %T", claim))
	}
}

// Claim implements GatewayClaimingServer.
func (gcls *gatewayClaimingServer) Claim(
	ctx context.Context,
	req *ttnpb.ClaimGatewayRequest,
) (ids *ttnpb.GatewayIdentifiers, retErr error) {
	logger := log.FromContext(ctx)

	gatewayEUI, authCode, err := parseClaimRequest(req)
	if err != nil {
		return nil, err
	}
	logger = logger.WithFields(log.Fields(
		"gateway_eui", gatewayEUI,
	))
	gatewayID := req.TargetGatewayId
	if gatewayID == "" {
		gatewayID = strings.ToLower(gatewayEUI.String())
	}
	ids = &ttnpb.GatewayIdentifiers{
		Eui:       gatewayEUI.Bytes(),
		GatewayId: gatewayID,
	}

	// Check if the gateway already exists.
	_, err = gcls.registry.GetIdentifiersForEUI(ctx, gatewayEUI)
	if err == nil {
		return nil, errGatewayAlreadyExists.WithAttributes("eui", gatewayEUI)
	} else if !errors.IsNotFound(err) {
		return nil, err
	}

	// Create the gateway in the IS. The gateway is created before claiming on the upstream because the upstream
	// needs the gateway to exist in order to create API keys for it.
	gateway := &ttnpb.Gateway{
		Ids: ids,
	}

	created, err := gcls.registry.Create(ctx, &ttnpb.CreateGatewayRequest{
		Gateway:      gateway,
		Collaborator: req.GetCollaborator(),
	})
	if err != nil {
		return nil, errCreateGateway.WithCause(err)
	}
	if createdIDs := created.GetIds(); createdIDs != nil {
		ids = createdIDs
	}

	defer func(ids *ttnpb.GatewayIdentifiers) {
		cleanupCtx, cancelCleanup := context.WithTimeout(context.WithoutCancel(ctx), claimCleanupTimeout)
		defer cancelCleanup()

		if retErr != nil {
			logger.Warn("Failed to claim gateway, deleting created gateway")
			if _, delErr := gcls.registry.Purge(cleanupCtx, ids); delErr != nil {
				logger.WithError(delErr).Warn("Failed to delete created gateway after failed claim")
			}
		}
	}(ids)

	if err := gcls.waitForCreatedGateway(ctx, ids); err != nil {
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
	res, err := claimer.Claim(ctx, ids, string(authCode), req.TargetGatewayServerAddress)
	if err != nil {
		observability.RegisterFailClaim(ctx, ids.GetEntityIdentifiers(), err)
		return nil, errClaim.WithCause(err)
	}

	// Unclaim if update fails.
	defer func(ids *ttnpb.GatewayIdentifiers) {
		cleanupCtx, cancelCleanup := context.WithTimeout(context.WithoutCancel(ctx), claimCleanupTimeout)
		defer cancelCleanup()

		if retErr != nil {
			observability.RegisterAbortClaim(ctx, ids.GetEntityIdentifiers(), retErr)
			if err := claimer.Unclaim(cleanupCtx, ids); err != nil {
				logger.WithError(err).Warn("Failed to unclaim gateway")
			}
			return
		}
		observability.RegisterSuccessClaim(ctx, ids.GetEntityIdentifiers())
	}(ids)

	// Update the gateway in the IS. If the update fails, the gateway will be unclaimed in the above deferred function
	// and deleted in the previous one.
	gateway = &ttnpb.Gateway{
		Ids:                            ids,
		GatewayServerAddress:           req.TargetGatewayServerAddress,
		EnforceDutyCycle:               true,
		RequireAuthenticatedConnection: true,
		FrequencyPlanIds:               req.TargetFrequencyPlanIds,
		Antennas:                       res.Antennas,
	}

	fieldMask := &fieldmaskpb.FieldMask{
		Paths: []string{
			"gateway_server_address",
			"enforce_duty_cycle",
			"require_authenticated_connection",
			"frequency_plan_ids",
			"antennas",
		},
	}

	if res.LBSLNSKey != nil {
		gateway.LbsLnsSecret = &ttnpb.Secret{Value: []byte(res.LBSLNSKey.Key)}
		fieldMask.Paths = append(fieldMask.Paths, "lbs_lns_secret")
	}

	_, err = gcls.registry.Update(ctx, &ttnpb.UpdateGatewayRequest{
		Gateway:   gateway,
		FieldMask: fieldMask,
	})
	if err != nil {
		return nil, errUpdateGateway.WithCause(err)
	}

	return ids, nil
}

// waitForCreatedGateway waits until the created gateway is visible to the caller's credentials. Rights on the new
// gateway are computed against a possibly lagging IS read replica; a missing gateway surfaces as a not-found error
// for admin callers and is masked as a permission-denied error otherwise.
func (gcls *gatewayClaimingServer) waitForCreatedGateway(ctx context.Context, ids *ttnpb.GatewayIdentifiers) error {
	getCreatedGatewayTask := retry.Task{
		Name: "get created gateway",
		F: func() (bool, error) {
			_, err := gcls.registry.Get(ctx, &ttnpb.GetGatewayRequest{
				GatewayIds: ids,
				FieldMask:  ttnpb.FieldMask("ids"),
			})
			switch {
			case err == nil:
				return false, nil
			case errors.IsNotFound(err), errors.IsPermissionDenied(err):
				return true, err
			default:
				return false, err
			}
		},
		WaitTime:    500 * time.Millisecond,
		Jitter:      0.2,
		MaxAttempts: 5,
	}
	if err := getCreatedGatewayTask.Do(ctx); err != nil {
		return errFetchCreatedGateway.WithCause(err)
	}
	return nil
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
	var (
		eui              = types.MustEUI64(in.Eui).OrZero()
		claimer          = gcls.upstream.Claimer(eui)
		supportsClaiming = claimer != nil
		isManaged        bool
	)
	if supportsClaiming {
		var err error
		isManaged, err = claimer.IsManagedGateway(ctx, eui)
		if err != nil {
			return nil, err
		}
	}

	return &ttnpb.GetInfoByGatewayEUIResponse{
		Eui:              in.Eui,
		SupportsClaiming: supportsClaiming,
		IsManaged:        isManaged,
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
	claimer := gcls.upstream.Claimer(gatewayEUI)
	if claimer == nil {
		return nil, errGatewayClaimingNotSupported.WithAttributes("eui", gatewayEUI)
	}

	if err := claimer.Unclaim(ctx, gtw.Ids); err != nil {
		observability.RegisterFailUnclaim(ctx, gtw.GetEntityIdentifiers(), err)
		return nil, err
	}
	observability.RegisterSuccessUnclaim(ctx, gtw.GetEntityIdentifiers())

	return ttnpb.Empty, nil
}
