// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mac

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

func newDev() *ttnpb.EndDevice {
	return &ttnpb.EndDevice{
		LoRaWANVersion:    ttnpb.MAC_V1_1,
		LoRaWANPHYVersion: ttnpb.PHY_V1_1,
		FrequencyPlanID:   "US_902_928",
		MACState:          &ttnpb.MACState{},
	}
}

func newMsg() ttnpb.Message {
	return ttnpb.Message{
		MHDR: ttnpb.MHDR{
			MType: ttnpb.MType_CONFIRMED_UP,
			Major: ttnpb.Major_LORAWAN_R1,
		},
		Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{
			FPort: 1,
			FHDR:  ttnpb.FHDR{
			// FOpts are set in the individual tests
			},
		}},
	}
}

func newUplink() *ttnpb.UplinkMessage {
	return &ttnpb.UplinkMessage{
		Payload: newMsg(),
		Settings: ttnpb.TxSettings{
			Modulation:      ttnpb.Modulation_LORA,
			Bandwidth:       125,
			SpreadingFactor: 12,
		},
	}
}
