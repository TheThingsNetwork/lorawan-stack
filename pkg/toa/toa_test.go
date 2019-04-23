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

package toa

import (
	"fmt"
	"testing"
	"time"

	assertions "github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func buildLoRaDownlinkFromParameters(payloadSize int, frequency uint64, dataRate ttnpb.DataRate, codingRate string) (downlink ttnpb.DownlinkMessage, err error) {
	payload := []byte{}
	for i := 0; i < payloadSize; i++ {
		payload = append(payload, 0)
	}
	scheduled := &ttnpb.TxSettings{
		Frequency:  frequency,
		DataRate:   dataRate,
		CodingRate: codingRate,
	}
	downlink = ttnpb.DownlinkMessage{
		RawPayload: payload,
		Settings: &ttnpb.DownlinkMessage_Scheduled{
			Scheduled: scheduled,
		},
	}
	return downlink, nil
}

func TestInvalidLoRa(t *testing.T) {
	a := assertions.New(t)

	// Invalid coding rate.
	{
		downlink, err := buildLoRaDownlinkFromParameters(10, 868100000, ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_LoRa{
				LoRa: &ttnpb.LoRaDataRate{
					SpreadingFactor: 10,
					Bandwidth:       125000,
				},
			},
		}, "1/9")
		_, err = Compute(len(downlink.RawPayload), *downlink.GetScheduled())
		a.So(err, should.NotBeNil)
	}

	// Invalid spreading factor.
	{
		downlink, err := buildLoRaDownlinkFromParameters(10, 868100000, ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_LoRa{
				LoRa: &ttnpb.LoRaDataRate{
					SpreadingFactor: 0,
					Bandwidth:       125000,
				},
			},
		}, "4/5")
		_, err = Compute(len(downlink.RawPayload), *downlink.GetScheduled())
		a.So(err, should.NotBeNil)
	}

	// Invalid bandwidth.
	{
		downlink, err := buildLoRaDownlinkFromParameters(10, 868100000, ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_LoRa{
				LoRa: &ttnpb.LoRaDataRate{
					SpreadingFactor: 7,
					Bandwidth:       0,
				},
			},
		}, "4/5")
		_, err = Compute(len(downlink.RawPayload), *downlink.GetScheduled())
		a.So(err, should.NotBeNil)
	}
}

func TestDifferentLoRaSFs(t *testing.T) {
	a := assertions.New(t)
	sfTests := map[ttnpb.DataRate]uint{
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 7, Bandwidth: 125000}}}:  41216,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 8, Bandwidth: 125000}}}:  72192,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 9, Bandwidth: 125000}}}:  144384,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 10, Bandwidth: 125000}}}: 288768,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 11, Bandwidth: 125000}}}: 577536,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 12, Bandwidth: 125000}}}: 991232,
	}
	for dr, us := range sfTests {
		dl, err := buildLoRaDownlinkFromParameters(10, 868100000, dr, "4/5")
		toa, err := Compute(len(dl.RawPayload), *dl.GetScheduled())
		a.So(err, should.BeNil)
		a.So(toa, should.AlmostEqual, time.Duration(us)*time.Microsecond)
	}
}

func TestDifferentLoRaBWs(t *testing.T) {
	a := assertions.New(t)
	bwTests := map[ttnpb.DataRate]uint{
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 7, Bandwidth: 125000}}}: 41216,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 7, Bandwidth: 250000}}}: 20608,
		{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 7, Bandwidth: 500000}}}: 10304,
	}
	for dr, us := range bwTests {
		dl, err := buildLoRaDownlinkFromParameters(10, 868100000, dr, "4/5")
		toa, err := Compute(len(dl.RawPayload), *dl.GetScheduled())
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
		dl, err := buildLoRaDownlinkFromParameters(10, 868100000, ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 7, Bandwidth: 125000}}}, cr)
		toa, err := Compute(len(dl.RawPayload), *dl.GetScheduled())
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
		dl, err := buildLoRaDownlinkFromParameters(size, 868100000, ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{SpreadingFactor: 7, Bandwidth: 125000}}}, "4/5")
		toa, err := Compute(len(dl.RawPayload), *dl.GetScheduled())
		a.So(err, should.BeNil)
		a.So(toa, should.AlmostEqual, time.Duration(us)*time.Microsecond)
	}
}

func TestFSK(t *testing.T) {
	a := assertions.New(t)
	payloadSize := 200
	scheduled := ttnpb.TxSettings{
		Frequency: 868300000,
		DataRate: ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_FSK{
				FSK: &ttnpb.FSKDataRate{
					BitRate: 50000,
				},
			},
		},
	}
	toa, err := Compute(payloadSize, scheduled)
	a.So(err, should.BeNil)
	a.So(toa, should.AlmostEqual, 33760*time.Microsecond)
}

func getDownlink() ttnpb.DownlinkMessage { return ttnpb.DownlinkMessage{} }

func ExampleCompute() {
	var downlink ttnpb.DownlinkMessage
	downlink = getDownlink()

	toa, err := Compute(len(downlink.RawPayload), *downlink.GetScheduled())
	if err != nil {
		panic(err)
	}

	fmt.Println("Time on air:", toa)
}
