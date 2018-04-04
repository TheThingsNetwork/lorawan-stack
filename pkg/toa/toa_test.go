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

package toa

import (
	"fmt"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	assertions "github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func buildLoRaDownlinkFromParameters(payloadSize int, dataRate types.DataRate, codingRate string) (downlink ttnpb.DownlinkMessage, err error) {
	payload := []byte{}
	for i := 0; i < payloadSize; i++ {
		payload = append(payload, 0)
	}

	downlink = ttnpb.DownlinkMessage{
		RawPayload: payload,
		Settings: ttnpb.TxSettings{
			CodingRate: codingRate,
			Modulation: ttnpb.Modulation_LORA,
		},
	}

	bw, err := dataRate.Bandwidth()
	if err != nil {
		return
	}

	sf, err := dataRate.SpreadingFactor()
	if err != nil {
		return
	}

	downlink.Settings.Bandwidth = bw
	downlink.Settings.SpreadingFactor = uint32(sf)
	return downlink, nil
}

func TestInvalidModulation(t *testing.T) {
	a := assertions.New(t)

	dl := ttnpb.DownlinkMessage{
		Settings: ttnpb.TxSettings{
			Modulation: ttnpb.Modulation_FSK + ttnpb.Modulation_LORA + 1,
		},
	}

	_, err := Compute(dl.RawPayload, dl.Settings)
	a.So(err, should.NotBeNil)
}

func TestInvalidLoRa(t *testing.T) {
	a := assertions.New(t)

	_, err := buildLoRaDownlinkFromParameters(10, types.DataRate{LoRa: "SFUT"}, "4/5")
	a.So(err, should.NotBeNil)

	downlink, err := buildLoRaDownlinkFromParameters(10, types.DataRate{LoRa: "SF10BW125"}, "1/9")
	a.So(err, should.BeNil)
	_, err = Compute(downlink.RawPayload, downlink.Settings)
	a.So(err, should.NotBeNil)

	downlink, err = buildLoRaDownlinkFromParameters(10, types.DataRate{LoRa: "SF7BW125"}, "1/9")
	a.So(err, should.BeNil)
	downlink.Settings.SpreadingFactor = 0
	_, err = Compute(downlink.RawPayload, downlink.Settings)
	a.So(err, should.NotBeNil)

	downlink, err = buildLoRaDownlinkFromParameters(10, types.DataRate{LoRa: "SF7BW125"}, "4/5")
	a.So(err, should.BeNil)
	downlink.Settings.Bandwidth = 0
	_, err = Compute(downlink.RawPayload, downlink.Settings)
	a.So(err, should.NotBeNil)
}

func TestDifferentLoRaSFs(t *testing.T) {
	a := assertions.New(t)

	sfTests := map[types.DataRate]uint{
		{LoRa: "SF7BW125"}:  41216,
		{LoRa: "SF8BW125"}:  72192,
		{LoRa: "SF9BW125"}:  144384,
		{LoRa: "SF10BW125"}: 288768,
		{LoRa: "SF11BW125"}: 577536,
		{LoRa: "SF12BW125"}: 991232,
	}

	for dr, us := range sfTests {
		dl, err := buildLoRaDownlinkFromParameters(10, dr, "4/5")
		a.So(err, should.BeNil)
		toa, err := Compute(dl.RawPayload, dl.Settings)
		a.So(err, should.BeNil)
		a.So(toa, should.AlmostEqual, time.Duration(us)*time.Microsecond)
	}
}

func TestDifferentLoRaBWs(t *testing.T) {
	a := assertions.New(t)

	bwTests := map[types.DataRate]uint{
		{LoRa: "SF7BW125"}: 41216,
		{LoRa: "SF7BW250"}: 20608,
		{LoRa: "SF7BW500"}: 10304,
	}

	for dr, us := range bwTests {
		dl, err := buildLoRaDownlinkFromParameters(10, dr, "4/5")
		a.So(err, should.BeNil)
		toa, err := Compute(dl.RawPayload, dl.Settings)
		a.So(err, should.BeNil)
		a.So(toa, should.AlmostEqual, time.Duration(us)*time.Microsecond)
	}
}

func TestDifferentLoRaCRs(t *testing.T) {
	a := assertions.New(t)

	crTests := map[string]uint{
		"4/5": 41216,
		"4/6": 45312,
		"4/7": 49408,
		"4/8": 53504,
	}

	for cr, us := range crTests {
		dl, err := buildLoRaDownlinkFromParameters(10, types.DataRate{LoRa: "SF7BW125"}, cr)
		a.So(err, should.BeNil)
		toa, err := Compute(dl.RawPayload, dl.Settings)
		a.So(err, should.BeNil)
		a.So(toa, should.AlmostEqual, time.Duration(us)*time.Microsecond)
	}
}

func TestDifferentLoRaPayloadSizes(t *testing.T) {
	a := assertions.New(t)

	plTests := map[int]uint{
		13: 46336,
		14: 46336,
		15: 46336,
		16: 51456,
		17: 51456,
		18: 51456,
		19: 51456,
	}

	for size, us := range plTests {
		dl, err := buildLoRaDownlinkFromParameters(size, types.DataRate{LoRa: "SF7BW125"}, "4/5")
		a.So(err, should.BeNil)
		toa, err := Compute(dl.RawPayload, dl.Settings)
		a.So(err, should.BeNil)
		a.So(toa, should.AlmostEqual, time.Duration(us)*time.Microsecond)
	}

}

func TestFSK(t *testing.T) {
	a := assertions.New(t)

	payload := []byte{}
	payloadSize := 200
	for i := 0; i < payloadSize; i++ {
		payload = append(payload, 0)
	}

	d := ttnpb.DownlinkMessage{
		RawPayload: payload,
		Settings: ttnpb.TxSettings{
			BitRate:    50000,
			Modulation: ttnpb.Modulation_FSK,
		},
	}

	toa, err := Compute(d.RawPayload, d.Settings)
	a.So(err, should.BeNil)
	a.So(toa, should.AlmostEqual, 33760*time.Microsecond)
}

func getDownlink() ttnpb.DownlinkMessage { return ttnpb.DownlinkMessage{} }

func ExampleCompute() {
	var downlink ttnpb.DownlinkMessage
	downlink = getDownlink()

	toa, err := Compute(downlink.RawPayload, downlink.Settings)
	if err != nil {
		panic(err)
	}

	fmt.Println("Time on air:", toa)
}
