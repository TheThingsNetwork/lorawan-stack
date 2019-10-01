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

package udp

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/datarate"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestStatusPacket(t *testing.T) {
	statusPacket := `{
		"stat":{
		   "time":"2017-06-08 09:40:42 GMT",
		   "lati":52.34223,
		   "long":5.29685,
		   "alti":66,
		   "rxnb":0,
		   "rxok":0,
		   "rxfw":0,
		   "ackr":0.0,
		   "dwnb":0,
		   "txnb":0
		}
	 }`
	var d Data
	err := json.Unmarshal([]byte(statusPacket), &d)
	if err != nil {
		t.Error("Couldn't unmarshal status data:", err)
	}

	a := assertions.New(t)
	a.So(d, should.NotBeNil)
	a.So(d.Stat, should.NotBeNil)
	a.So(*d.Stat.Alti, should.Equal, 66)
	a.So(d.Stat.RxNb, should.Equal, 0)
	a.So(d.Stat.RxOk, should.Equal, 0)
	a.So(d.Stat.RxFW, should.Equal, 0)
	a.So(d.Stat.ACKR, should.AlmostEqual, 0.0)
	a.So(d.Stat.DWNb, should.Equal, 0)
	a.So(d.Stat.TxNb, should.Equal, 0)
}

func TestUplinkPacket(t *testing.T) {
	uplinkPacket := `{
		"rxpk":[
		   {
			  "tmst":445526776,
			  "chan":0,
			  "rfch":0,
			  "freq":868.099975,
			  "stat":1,
			  "modu":"LORA",
			  "datr":"SF7BW125",
			  "codr":"4/5",
			  "lsnr":-12,
			  "rssi":-112,
			  "size":61,
			  "data":"tlJ+3kao1MjU3ol8kuTwhziot4L/wQGMXngnecZaq5dXGpqZFTHWkzg/Hea7Y4NEjZND1gARpWtPdwC1vQ=="
		   }
		]
	 }`
	var d Data
	err := json.Unmarshal([]byte(uplinkPacket), &d)
	if err != nil {
		t.Error("Couldn't unmarshal uplink data:", err)
	}

	a := assertions.New(t)
	a.So(d, should.NotBeNil)
	a.So(d.RxPacket, should.NotBeNil)
	a.So(len(d.RxPacket), should.Equal, 1)

	uplink := d.RxPacket[0]
	a.So(uplink.Freq, should.AlmostEqual, 868.099975)
	a.So(uplink.Tmst, should.Equal, 445526776)
	a.So(uplink.Chan, should.Equal, 0)
	a.So(uplink.RFCh, should.Equal, 0)
	a.So(uplink.Stat, should.Equal, 1)
	a.So(uplink.Modu, should.Equal, "LORA")
	a.So(uplink.DatR, should.Resemble, datarate.DataRate{DataRate: ttnpb.DataRate{
		Modulation: &ttnpb.DataRate_LoRa{
			LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 7,
				Bandwidth:       125000,
			},
		},
	}})
	a.So(uplink.CodR, should.Equal, "4/5")
	a.So(uplink.LSNR, should.AlmostEqual, -12.0)
	a.So(uplink.RSSI, should.Equal, -112)
}

func TestDownlinkPacket(t *testing.T) {
	downlinkPacket := `{"txpk":
		{
		"imme":true,
		"freq":864.123456,
		"rfch":0,
		"powe":14,
		"modu":"LORA",
		"datr":"SF11BW125",
		"codr":"4/6",
		"ipol":false,
		"size":32,
		"data":"H3P3N2i9qc4yt7rK7ldqoeCVJGBybzPY5h1Dd7P7p8v"
		}}`
	var d Data
	err := json.Unmarshal([]byte(downlinkPacket), &d)
	if err != nil {
		t.Error("Couldn't unmarshal downlink data:", err)
	}

	a := assertions.New(t)
	a.So(d, should.NotBeNil)
	a.So(d.TxPacket, should.NotBeNil)

	tx := *d.TxPacket
	a.So(tx.Imme, should.Equal, true)
	a.So(tx.Freq, should.AlmostEqual, 864.123456)
	a.So(tx.RFCh, should.Equal, 0)
	a.So(tx.Powe, should.Equal, 14)
	a.So(tx.Modu, should.Equal, "LORA")
	a.So(tx.DatR, should.Resemble, datarate.DataRate{DataRate: ttnpb.DataRate{
		Modulation: &ttnpb.DataRate_LoRa{
			LoRa: &ttnpb.LoRaDataRate{
				SpreadingFactor: 11,
				Bandwidth:       125000,
			},
		},
	}})
	a.So(tx.CodR, should.Equal, "4/6")
	a.So(tx.IPol, should.Equal, false)
}
