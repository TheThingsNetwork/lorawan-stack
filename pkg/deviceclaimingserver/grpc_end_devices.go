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

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
)

var (
	errParseQRCode       = errors.Define("parse_qr_code", "parse QR code failed")
	errQRCodeData        = errors.DefineInvalidArgument("qr_code_data", "invalid QR code data")
	errNoJoinEUI         = errors.DefineInvalidArgument("no_join_eui", "failed to extract JoinEUI from request")
	errMethodUnavailable = errors.DefineUnimplemented("method_unavailable", "method unavailable")
)

// Fallback defines methods for the fallback server.
// TODO: Remove this interface (https://github.com/TheThingsIndustries/lorawan-stack/issues/3036).
type Fallback interface {
	web.Registerer
	Claim(ctx context.Context, req *ttnpb.ClaimEndDeviceRequest) (ids *ttnpb.EndDeviceIdentifiers, err error)
	AuthorizeApplication(context.Context, *ttnpb.AuthorizeApplicationRequest) (*pbtypes.Empty, error)
	UnauthorizeApplication(context.Context, *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error)
}

// noopEDCS is a no-op EDCS.
type noopEDCS struct{}

// SupportsJoinEUI implements EndDeviceClaimingServer.
func (noopEDCS) SupportsJoinEUI(types.EUI64) bool {
	return false
}

// RegisterRoutes implements EndDeviceClaimingServer.
func (noopEDCS) RegisterRoutes(server *web.Server) {
}

// Claim implements EndDeviceClaimingServer.
func (noopEDCS) Claim(ctx context.Context, req *ttnpb.ClaimEndDeviceRequest) (ids *ttnpb.EndDeviceIdentifiers, err error) {
	return nil, errMethodUnavailable.New()
}

// Unclaim implements EndDeviceClaimingServer.
func (noopEDCS) Unclaim(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	return nil, errMethodUnavailable.New()
}

// GetInfoByJoinEUI implements EndDeviceClaimingServer.
func (noopEDCS) GetInfoByJoinEUI(ctx context.Context, in *ttnpb.GetInfoByJoinEUIRequest) (*ttnpb.GetInfoByJoinEUIResponse, error) {
	return nil, errMethodUnavailable.New()
}

// GetClaimStatus implements EndDeviceClaimingServer.
func (noopEDCS) GetClaimStatus(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*ttnpb.GetClaimStatusResponse, error) {
	return nil, errMethodUnavailable.New()
}

// AuthorizeApplication implements EndDeviceClaimingServer.
func (noopEDCS) AuthorizeApplication(ctx context.Context, req *ttnpb.AuthorizeApplicationRequest) (*pbtypes.Empty, error) {
	return nil, errMethodUnavailable.New()
}

// UnauthorizeApplication implements EndDeviceClaimingServer.
func (noopEDCS) UnauthorizeApplication(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	return nil, errMethodUnavailable.New()
}

// endDeviceClaimingServer is the front facing entity for gRPC requests.
type endDeviceClaimingServer struct {
	DCS *DeviceClaimingServer
}

// Claim implements EndDeviceClaimingServer.
func (edcs *endDeviceClaimingServer) Claim(ctx context.Context, req *ttnpb.ClaimEndDeviceRequest) (*ttnpb.EndDeviceIdentifiers, error) {
	// Check that the collaborator has necessary rights before attempting to claim it on an upstream.
	// Since this is part of the create device flow, we check that the collaborator has the rights to create devices in the application.
	targetAppID := req.GetTargetApplicationIds()
	if err := rights.RequireApplication(ctx, targetAppID,
		ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	); err != nil {
		return nil, err
	}

	var (
		joinEUI, devEUI         types.EUI64
		claimAuthenticationCode string
	)
	if authenticatedIDs := req.GetAuthenticatedIdentifiers(); authenticatedIDs != nil {
		joinEUI = types.MustEUI64(req.GetAuthenticatedIdentifiers().JoinEui).OrZero()
		devEUI = types.MustEUI64(req.GetAuthenticatedIdentifiers().DevEui).OrZero()
		claimAuthenticationCode = req.GetAuthenticatedIdentifiers().AuthenticationCode
	} else if qrCode := req.GetQrCode(); qrCode != nil {
		conn, err := edcs.DCS.GetPeerConn(ctx, ttnpb.ClusterRole_QR_CODE_GENERATOR, nil)
		if err != nil {
			return nil, err
		}
		qrg := ttnpb.NewEndDeviceQRCodeGeneratorClient(conn)
		callOpt, err := rpcmetadata.WithForwardedAuth(ctx, edcs.DCS.AllowInsecureForCredentials())
		if err != nil {
			return nil, err
		}
		data, err := qrg.Parse(ctx, &ttnpb.ParseEndDeviceQRCodeRequest{
			QrCode: qrCode,
		}, callOpt)
		dev := data.GetEndDeviceTemplate().GetEndDevice()
		if dev == nil {
			return nil, errParseQRCode.New()
		}
		joinEUI = types.MustEUI64(dev.GetIds().JoinEui).OrZero()
		devEUI = types.MustEUI64(dev.GetIds().DevEui).OrZero()
		claimAuthenticationCode = dev.ClaimAuthenticationCode.Value
	} else {
		return nil, errNoJoinEUI.New()
	}

	err := edcs.DCS.endDeviceClaimingUpstream.Claim(ctx, joinEUI, devEUI, claimAuthenticationCode)
	if err != nil {
		if errors.IsAborted(err) {
			log.FromContext(ctx).Warn("No upstream supports JoinEUI, use fallback")
			return edcs.DCS.endDeviceClaimingFallback.Claim(ctx, req)
		}
		return nil, err
	}

	// Echo identifiers from the request.
	return &ttnpb.EndDeviceIdentifiers{
		DeviceId:       req.TargetDeviceId,
		ApplicationIds: req.TargetApplicationIds,
		DevEui:         devEUI.Bytes(),
		JoinEui:        joinEUI.Bytes(),
	}, nil
}

// Unclaim implements EndDeviceClaimingServer.
func (edcs *endDeviceClaimingServer) Unclaim(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	return edcs.DCS.endDeviceClaimingUpstream.Unclaim(ctx, in)
}

// GetInfoByJoinEUI implements EndDeviceClaimingServer.
func (edcs *endDeviceClaimingServer) GetInfoByJoinEUI(ctx context.Context, in *ttnpb.GetInfoByJoinEUIRequest) (*ttnpb.GetInfoByJoinEUIResponse, error) {
	return edcs.DCS.endDeviceClaimingUpstream.GetInfoByJoinEUI(ctx, in)
}

// GetClaimStatus implements EndDeviceClaimingServer.
func (edcs *endDeviceClaimingServer) GetClaimStatus(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*ttnpb.GetClaimStatusResponse, error) {
	return edcs.DCS.endDeviceClaimingUpstream.GetClaimStatus(ctx, in)
}

// AuthorizeApplication implements EndDeviceClaimingServer.
func (edcs *endDeviceClaimingServer) AuthorizeApplication(ctx context.Context, req *ttnpb.AuthorizeApplicationRequest) (*pbtypes.Empty, error) {
	return edcs.DCS.endDeviceClaimingFallback.AuthorizeApplication(ctx, req)
}

// UnauthorizeApplication implements EndDeviceClaimingServer.
func (edcs *endDeviceClaimingServer) UnauthorizeApplication(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	return edcs.DCS.endDeviceClaimingFallback.UnauthorizeApplication(ctx, ids)
}
