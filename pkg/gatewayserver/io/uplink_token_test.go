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

package io_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestUplinkToken(t *testing.T) {
	a := assertions.New(t)

	expected := &ttnpb.DownlinkPath{
		GatewayAntennaIdentifiers: ttnpb.GatewayAntennaIdentifiers{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{
				GatewayID: "foo-gateway",
			},
			AntennaIndex: 0,
		},
		UplinkTimestamp: 12345678,
	}

	uplinkToken, err := io.PathToUplinkToken(expected)
	a.So(err, should.BeNil)

	actual, err := io.UplinkTokenToPath(uplinkToken)
	a.So(err, should.BeNil)
	a.So(actual, should.Resemble, expected)
}
