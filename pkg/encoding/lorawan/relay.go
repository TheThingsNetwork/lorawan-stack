// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package lorawan

import (
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/byteutil"
)

// MarshalRelayForwardDownlinkReq marshals a RelayForwardDownlinkReq.
func MarshalRelayForwardDownlinkReq(req *ttnpb.RelayForwardDownlinkReq) ([]byte, error) {
	if len(req.RawPayload) == 0 {
		return nil, errMissing("RawPayload")
	}
	return req.RawPayload, nil
}

// UnmarshalRelayForwardDownlinkReq unmarshals b into req.
func UnmarshalRelayForwardDownlinkReq(b []byte, req *ttnpb.RelayForwardDownlinkReq) error {
	if len(b) == 0 {
		return errMissing("RawPayload")
	}
	req.RawPayload = b
	return nil
}

// MarshalRelayForwardUplinkReq marshals a RelayForwardUplinkReq.
func MarshalRelayForwardUplinkReq(phy *band.Band, req *ttnpb.RelayForwardUplinkReq) ([]byte, error) {
	if len(req.RawPayload) == 0 {
		return nil, errMissing("RawPayload")
	}
	var uplinkMetadata uint32
	if req.WorChannel > 1 {
		return nil, errExpectedLowerOrEqual("WORChannel", 1)(req.WorChannel)
	}
	uplinkMetadata |= uint32(req.WorChannel&0x3) << 16
	if req.Rssi < -142 || req.Rssi > -15 {
		return nil, errExpectedBetween("RSSI", -142, -15)(req.Rssi)
	}
	uplinkMetadata |= uint32(-(req.Rssi+15)&0x7f) << 9
	if req.Snr < -20 || req.Snr > 11 {
		return nil, errExpectedBetween("SNR", -20, 11)(req.Snr)
	}
	uplinkMetadata |= uint32((req.Snr+20)&0x1f) << 4
	drIdx, _, found := phy.FindUplinkDataRate(req.DataRate)
	if !found {
		return nil, errMissing("DataRate")
	}
	uplinkMetadata |= uint32(drIdx & 0xf)
	b := make([]byte, 0, 6+len(req.RawPayload))
	b = byteutil.AppendUint32(b, uplinkMetadata, 3)
	b = byteutil.AppendUint64(b, req.Frequency/phy.FreqMultiplier, 3)
	if n := len(req.RawPayload); n == 0 {
		return nil, errExpectedLengthHigherOrEqual("RawPayload", 1)(n)
	}
	b = append(b, req.RawPayload...)
	return b, nil
}

// UnmarshalRelayForwardUplinkReq unmarshals b into req.
func UnmarshalRelayForwardUplinkReq(phy *band.Band, b []byte, req *ttnpb.RelayForwardUplinkReq) error {
	if n := len(b); n < 6 {
		return errExpectedLengthHigherOrEqual("RawPayload", 6)(n)
	}
	uplinkMetadata := byteutil.ParseUint32(b[:3])
	req.WorChannel = ttnpb.RelayWORChannel(uplinkMetadata>>16) & 0x3
	req.Rssi = -int32(uplinkMetadata>>9&0x7f) - 15
	req.Snr = int32(uplinkMetadata>>4&0x1f) - 20
	drIdx := ttnpb.DataRateIndex(uplinkMetadata & 0xf)
	dataRate, ok := phy.DataRates[drIdx]
	if !ok {
		return errMissing("DataRate")
	}
	req.DataRate = dataRate.Rate
	req.Frequency = byteutil.ParseUint64(b[3:6]) * phy.FreqMultiplier
	req.RawPayload = b[6:]
	return nil
}
