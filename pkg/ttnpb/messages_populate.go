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
	"math"

	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// NewPopulatedUplinkMessage is used for compatibility with gogoproto, and in cases, where the
// contents of the message are not important. It's advised to use one of:
// - NewPopulatedUplinkMessageUplink
// - NewPopulatedUplinkMessageJoinRequest
// - NewPopulatedUplinkMessageRejoinRequest
func NewPopulatedUplinkMessage(r randyMessages, easy bool) (msg *UplinkMessage) {
	switch r.Intn(3) {
	case 0:
		return NewPopulatedUplinkMessageUplink(r, *types.NewPopulatedAES128Key(r), *types.NewPopulatedAES128Key(r), r.Intn(2) == 1)
	case 1:
		return NewPopulatedUplinkMessageJoinRequest(r)
	case 2:
		return NewPopulatedUplinkMessageRejoinRequest(r, RejoinType(r.Intn(3)))
	}
	panic("unreachable")
}
func NewPopulatedUplinkMessageUplink(r randyLorawan, sNwkSIntKey, fNwkSIntKey types.AES128Key, confirmed bool) *UplinkMessage {
	out := &UplinkMessage{}
	out.Settings = *NewPopulatedTxSettings(r, false)
	out.RxMetadata = make([]*RxMetadata, 1+r.Intn(5))
	for i := 0; i < len(out.RxMetadata); i++ {
		out.RxMetadata[i] = NewPopulatedRxMetadata(r, false)
	}

	msg := NewPopulatedMessageUplink(r, sNwkSIntKey, fNwkSIntKey, uint8(out.Settings.DataRateIndex), uint8(out.GatewayChannelIndex), confirmed)
	out.Payload = msg
	var err error
	out.RawPayload, err = PopulatorConfig.LoRaWAN.AppendMessage(out.RawPayload, *msg)
	if err != nil {
		panic(fmt.Sprintf("could not encode raw payload to LoRaWAN: %s", err))
	}
	return out
}

func NewPopulatedUplinkMessageJoinRequest(r randyLorawan) *UplinkMessage {
	out := &UplinkMessage{}
	out.Settings = *NewPopulatedTxSettings(r, false)
	out.RxMetadata = make([]*RxMetadata, 1+r.Intn(5))
	for i := 0; i < len(out.RxMetadata); i++ {
		out.RxMetadata[i] = NewPopulatedRxMetadata(r, false)
	}

	msg := NewPopulatedMessageJoinRequest(r)
	out.Payload = msg
	var err error
	out.RawPayload, err = PopulatorConfig.LoRaWAN.AppendMessage(out.RawPayload, *msg)
	if err != nil {
		panic(fmt.Sprintf("failed to encode uplink message to LoRaWAN: %s", err))
	}
	return out
}

func NewPopulatedUplinkMessageRejoinRequest(r randyLorawan, typ RejoinType) *UplinkMessage {
	out := &UplinkMessage{}
	out.Settings = *NewPopulatedTxSettings(r, false)
	out.RxMetadata = make([]*RxMetadata, 1+r.Intn(5))
	for i := 0; i < len(out.RxMetadata); i++ {
		out.RxMetadata[i] = NewPopulatedRxMetadata(r, false)
	}

	msg := NewPopulatedMessageRejoinRequest(r, typ)
	out.Payload = msg
	var err error
	out.RawPayload, err = PopulatorConfig.LoRaWAN.AppendMessage(out.RawPayload, *msg)
	if err != nil {
		panic(fmt.Sprintf("failed to encode uplink message to LoRaWAN: %s", err))
	}
	return out
}

func NewPopulatedDownlinkMessage(r randyMessages, easy bool) *DownlinkMessage {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("NewPopulatedDownlinkMessage: %s", r)
		}
	}()

	out := &DownlinkMessage{}
	if r.Intn(2) == 0 {
		out.Settings = &DownlinkMessage_Request{
			Request: NewPopulatedTxRequest(r, false),
		}
	} else {
		out.Settings = &DownlinkMessage_Scheduled{
			Scheduled: NewPopulatedTxSettings(r, false),
		}
	}

	msg := NewPopulatedMessageDownlink(r, *types.NewPopulatedAES128Key(r), r.Intn(2) == 1)
	out.Payload = msg

	var err error
	out.RawPayload, err = PopulatorConfig.LoRaWAN.AppendMessage(out.RawPayload, *msg)
	if err != nil {
		panic(fmt.Sprintf("failed to encode downlink message to LoRaWAN: %s", err))
	}
	return out
}

func NewPopulatedApplicationDownlink(r randyMessages, _ bool) *ApplicationDownlink {
	out := &ApplicationDownlink{}
	out.FPort = 1 + uint32(r.Intn(222))
	out.FCnt = r.Uint32() % math.MaxUint16
	out.FRMPayload = make([]byte, r.Intn(255))
	for i := range out.FRMPayload {
		out.FRMPayload[i] = byte(r.Intn(256))
	}
	return out
}
