// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

// Package lora contains LoRa modulation utilities.
package lora

import (
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"

	pbtypes "github.com/gogo/protobuf/types"
)

// AdjustedRSSI returns the LoRa RSSI: the channel RSSI adjusted for SNR.
// Below -5 dB, the SNR is added to the channel RSSI.
// Between -5 dB and 10 dB, the SNR is scaled to 0 and added to the channel RSSI.
func AdjustedRSSI(channelRSSI, snr float32) float32 {
	rssi := channelRSSI
	if snr <= -5.0 {
		rssi += snr
	} else if snr < 10.0 {
		rssi += snr/3.0 - 10.0/3.0
	}
	return rssi
}

// ExtractRxMetadataReceivedAt returns the ReceivedAt timestamp based on the
// provided metadatas or nil if no md.GpsTime nor md.ReceivedAt is found.
func ExtractRxMetadataReceivedAt(mds []*ttnpb.RxMetadata) *pbtypes.Timestamp {
	var ts *pbtypes.Timestamp
	for _, md := range mds {
		if t := md.GpsTime; t != nil {
			return t
		}
		if ts == nil && md.ReceivedAt != nil {
			ts = md.ReceivedAt
		}
	}
	return ts
}

// ExtractUplinkReceivedAt returns the correct ReceivedAt timestamp based on
// the uplink message's RxMetadata and the msg.ReceivedAt.
func ExtractUplinkReceivedAt(msg *ttnpb.UplinkMessage) *pbtypes.Timestamp {
	mdReceivedAt := ExtractRxMetadataReceivedAt(msg.RxMetadata)
	if mdReceivedAt != nil {
		return mdReceivedAt
	}
	return msg.ReceivedAt
}

// ExtractApplicationUplinkReceivedAt returns the correct ReceivedAt timestamp based on
// the application uplink message's RxMetadata and the msg.ReceivedAt.
func ExtractApplicationUplinkReceivedAt(msg *ttnpb.ApplicationUplink) *pbtypes.Timestamp {
	mdReceivedAt := ExtractRxMetadataReceivedAt(msg.RxMetadata)
	if mdReceivedAt != nil {
		return mdReceivedAt
	}
	return msg.ReceivedAt
}
