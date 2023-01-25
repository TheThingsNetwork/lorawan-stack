// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

	qrcodegen "github.com/skip2/go-qrcode"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

var errUnauthenticated = errors.DefineUnauthenticated("unauthenticated", "call was not authenticated")

type endDeviceQRCodeGeneratorServer struct {
	ttnpb.UnimplementedEndDeviceQRCodeGeneratorServer

	QRG *QRCodeGenerator
}

// GetFormat implements EndDeviceQRCodeGenerator.
func (s *endDeviceQRCodeGeneratorServer) GetFormat(ctx context.Context, req *ttnpb.GetQRCodeFormatRequest) (*ttnpb.QRCodeFormat, error) {
	_, err := rpcmetadata.WithForwardedAuth(ctx, s.QRG.AllowInsecureForCredentials())
	if err != nil {
		return nil, err
	}
	format := s.QRG.endDevices.GetEndDeviceFormat(req.FormatId)
	if format == nil {
		return nil, errFormatNotFound.New()
	}
	return format.Format(), nil
}

// ListFormats implements EndDeviceQRCodeGenerator.
func (s *endDeviceQRCodeGeneratorServer) ListFormats(ctx context.Context, _ *emptypb.Empty) (*ttnpb.QRCodeFormats, error) {
	_, err := rpcmetadata.WithForwardedAuth(ctx, s.QRG.AllowInsecureForCredentials())
	if err != nil {
		return nil, err
	}
	res := &ttnpb.QRCodeFormats{
		Formats: make(map[string]*ttnpb.QRCodeFormat),
	}
	for k, f := range s.QRG.endDevices.GetEndDeviceFormats() {
		res.Formats[k] = f.Format()
	}
	return res, nil
}

// Generate implements EndDeviceQRCodeGenerator.
func (s *endDeviceQRCodeGeneratorServer) Generate(ctx context.Context, req *ttnpb.GenerateEndDeviceQRCodeRequest) (*ttnpb.GenerateQRCodeResponse, error) {
	_, err := rpcmetadata.WithForwardedAuth(ctx, s.QRG.AllowInsecureForCredentials())
	if err != nil {
		return nil, err
	}
	formatter := s.QRG.endDevices.GetEndDeviceFormat(req.FormatId)
	if formatter == nil {
		return nil, errFormatNotFound.New()
	}
	data := formatter.New()
	if err := data.Encode(req.EndDevice); err != nil {
		return nil, err
	}
	if err := data.Validate(); err != nil {
		return nil, err
	}
	text, err := data.MarshalText()
	if err != nil {
		return nil, err
	}
	res := &ttnpb.GenerateQRCodeResponse{
		Text: string(text),
	}
	if req.Image != nil {
		qr, err := qrcodegen.New(string(text), qrcodegen.Medium)
		if err != nil {
			return nil, err
		}
		data, err := qr.PNG(int(req.Image.ImageSize))
		if err != nil {
			return nil, err
		}
		res.Image = &ttnpb.Picture{
			Embedded: &ttnpb.Picture_Embedded{
				MimeType: "image/png",
				Data:     data,
			},
		}
	}
	return res, nil
}

// Parse implements EndDeviceQRCodeGenerator.
func (s *endDeviceQRCodeGeneratorServer) Parse(ctx context.Context, req *ttnpb.ParseEndDeviceQRCodeRequest) (*ttnpb.ParseEndDeviceQRCodeResponse, error) {
	_, err := rpcmetadata.WithForwardedAuth(ctx, s.QRG.AllowInsecureForCredentials())
	if err != nil {
		return nil, err
	}
	data, err := s.QRG.endDevices.Parse(req.FormatId, req.QrCode)
	if err != nil {
		return nil, err
	}
	return &ttnpb.ParseEndDeviceQRCodeResponse{
		FormatId:          data.FormatID(),
		EndDeviceTemplate: data.EndDeviceTemplate(),
	}, nil
}
