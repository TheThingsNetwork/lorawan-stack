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

package qrcodegenerator

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type gatewayQRCodeGeneratorServer struct {
	ttnpb.UnimplementedGatewayQRCodeGeneratorServer

	QRG *QRCodeGenerator
}

// GetFormat implements EndDeviceQRCodeGenerator.
func (s *gatewayQRCodeGeneratorServer) GetFormat(
	ctx context.Context,
	req *ttnpb.GetQRCodeFormatRequest,
) (*ttnpb.QRCodeFormat, error) {
	_, err := rpcmetadata.WithForwardedAuth(ctx, s.QRG.AllowInsecureForCredentials())
	if err != nil {
		return nil, err
	}
	format := s.QRG.gateways.GetGatewayFormat(req.FormatId)
	if format == nil {
		return nil, errFormatNotFound.New()
	}
	return format.Format(), nil
}

// Parse implements EndDeviceQRCodeGenerator.
func (s *gatewayQRCodeGeneratorServer) Parse(
	ctx context.Context,
	req *ttnpb.ParseGatewayQRCodeRequest,
) (*ttnpb.ParseGatewayQRCodeResponse, error) {
	_, err := rpcmetadata.WithForwardedAuth(ctx, s.QRG.AllowInsecureForCredentials())
	if err != nil {
		return nil, err
	}

	data, err := s.QRG.gateways.Parse(req.FormatId, req.QrCode)
	if err != nil {
		return nil, err
	}

	return &ttnpb.ParseGatewayQRCodeResponse{
		FormatId: data.FormatID(),
		ClaimGatewayRequest: &ttnpb.ClaimGatewayRequest{
			SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
				AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
					GatewayEui:         data.GatewayEUI().Bytes(),
					AuthenticationCode: []byte(data.OwnerToken()),
				},
			},
		},
	}, nil
}
