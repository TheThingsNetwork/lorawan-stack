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

// PayloadEncoder represents a payload encoder message processor.
type PayloadEncoder interface {
	Encode(ctx context.Context, message *ttnpb.DownlinkMessage, model *ttnpb.EndDeviceVersion, parameter string) (*ttnpb.DownlinkMessage, error)
}

// PayloadDecoder represents a payload decoder message processor.
type PayloadDecoder interface {
	Decode(ctx context.Context, message *ttnpb.UplinkMessage, model *ttnpb.EndDeviceVersion, parameter string) (*ttnpb.UplinkMessage, error)
}

// PayloadEncodeDecoder is the interface that groups the Encode and Decode methods.
type PayloadEncodeDecoder interface {
	PayloadEncoder
	PayloadDecoder
}
