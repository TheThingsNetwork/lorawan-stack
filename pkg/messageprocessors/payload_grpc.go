// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package messageprocessors

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// PayloadEncoderRPC implements the DownlinkMessageProcessorServer using a payload encoder.
type PayloadEncoderRPC struct {
	PayloadEncoder
}

// Process implements the DownlinkMessageProcessorServer interface.
func (r *PayloadEncoderRPC) Process(ctx context.Context, req *ttnpb.ProcessDownlinkMessageRequest) (*ttnpb.ApplicationDownlink, error) {
	msg := &req.Message
	if err := r.Encode(ctx, req.EndDeviceIdentifiers, &req.EndDeviceVersionIDs, msg, req.Parameter); err != nil {
		return nil, err
	}
	return msg, nil
}

// PayloadDecoderRPC implements the UplinkMessageProcessorServer using a payload decoder.
type PayloadDecoderRPC struct {
	PayloadDecoder
}

// Process implements the UplinkMessageProcessorServer interface.
func (r *PayloadDecoderRPC) Process(ctx context.Context, req *ttnpb.ProcessUplinkMessageRequest) (*ttnpb.ApplicationUplink, error) {
	msg := &req.Message
	if err := r.Decode(ctx, req.EndDeviceIdentifiers, &req.EndDeviceVersionIDs, msg, req.Parameter); err != nil {
		return nil, err
	}
	return msg, nil
}
