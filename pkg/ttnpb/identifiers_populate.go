// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"github.com/TheThingsNetwork/ttn/pkg/types"
)

const validIDChars = "abcdefghijklmnopqrstuvwxyz1234567890"

func NewPopulatedID(r randyIdentifiers) string {
	b := make([]byte, 2+r.Intn(35))
	for i := 0; i < len(b); i++ {
		b[i] = validIDChars[r.Intn(len(validIDChars))]
	}
	for n := 0; n < len(b)/8; n++ {
		i := 1 + r.Intn(len(b)-2)
		if b[i-1] != '_' && b[i-1] != '-' && b[i+1] != '_' && b[i+1] != '-' {
			b[i] = "-_"[r.Intn(2)]
		}
	}
	return string(b)
}

func NewPopulatedEndDeviceIdentifiers(r randyIdentifiers, easy bool) *EndDeviceIdentifiers {
	out := &EndDeviceIdentifiers{}
	if r.Intn(10) == 0 {
		out.DeviceID = NewPopulatedID(r)
	}
	if r.Intn(10) == 0 {
		out.ApplicationID = NewPopulatedID(r)
	}
	out.DevEUI = types.NewPopulatedEUI64(r)
	out.JoinEUI = types.NewPopulatedEUI64(r)
	out.DevAddr = types.NewPopulatedDevAddr(r)
	return out
}
