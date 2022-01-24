// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type qrCodeParserServer struct {
	QRG *QRCodeGenerator
}

// Parse implements QRCodeParserServer.
func (s *qrCodeParserServer) Parse(ctx context.Context, req *ttnpb.ParseQRCodeRequest) (*ttnpb.ParseQRCodeResponse, error) {
	data, err := s.QRG.qrCode.Parse(req.FormatId, req.QrCode)
	if err != nil {
		return nil, err
	}
	return &ttnpb.ParseQRCodeResponse{
		EntityOnboardingData: data.GetEntityOnboardingData(),
	}, nil
}
