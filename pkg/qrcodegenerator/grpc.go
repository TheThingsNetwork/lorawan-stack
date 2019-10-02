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

	pbtypes "github.com/gogo/protobuf/types"
	qrcodegen "github.com/skip2/go-qrcode"
	"go.thethings.network/lorawan-stack/pkg/qrcode"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type endDeviceQRCodeGeneratorServer struct {
	QRG *QRCodeGenerator
}

func (s *endDeviceQRCodeGeneratorServer) GetFormat(ctx context.Context, req *ttnpb.GetQRCodeFormatRequest) (*ttnpb.QRCodeFormat, error) {
	format := qrcode.GetEndDeviceFormat(req.FormatID)
	if format == nil {
		return nil, errFormatNotFound
	}
	return format.Format(), nil
}

func (s *endDeviceQRCodeGeneratorServer) ListFormats(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.QRCodeFormats, error) {
	res := &ttnpb.QRCodeFormats{
		Formats: make(map[string]*ttnpb.QRCodeFormat),
	}
	for k, f := range qrcode.GetEndDeviceFormats() {
		res.Formats[k] = f.Format()
	}
	return res, nil
}

func (s *endDeviceQRCodeGeneratorServer) Generate(ctx context.Context, req *ttnpb.GenerateEndDeviceQRCodeRequest) (*ttnpb.GenerateQRCodeResponse, error) {
	formatter := qrcode.GetEndDeviceFormat(req.FormatID)
	if formatter == nil {
		return nil, errFormatNotFound
	}
	data := formatter.New()
	if err := data.Encode(&req.EndDevice); err != nil {
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
