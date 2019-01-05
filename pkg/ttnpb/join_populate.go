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

package ttnpb

import (
	"fmt"

	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/types"
)

func NewPopulatedJoinRequest(r randyJoin, easy bool) *JoinRequest {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("NewPopulatedJoinRequest: %s", r)
		}
	}()

	out := &JoinRequest{}
	out.SelectedMACVersion = MACVersion(r.Intn(5))
	out.NetID = *types.NewPopulatedNetID(r)
	out.DownlinkSettings = *NewPopulatedDLSettings(r, easy)
	out.RxDelay = RxDelay(r.Uint32() % 16)
	if r.Intn(10) != 0 {
		out.CFList = NewPopulatedCFList(r, false)
	}

	msg := NewPopulatedMessageJoinRequest(r)
	out.Payload = msg

	var err error
	out.RawPayload, err = PopulatorConfig.LoRaWAN.AppendMessage(out.RawPayload, *msg)
	if err != nil {
		panic(fmt.Sprintf("failed to encode join-request message to LoRaWAN: %s", err))
	}
	out.EndDeviceIdentifiers = *NewPopulatedEndDeviceIdentifiers(r, false)
	devEUI := msg.GetJoinRequestPayload().DevEUI
	out.EndDeviceIdentifiers.DevEUI = &devEUI
	return out
}
