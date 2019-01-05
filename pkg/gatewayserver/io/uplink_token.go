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

package io

import "go.thethings.network/lorawan-stack/pkg/ttnpb"

// UplinkToken returns an uplink token from the given downlink path.
func UplinkToken(ids ttnpb.GatewayAntennaIdentifiers, timestamp uint32) ([]byte, error) {
	token := ttnpb.UplinkToken{
		GatewayAntennaIdentifiers: ids,
		Timestamp:                 timestamp,
	}
	return token.Marshal()
}

// MustUplinkToken returns an uplink token from the given downlink path.
// This function panics if an error occurs. Use UplinkToken to handle errors.
func MustUplinkToken(ids ttnpb.GatewayAntennaIdentifiers, timestamp uint32) []byte {
	token, err := UplinkToken(ids, timestamp)
	if err != nil {
		panic(err)
	}
	return token
}

// ParseUplinkToken returns the downlink path from the given uplink token.
func ParseUplinkToken(buf []byte) (ids ttnpb.GatewayAntennaIdentifiers, timestamp uint32, err error) {
	var token ttnpb.UplinkToken
	if err = token.Unmarshal(buf); err != nil {
		return
	}
	ids = token.GatewayAntennaIdentifiers
	timestamp = token.Timestamp
	return
}
