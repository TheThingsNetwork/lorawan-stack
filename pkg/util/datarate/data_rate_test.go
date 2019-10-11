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

package datarate_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/datarate"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestDataRate(t *testing.T) {
	a := assertions.New(t)

	table := map[string]datarate.DR{
		`"SF7BW125"`: {DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 7, Bandwidth: 125000}}}},
		`50000`:      {DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_FSK{FSK: &ttnpb.FSKDataRate{BitRate: 50000}}}},
	}

	for s, dr := range table {
		enc, err := dr.MarshalJSON()
		a.So(err, should.BeNil)
		a.So(string(enc), should.Equal, s)

		var dec datarate.DR
		err = dec.UnmarshalJSON(enc)
		a.So(err, should.BeNil)
		a.So(dec, should.Resemble, dr)
	}

	var dr datarate.DR
	err := dr.UnmarshalJSON([]byte{})
	a.So(err, should.NotBeNil)
}

func TestValidLoRaDataRateParsing(t *testing.T) {
	a := assertions.New(t)

	table := map[string]datarate.DR{
		"SF6BW125":   {DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 6, Bandwidth: 125000}}}},
		"SF9BW500":   {DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 9, Bandwidth: 500000}}}},
		"SF5BW31.25": {DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 5, Bandwidth: 31250}}}},
	}
	for dr, expected := range table {
		actual, err := datarate.ParseLoRa(dr)
		a.So(err, should.BeNil)
		a.So(actual, should.Resemble, expected)
	}
}

func TestInvalidLoRaDataRateParsing(t *testing.T) {
	a := assertions.New(t)

	table := []string{
		"6BW125",
		"SF9B500",
	}
	for _, dr := range table {
		_, err := datarate.ParseLoRa(dr)
		a.So(err, should.NotBeNil)
	}
}

func TestStringer(t *testing.T) {
	a := assertions.New(t)

	table := map[datarate.DR]string{
		{DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 6, Bandwidth: 125000}}}}: "SF6BW125",
		{DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 9, Bandwidth: 500000}}}}: "SF9BW500",
		{DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 5, Bandwidth: 31250}}}}:  "SF5BW31.25",
		{DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_FSK{FSK: &ttnpb.FSKDataRate{BitRate: 50000}}}}:                           "50000",
	}

	for dr, expected := range table {
		a.So(dr.String(), should.Equal, expected)
	}
}
