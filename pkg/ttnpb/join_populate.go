// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/types"
)

func NewPopulatedJoinRequest(r randyJoin, easy bool) *JoinRequest {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("NewPopulatedJoinRequest: %s", r)
		}
	}()

	out := &JoinRequest{}
	out.SelectedMacVersion = MACVersion(r.Intn(5))
	out.NetID = *types.NewPopulatedNetID(r)
	out.DownlinkSettings = *NewPopulatedDLSettings(r, easy)
	out.RxDelay = r.Uint32()
	if r.Intn(10) != 0 {
		out.CFList = NewPopulatedCFList(r, false)
	}

	msg := NewPopulatedMessageJoinRequest(r)
	out.Payload = *msg

	var err error
	out.RawPayload, err = msg.AppendLoRaWAN(out.RawPayload)
	if err != nil {
		panic(errors.NewWithCause(err, "failed to encode join request message to LoRaWAN"))
	}
	out.EndDeviceIdentifiers = *NewPopulatedEndDeviceIdentifiers(r, false)
	devEUI := msg.GetJoinRequestPayload().DevEUI
	out.EndDeviceIdentifiers.DevEUI = &devEUI
	return out
}
