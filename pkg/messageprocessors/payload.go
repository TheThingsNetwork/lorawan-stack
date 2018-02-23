// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package messageprocessors

import (
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// PayloadEncoder represents a payload encoder message processor.
type PayloadEncoder interface {
	Encode(model *ttnpb.EndDeviceModel, parameter string, message *ttnpb.DownlinkMessage) (*ttnpb.DownlinkMessage, error)
}

// PayloadDecoder represents a payload decoder message processor.
type PayloadDecoder interface {
	Decode(model *ttnpb.EndDeviceModel, parameter string, message *ttnpb.UplinkMessage) (*ttnpb.UplinkMessage, error)
}

// PayloadEncodeDecoder is the interface that groups the Encode and Decode methods.
type PayloadEncodeDecoder interface {
	PayloadEncoder
	PayloadDecoder
}
