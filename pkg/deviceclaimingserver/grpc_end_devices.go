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

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/enddevices"
	claimerrors "go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/enddevices/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	errParseQRCode = errors.Define("parse_qr_code", "parse QR code failed")
	errQRCodeData  = errors.DefineInvalidArgument("qr_code_data", "invalid QR code data")
	errNoJoinEUI   = errors.DefineInvalidArgument("no_join_eui", "extract JoinEUI from request")
	errNoEUIs      = errors.DefineFailedPrecondition(
		"no_euis",
		"DevEUI/JoinEUI not set for device",
	)
	errDeviceNotFound       = errors.DefineNotFound("device_not_found", "device not found")
	errClaimingNotSupported = errors.DefineAborted(
		"claiming_not_supported",
		"claiming not supported for JoinEUI `{eui}`",
	)
	errNoDevicesFound = errors.DefineInvalidArgument(
		"no_devices_found",
		"no devices in batch found in the device registry",
	)
)

// endDeviceClaimingServer is the front facing entity for gRPC requests.
type endDeviceClaimingServer struct {
	ttnpb.UnimplementedEndDeviceClaimingServerServer

	DCS *DeviceClaimingServer
}

// Claim implements EndDeviceClaimingServer.
func (edcs *endDeviceClaimingServer) Claim(
	ctx context.Context,
	req *ttnpb.ClaimEndDeviceRequest,
) (*ttnpb.EndDeviceIdentifiers, error) {
	// Check that the collaborator has necessary rights before attempting to claim it on an upstream.
	// Since this is part of the create device flow,
	// we check that the collaborator has the rights to create devices in the application.
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
		if err != nil {
			return nil, errQRCodeData.WithCause(err)
		}
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

	claimer := edcs.DCS.endDeviceClaimingUpstream.JoinEUIClaimer(ctx, joinEUI)
	if claimer == nil {
		return nil, errClaimingNotSupported.WithAttributes("eui", joinEUI)
	}

	err := claimer.Claim(ctx, joinEUI, devEUI, claimAuthenticationCode)
	if err != nil {
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
func (edcs *endDeviceClaimingServer) Unclaim(
	ctx context.Context,
	in *ttnpb.EndDeviceIdentifiers,
) (*emptypb.Empty, error) {
	devs, err := edcs.DCS.getEndDevices(
		ctx,
		in.GetApplicationIds(),
		[]string{in.GetDeviceId()},
		ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	)
	if err != nil {
		return nil, err
	}
	if len(devs.EndDevices) != 1 {
		return nil, errDeviceNotFound.New()
	}
	ids := devs.EndDevices[0].GetIds()
	if ids.JoinEui == nil || ids.DevEui == nil {
		return nil, errNoEUIs.WithAttributes("ids", ids)
	}

	joinEUI := types.MustEUI64(ids.JoinEui).OrZero()
	claimer := edcs.DCS.endDeviceClaimingUpstream.JoinEUIClaimer(ctx, joinEUI)
	if claimer == nil {
		return nil, errClaimingNotSupported.WithAttributes("eui", joinEUI)
	}
	if err := claimer.Unclaim(ctx, ids); err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

// GetInfoByJoinEUI implements EndDeviceClaimingServer.
func (edcs *endDeviceClaimingServer) GetInfoByJoinEUI(
	ctx context.Context,
	in *ttnpb.GetInfoByJoinEUIRequest,
) (*ttnpb.GetInfoByJoinEUIResponse, error) {
	joinEUI := types.MustEUI64(in.JoinEui).OrZero()
	claimer := edcs.DCS.endDeviceClaimingUpstream.JoinEUIClaimer(ctx, joinEUI)
	return &ttnpb.GetInfoByJoinEUIResponse{
		JoinEui:          joinEUI.Bytes(),
		SupportsClaiming: claimer != nil,
	}, nil
}

// GetClaimStatus implements EndDeviceClaimingServer.
func (edcs *endDeviceClaimingServer) GetClaimStatus(
	ctx context.Context,
	in *ttnpb.EndDeviceIdentifiers,
) (*ttnpb.GetClaimStatusResponse, error) {
	devs, err := edcs.DCS.getEndDevices(
		ctx,
		in.GetApplicationIds(),
		[]string{in.GetDeviceId()},
		ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	)
	if err != nil {
		return nil, err
	}
	if len(devs.EndDevices) != 1 {
		return nil, errDeviceNotFound.New()
	}
	ids := devs.EndDevices[0].GetIds()
	if ids.JoinEui == nil || ids.DevEui == nil {
		return nil, errNoEUIs.WithAttributes("ids", ids)
	}
	joinEUI := types.MustEUI64(ids.JoinEui).OrZero()
	claimer := edcs.DCS.endDeviceClaimingUpstream.JoinEUIClaimer(ctx, joinEUI)
	if claimer == nil {
		return nil, errClaimingNotSupported.WithAttributes("eui", joinEUI)
	}
	return claimer.GetClaimStatus(ctx, ids)
}

func (dcs *DeviceClaimingServer) getEndDevices(
	ctx context.Context,
	appID *ttnpb.ApplicationIdentifiers,
	deviceIDs []string,
	requiredRights ...ttnpb.Right,
) (*ttnpb.EndDevices, error) {
	if err := rights.RequireApplication(
		ctx,
		appID,
		requiredRights...,
	); err != nil {
		return nil, err
	}
	conn, err := dcs.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return nil, err
	}
	client := ttnpb.NewEndDeviceBatchRegistryClient(conn)

	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, dcs.AllowInsecureForCredentials())
	if err != nil {
		return nil, err
	}
	return client.Get(ctx, &ttnpb.BatchGetEndDevicesRequest{
		ApplicationIds: appID,
		DeviceIds:      deviceIDs,
		FieldMask:      ttnpb.FieldMask("ids"),
	}, callOpt)
}

// endDeviceBatchClaimingServer is the front facing entity for gRPC requests.
type endDeviceBatchClaimingServer struct {
	ttnpb.UnimplementedEndDeviceBatchClaimingServerServer

	DCS *DeviceClaimingServer
}

// GetInfoByJoinEUI implements EndDeviceClaimingServer.
func (srv *endDeviceBatchClaimingServer) GetInfoByJoinEUIs(
	ctx context.Context,
	in *ttnpb.GetInfoByJoinEUIsRequest,
) (*ttnpb.GetInfoByJoinEUIsResponse, error) {
	ret := &ttnpb.GetInfoByJoinEUIsResponse{
		Infos: make([]*ttnpb.GetInfoByJoinEUIResponse, 0, len(in.Requests)),
	}
	for _, req := range in.Requests {
		joinEUI := types.MustEUI64(req.JoinEui).OrZero()
		claimer := srv.DCS.endDeviceClaimingUpstream.JoinEUIClaimer(ctx, joinEUI)
		ret.Infos = append(ret.Infos, &ttnpb.GetInfoByJoinEUIResponse{
			JoinEui:          joinEUI.Bytes(),
			SupportsClaiming: claimer != nil,
		})
	}
	return ret, nil
}

// Unclaim implements EndDeviceBatchClaimingServer.
func (srv *endDeviceBatchClaimingServer) Unclaim(
	ctx context.Context,
	in *ttnpb.BatchUnclaimEndDevicesRequest,
) (*ttnpb.BatchUnclaimEndDevicesResponse, error) {
	ret := &ttnpb.BatchUnclaimEndDevicesResponse{
		ApplicationIds: in.GetApplicationIds(),
		Failed:         make(map[string]*ttnpb.ErrorDetails),
	}
	devs, err := srv.DCS.getEndDevices(
		ctx,
		in.GetApplicationIds(),
		in.GetDeviceIds(),
		ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	)
	if err != nil {
		return nil, err
	}
	if len(devs.EndDevices) == 0 {
		return nil, errNoDevicesFound.New()
	}

	// End Devices within an application can have different JoinEUIs.
	// We group them by the JoinEUI and claim them in batches.
	// We also keep track of EUIs to device IDs to report errors (if any).
	devEUIToDeviceID := make(map[types.EUI64]string)
	claimers := make(map[enddevices.EndDeviceClaimer][]*ttnpb.EndDeviceIdentifiers)
	for _, dev := range devs.EndDevices {
		ids := dev.GetIds()
		if ids.JoinEui == nil || ids.DevEui == nil {
			ret.Failed[ids.DeviceId] = ttnpb.ErrorDetailsToProto(
				errNoEUIs.New(),
			)
			continue
		}
		joinEUI := types.MustEUI64(ids.JoinEui).OrZero()
		claimer := srv.DCS.endDeviceClaimingUpstream.JoinEUIClaimer(ctx, joinEUI)
		if claimer == nil {
			ret.Failed[ids.DeviceId] = ttnpb.ErrorDetailsToProto(
				errClaimingNotSupported.WithAttributes("eui", joinEUI),
			)
			continue
		}
		devIDs := claimers[claimer]
		if devIDs == nil {
			devIDs = make([]*ttnpb.EndDeviceIdentifiers, 0)
		}
		devIDs = append(devIDs, ids)
		claimers[claimer] = devIDs
		devEUIToDeviceID[types.EUI64(ids.DevEui)] = ids.DeviceId
	}

	// Claim in batches of Join EUIs (claimers).
	for claimer, devIDs := range claimers {
		err := claimer.BatchUnclaim(ctx, devIDs)
		if err != nil {
			var errs claimerrors.DeviceErrors
			if !errors.As(err, &errs) {
				return nil, err
			}
			for eui, err := range errs.Errors {
				ret.Failed[devEUIToDeviceID[eui]] = ttnpb.ErrorDetailsToProto(err)
			}
		}
	}
	return ret, nil
}
