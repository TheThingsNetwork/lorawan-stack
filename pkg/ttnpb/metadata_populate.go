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
	pbtypes "github.com/gogo/protobuf/types"
)

func NewPopulatedRxMetadata(r randyMetadata, easy bool) *RxMetadata {
	this := &RxMetadata{}
	v1 := NewPopulatedGatewayIdentifiers(r, easy)
	this.GatewayIdentifiers = *v1
	this.AntennaIndex = r.Uint32()
	if r.Intn(10) != 0 {
		this.Time = pbtypes.NewPopulatedStdTime(r, easy)
	}
	this.Timestamp = r.Uint32()
	this.FineTimestamp = uint64(r.Uint32())
	v2 := r.Intn(100)
	this.EncryptedFineTimestamp = make([]byte, v2)
	for i := 0; i < v2; i++ {
		this.EncryptedFineTimestamp[i] = byte(r.Intn(256))
	}
	this.EncryptedFineTimestampKeyID = randStringMetadata(r)
	this.RSSI = float32(r.Float32())
	if r.Intn(2) == 0 {
		this.RSSI *= -1
	}
	this.ChannelRSSI = float32(r.Float32())
	if r.Intn(2) == 0 {
		this.ChannelRSSI *= -1
	}
	this.RSSIStandardDeviation = float32(r.Float32())
	if r.Intn(2) == 0 {
		this.RSSIStandardDeviation *= -1
	}
	this.SNR = float32(r.Float32())
	if r.Intn(2) == 0 {
		this.SNR *= -1
	}
	this.FrequencyOffset = r.Int63()
	if r.Intn(2) == 0 {
		this.FrequencyOffset *= -1
	}
	if r.Intn(10) != 0 {
		this.Location = NewPopulatedLocation(r, easy)
	}
	this.DownlinkPathConstraint = DownlinkPathConstraint([]int32{0, 1, 2}[r.Intn(3)])
	v3 := r.Intn(100)
	this.UplinkToken = make([]byte, v3)
	for i := 0; i < v3; i++ {
		this.UplinkToken[i] = byte(r.Intn(256))
	}
	if r.Intn(2) == 0 {
		this.SignalRSSI = &pbtypes.FloatValue{
			Value: -r.Float32(),
		}
	}
	this.ChannelIndex = uint32(r.Intn(256))
	if r.Intn(10) != 0 {
		this.Advanced = pbtypes.NewPopulatedStruct(r, easy)
	}
	if !easy && r.Intn(10) != 0 {
	}
	return this
}
