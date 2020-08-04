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

package io_test

import (
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestUplinkToken(t *testing.T) {
	a := assertions.New(t)

	ids := ttnpb.GatewayAntennaIdentifiers{
		GatewayIdentifiers: ttnpb.GatewayIdentifiers{
			GatewayID: "foo-gateway",
		},
		AntennaIndex: 0,
	}
	timestamp := uint32(12345678)
	concentratorTime := scheduling.ConcentratorTime(12345678000)
	serverTime := time.Now()

	uplinkToken, err := io.UplinkToken(ids, timestamp, concentratorTime, serverTime)
	a.So(err, should.BeNil)

	token, err := io.ParseUplinkToken(uplinkToken)
	a.So(err, should.BeNil)
	a.So(token.GatewayAntennaIdentifiers, should.Resemble, ids)
	a.So(token.Timestamp, should.Equal, timestamp)
}
