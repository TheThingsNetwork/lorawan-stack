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

import (
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UplinkToken returns an uplink token from the given downlink path.
func UplinkToken(ids *ttnpb.GatewayAntennaIdentifiers, timestamp uint32, concentratorTime scheduling.ConcentratorTime, serverTime time.Time, gatewayTime *time.Time) ([]byte, error) {
	token := ttnpb.UplinkToken{
		Ids:              ids,
		Timestamp:        timestamp,
		ServerTime:       timestamppb.New(serverTime),
		ConcentratorTime: int64(concentratorTime),
		GatewayTime:      ttnpb.ProtoTime(gatewayTime),
	}
	return proto.Marshal(&token)
}

// MustUplinkToken returns an uplink token from the given downlink path.
// This function panics if an error occurs. Use UplinkToken to handle errors.
func MustUplinkToken(ids *ttnpb.GatewayAntennaIdentifiers, timestamp uint32, concentratorTime scheduling.ConcentratorTime, serverTime time.Time, gatewayTime *time.Time) []byte {
	token, err := UplinkToken(ids, timestamp, concentratorTime, serverTime, gatewayTime)
	if err != nil {
		panic(err)
	}
	return token
}

// ParseUplinkToken returns the downlink path from the given uplink token.
func ParseUplinkToken(buf []byte) (*ttnpb.UplinkToken, error) {
	var token ttnpb.UplinkToken
	if err := proto.Unmarshal(buf, &token); err != nil {
		return nil, err
	}
	if err := token.ValidateFields(); err != nil {
		return nil, err
	}
	return &token, nil
}
