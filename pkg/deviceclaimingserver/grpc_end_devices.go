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
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/qrcode"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
)

var errMethodUnavailable = errors.DefineUnimplemented("method_unavailable", "method available")

// noopEDCS is a no-op EDCS.
type noopEDCS struct {
}

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

var (
	errParseQRCode = errors.Define("parse_qr_code", "parse QR code failed")
	errQRCodeData  = errors.DefineInvalidArgument("qr_code_data", "invalid QR code data")
	errNoJoinEUI   = errors.DefineInvalidArgument("no_join_eui", "failed to extract JoinEUI from request")
)

// Claim implements EndDeviceClaimingServer.
func (edcs *endDeviceClaimingServer) Claim(ctx context.Context, req *ttnpb.ClaimEndDeviceRequest) (ids *ttnpb.EndDeviceIdentifiers, err error) {
	for _, edcs := range edcs.DCS.endDeviceClaimingUpstreams {
		var joinEUI types.EUI64
		if authenticatedIDs := req.GetAuthenticatedIdentifiers(); authenticatedIDs != nil {
			joinEUI = req.GetAuthenticatedIdentifiers().JoinEui
		} else if qrCode := req.GetQrCode(); qrCode != nil {
			data, err := qrcode.Parse(qrCode)
			if err != nil {
				return nil, errParseQRCode.WithCause(err)
			}
			authIDs, ok := data.(qrcode.AuthenticatedEndDeviceIdentifiers)
			if !ok {
				return nil, errQRCodeData.New()
			}
			joinEUI, _, _ = authIDs.AuthenticatedEndDeviceIdentifiers()
		} else {
			return nil, errNoJoinEUI.New()
		}
		if edcs.SupportsJoinEUI(joinEUI) {
			return edcs.Claim(ctx, req)
		}
	}
	// Use default if no EDCS supports this EUI.
	// TODO: Remove this option and return JoinEUI not provisioned error (https://github.com/TheThingsIndustries/lorawan-stack/issues/3036).
	return edcs.DCS.endDeviceClaimingUpstreams[defaultType].Claim(ctx, req)
}

// AuthorizeApplication implements EndDeviceClaimingServer.
func (edcs *endDeviceClaimingServer) AuthorizeApplication(ctx context.Context, req *ttnpb.AuthorizeApplicationRequest) (*pbtypes.Empty, error) {
	return edcs.DCS.endDeviceClaimingUpstreams[defaultType].AuthorizeApplication(ctx, req)
}

// UnauthorizeApplication implements EndDeviceClaimingServer.
func (edcs *endDeviceClaimingServer) UnauthorizeApplication(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	return edcs.DCS.endDeviceClaimingUpstreams[defaultType].UnauthorizeApplication(ctx, ids)
}
