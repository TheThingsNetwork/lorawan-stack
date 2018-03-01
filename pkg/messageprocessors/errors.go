// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package messageprocessors

import "github.com/TheThingsNetwork/ttn/pkg/errors"

// ErrNoMACPayload represents the ErrDescriptor of the error returned when the message does not
// contain MACPayload.
var ErrNoMACPayload = &errors.ErrDescriptor{
	MessageFormat: "Message does not contain MAC payload",
	Type:          errors.InvalidArgument,
	Code:          1,
}

func init() {
	ErrNoMACPayload.Register()
}
