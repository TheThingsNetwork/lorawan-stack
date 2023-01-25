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

package devicetemplateconverter

import (
	"bytes"
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/devicetemplateconverter/profilefetcher"
	"go.thethings.network/lorawan-stack/v3/pkg/devicetemplates"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type endDeviceTemplateConverterServer struct {
	ttnpb.UnimplementedEndDeviceTemplateConverterServer

	DTC *DeviceTemplateConverter
}

// ListFormats implements ttnpb.DeviceTemplateServiceServer.
func (s *endDeviceTemplateConverterServer) ListFormats(
	context.Context,
	*emptypb.Empty,
) (*ttnpb.EndDeviceTemplateFormats, error) {
	formats := make(map[string]*ttnpb.EndDeviceTemplateFormat, len(s.DTC.converters))
	for id, converter := range s.DTC.converters {
		formats[id] = converter.Format()
	}
	return &ttnpb.EndDeviceTemplateFormats{
		Formats: formats,
	}, nil
}

// Convert implements ttnpb.DeviceTemplateServiceServer.
func (s *endDeviceTemplateConverterServer) Convert(
	req *ttnpb.ConvertEndDeviceTemplateRequest,
	res ttnpb.EndDeviceTemplateConverter_ConvertServer,
) error {
	converter, ok := s.DTC.converters[req.FormatId]
	if !ok {
		return errNotFound.WithAttributes("id", req.FormatId)
	}

	ctx := res.Context()
	ctx = devicetemplates.NewContextWithProfileIDs(ctx, req.GetEndDeviceVersionIds())
	ctx = profilefetcher.NewContextWithFetcher(ctx, profilefetcher.NewTemplateFetcher(s.DTC.Component))

	return converter.Convert(
		ctx,
		bytes.NewReader(req.Data),
		res.Send,
	)
}
