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
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
)

var errMethodUnavailable = errors.DefineUnimplemented("method_unavailable", "method available")

// Fallback defines methods for the fallback server.
// TODO: Remove this interface (https://github.com/TheThingsIndustries/lorawan-stack/issues/3036).
type Fallback interface {
	web.Registerer
	Claim(ctx context.Context, req *ttnpb.ClaimEndDeviceRequest) (ids *ttnpb.EndDeviceIdentifiers, err error)
	AuthorizeApplication(context.Context, *ttnpb.AuthorizeApplicationRequest) (*pbtypes.Empty, error)
	UnauthorizeApplication(context.Context, *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error)
}

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
func (srv *endDeviceClaimingServer) Claim(ctx context.Context, req *ttnpb.ClaimEndDeviceRequest) (ids *ttnpb.EndDeviceIdentifiers, err error) {
	for _, edcs := range srv.DCS.endDeviceClaimingUpstreams {
		var joinEUI types.EUI64
		if authenticatedIDs := req.GetAuthenticatedIdentifiers(); authenticatedIDs != nil {
			joinEUI = req.GetAuthenticatedIdentifiers().JoinEui
		} else if qrCode := req.GetQrCode(); qrCode != nil {
			conn, err := srv.DCS.GetPeerConn(ctx, ttnpb.ClusterRole_QR_CODE_GENERATOR, nil)
			if err != nil {
				return nil, err
			}
			client := ttnpb.NewEndDeviceQRCodeGeneratorClient(conn)
			data, err := client.Parse(ctx, &ttnpb.ParseEndDeviceQRCodeRequest{
				QrCode: qrCode,
			})
			if err != nil {
				return nil, errParseQRCode.WithCause(err)
			}
			if edTemplate := data.GetEndDeviceTempate(); edTemplate != nil && edTemplate.GetEndDevice().Ids != nil && edTemplate.GetEndDevice().Ids.JoinEui != nil {
				joinEUI = *edTemplate.GetEndDevice().Ids.JoinEui
			} else {
				return nil, errQRCodeData.New()
			}
		} else {
			return nil, errNoJoinEUI.New()
		}
		if edcs.SupportsJoinEUI(joinEUI) {
			return edcs.Claim(ctx, req)
		}
		return nil, err
	}
	// Use default if no EDCS supports this EUI.
	// TODO: Remove this option and return JoinEUI not provisioned error (https://github.com/TheThingsIndustries/lorawan-stack/issues/3036).
	return srv.DCS.endDeviceClaimingUpstreams[defaultType].Claim(ctx, req)
}

// AuthorizeApplication implements EndDeviceClaimingServer.
func (srv *endDeviceClaimingServer) AuthorizeApplication(ctx context.Context, req *ttnpb.AuthorizeApplicationRequest) (*pbtypes.Empty, error) {
	return srv.DCS.endDeviceClaimingUpstreams[defaultType].AuthorizeApplication(ctx, req)
}

// UnauthorizeApplication implements EndDeviceClaimingServer.
func (srv *endDeviceClaimingServer) UnauthorizeApplication(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	return srv.DCS.endDeviceClaimingUpstreams[defaultType].UnauthorizeApplication(ctx, ids)
}
