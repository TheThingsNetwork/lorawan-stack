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

package ttnpb

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/types"
)

func NewPopulatedUplinkMessage(r randyMessages, easy bool) *UplinkMessage {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("NewPopulatedUplinkMessage: %s", r)
		}
	}()

	out := &UplinkMessage{}
	out.Settings = *NewPopulatedTxSettings(r, false)
	out.RxMetadata = make([]RxMetadata, 1+r.Intn(5))
	for i := 0; i < len(out.RxMetadata); i++ {
		out.RxMetadata[i] = *NewPopulatedRxMetadata(r, false)
	}

	msg := NewPopulatedMessageUplink(r, *types.NewPopulatedAES128Key(r), *types.NewPopulatedAES128Key(r), uint8(out.Settings.DataRateIndex), uint8(out.RxMetadata[0].ChannelIndex), r.Intn(2) == 1)
	out.Payload = *msg

	var err error
	out.RawPayload, err = msg.AppendLoRaWAN(out.RawPayload)
	if err != nil {
		panic(errors.NewWithCause(err, "failed to encode uplink message to LoRaWAN"))
	}
	out.EndDeviceIdentifiers = *NewPopulatedEndDeviceIdentifiers(r, false)
	devAddr := msg.GetMACPayload().DevAddr
	out.EndDeviceIdentifiers.DevAddr = &devAddr
	return out
}

func NewPopulatedDownlinkMessage(r randyMessages, easy bool) *DownlinkMessage {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("NewPopulatedDownlinkMessage: %s", r)
		}
	}()

	out := &DownlinkMessage{}
	out.Settings = *NewPopulatedTxSettings(r, false)
	out.TxMetadata = *NewPopulatedTxMetadata(r, false)

	msg := NewPopulatedMessageDownlink(r, *types.NewPopulatedAES128Key(r), r.Intn(2) == 1)
	out.Payload = *msg

	var err error
	out.RawPayload, err = msg.AppendLoRaWAN(out.RawPayload)
	if err != nil {
		panic(errors.NewWithCause(err, "failed to encode downlink message to LoRaWAN"))
	}
	out.EndDeviceIdentifiers = *NewPopulatedEndDeviceIdentifiers(r, false)
	devAddr := msg.GetMACPayload().DevAddr
	out.EndDeviceIdentifiers.DevAddr = &devAddr
	return out
}
