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

package test

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func gateway() *ttnpb.Gateway {
	return &ttnpb.Gateway{
		GatewayIdentifiers: ttnpb.GatewayIdentifiers{
			GatewayID: "test-gateway",
			EUI:       &types.EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42},
		},
		Description: "My description",
		Platform:    "Kerklink",
		Attributes: map[string]string{
			"foo": "bar",
		},
		PrivacySettings: ttnpb.GatewayPrivacySettings{
			LocationPublic: true,
		},
		Antennas: []ttnpb.GatewayAntenna{
			{
				Location: ttnpb.Location{
					Latitude:  11.11,
					Longitude: 22.22,
					Altitude:  10,
				},
			},
		},
	}
}

func TestShouldBeGateway(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeGateway(gateway(), gateway()), should.Equal, success)

	modified := gateway()
	modified.CreatedAt = time.Now()

	a.So(ShouldBeGateway(modified, gateway()), should.NotEqual, success)
}

func TestShouldBeGatewayIgnoringAutoFields(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeGatewayIgnoringAutoFields(gateway(), gateway()), should.Equal, success)

	modified := gateway()
	modified.Platform = "foo"
	modified.Attributes["foz"] = "baz"
	modified.Antennas = append(modified.Antennas, ttnpb.GatewayAntenna{
		Gain: 12.12,
	})

	a.So(ShouldBeGatewayIgnoringAutoFields(modified, gateway()), should.NotEqual, success)
}
