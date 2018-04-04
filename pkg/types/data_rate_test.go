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

package types_test

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestValidSpreadingFactors(t *testing.T) {
	a := assertions.New(t)

	dataRates := map[types.DataRate]uint8{
		{LoRa: "SF7BW125"}:  7,
		{LoRa: "SF8BW125"}:  8,
		{LoRa: "SF9BW125"}:  9,
		{LoRa: "SF10BW125"}: 10,
		{LoRa: "SF11BW125"}: 11,
		{LoRa: "SF12BW125"}: 12,
	}

	for dr, sf := range dataRates {
		spreadingFactor, err := dr.SpreadingFactor()
		a.So(err, should.BeNil)
		a.So(spreadingFactor, should.Equal, sf)
	}
}

func TestValidBandwidth(t *testing.T) {
	a := assertions.New(t)

	dataRates := map[types.DataRate]uint32{
		{LoRa: "SF7BW125"}: 125000,
		{LoRa: "SF8BW250"}: 250000,
		{LoRa: "SF9BW500"}: 500000,
	}

	for dr, bw := range dataRates {
		bandwidth, err := dr.Bandwidth()
		a.So(err, should.BeNil)
		a.So(bandwidth, should.Equal, bw)
	}
}

func TestInvalidSpreadingFactors(t *testing.T) {
	a := assertions.New(t)

	dataRates := []types.DataRate{
		{LoRa: "SF13BW125"},
		{LoRa: "SF2BW125"},
		{LoRa: "SF0BW125"},
		{LoRa: "SFUT"},
		{LoRa: "NOSF"},
		{FSK: 125},
	}

	for _, dr := range dataRates {
		_, err := dr.SpreadingFactor()
		a.So(err, should.NotBeNil)
	}
}

func TestInvalidBandwidth(t *testing.T) {
	a := assertions.New(t)

	dataRates := []types.DataRate{
		{LoRa: "SF13BW3"},
		{LoRa: "SF2BW"},
		{LoRa: "SF0BWNO"},
		{LoRa: "NO"},
	}

	for _, dr := range dataRates {
		_, err := dr.Bandwidth()
		a.So(err, should.NotBeNil)
	}
}
