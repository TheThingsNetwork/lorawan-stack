// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package messageprocessors

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// PayloadEncoder represents a payload encoder message processor.
type PayloadEncoder interface {
	Encode(ctx context.Context, message *ttnpb.DownlinkMessage, model *ttnpb.EndDeviceModel, parameter string) (*ttnpb.DownlinkMessage, error)
}

// PayloadDecoder represents a payload decoder message processor.
type PayloadDecoder interface {
	Decode(ctx context.Context, message *ttnpb.UplinkMessage, model *ttnpb.EndDeviceModel, parameter string) (*ttnpb.UplinkMessage, error)
}

// PayloadEncodeDecoder is the interface that groups the Encode and Decode methods.
type PayloadEncodeDecoder interface {
	PayloadEncoder
	PayloadDecoder
}
