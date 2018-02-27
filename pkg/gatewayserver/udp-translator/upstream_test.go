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

package translator_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/udp"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/udp-translator"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/TheThingsNetwork/ttn/pkg/version"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestStatusGrLocation(t *testing.T) {
	a := assertions.New(t)

	status := []byte(`{"stat":{"rxfw":0,"hal":"5.1.0","fpga":2,"dsp":31,"lpps":2,"lmnw":3,"lmst":1,"lmok":3,"temp":30,"lati":52.34223,"long":5.29685,"txnb":0,"dwnb":0,"alti":66,"rxok":0,"boot":"2017-06-07 09:40:42 GMT","time":"2017-06-08 09:40:42 GMT","rxnb":0,"ackr":0.0}}`)
	var statusData udp.Data
	err := json.Unmarshal(status, &statusData)
	a.So(err, should.BeNil)

	location := ttnpb.Location{Longitude: 15.56, Latitude: 3.0}

	translate := translator.NewWithLocation(test.GetLogger(t), location)
	upstream, err := translate.Upstream(statusData, translator.Metadata{IP: "127.0.0.1", ID: ids})
	a.So(err, should.BeNil)

	a.So(upstream.GatewayStatus, should.NotBeNil)
	a.So(upstream.GatewayStatus.AntennasLocation, should.NotBeNil)
	a.So(len(upstream.GatewayStatus.AntennasLocation), should.Equal, 1)
	a.So(upstream.GatewayStatus.AntennasLocation[0].Latitude, should.Equal, 3.0)
	a.So(upstream.GatewayStatus.AntennasLocation[0].Longitude, should.Equal, 15.56)
}

func TestStatus(t *testing.T) {
	a := assertions.New(t)

	status := []byte(`{"stat":{"rxfw":0,"hal":"5.1.0","fpga":2,"dsp":31,"lpps":2,"lmnw":3,"lmst":1,"lmok":3,"temp":30,"lati":52.34223,"long":5.29685,"txnb":0,"dwnb":0,"alti":66,"rxok":0,"boot":"2017-06-07 09:40:42 GMT","time":"2017-06-08 09:40:42 GMT","rxnb":0,"ackr":0.0}}`)
	var statusData udp.Data
	err := json.Unmarshal(status, &statusData)
	a.So(err, should.BeNil)

	translate := translator.New(test.GetLogger(t))
	upstream, err := translate.Upstream(statusData, translator.Metadata{
		IP:       "127.0.0.1",
		ID:       ids,
		Versions: map[string]string{"bridge": version.TTN},
	})
	a.So(err, should.BeNil)

	a.So(upstream.GatewayStatus, should.NotBeNil)

	a.So(upstream.GatewayStatus.AntennasLocation, should.NotBeNil)
	a.So(len(upstream.GatewayStatus.AntennasLocation), should.Equal, 1)
	a.So(upstream.GatewayStatus.AntennasLocation[0].Longitude, should.AlmostEqual, 5.29685, 0.0001)
	a.So(upstream.GatewayStatus.AntennasLocation[0].Latitude, should.AlmostEqual, 52.34223, 0.0001)
	a.So(upstream.GatewayStatus.AntennasLocation[0].Altitude, should.AlmostEqual, 66)

	a.So(upstream.GatewayStatus.Versions, should.NotBeNil)
	a.So(upstream.GatewayStatus.Metrics, should.NotBeNil)

	a.So(upstream.GatewayStatus.Versions["bridge"], should.Equal, version.TTN)
	a.So(upstream.GatewayStatus.Versions["hal"], should.Equal, "5.1.0")
	a.So(upstream.GatewayStatus.Versions["fpga"], should.Equal, "2")
	a.So(upstream.GatewayStatus.Versions["dsp"], should.Equal, "31")

	a.So(upstream.GatewayStatus.Metrics["rxfw"], should.AlmostEqual, 0)
	a.So(upstream.GatewayStatus.Metrics["txnb"], should.AlmostEqual, 0)
	a.So(upstream.GatewayStatus.Metrics["dwnb"], should.AlmostEqual, 0)
	a.So(upstream.GatewayStatus.Metrics["rxok"], should.AlmostEqual, 0)
	a.So(upstream.GatewayStatus.Metrics["rxnb"], should.AlmostEqual, 0)
	a.So(upstream.GatewayStatus.Metrics["ackr"], should.AlmostEqual, 0)
	a.So(upstream.GatewayStatus.Metrics["temp"], should.AlmostEqual, 30)
	a.So(upstream.GatewayStatus.Metrics["lpps"], should.AlmostEqual, 2)
	a.So(upstream.GatewayStatus.Metrics["lmnw"], should.AlmostEqual, 3)
	a.So(upstream.GatewayStatus.Metrics["lmst"], should.AlmostEqual, 1)
	a.So(upstream.GatewayStatus.Metrics["lmok"], should.AlmostEqual, 3)

	a.So(upstream.GatewayStatus.BootTime, should.NotBeNil)
	a.So(upstream.GatewayStatus.Time, should.NotBeNil)
	currentTime := time.Date(2017, 06, 8, 9, 40, 42, 0, time.UTC)
	a.So(upstream.GatewayStatus.Time, should.Equal, currentTime)
	bootTime := time.Date(2017, 06, 7, 9, 40, 42, 0, time.UTC)
	a.So(upstream.GatewayStatus.BootTime, should.Equal, bootTime)
}

func TestUplink(t *testing.T) {
	a := assertions.New(t)

	rx := []byte(`{"rxpk":[{"tmst":368384825,"chan":0,"rfch":0,"freq":868.100000,"stat":1,"modu":"LORA","datr":"SF7BW125","codr":"4/5","lsnr":-11,"rssi":-107,"size":108,"data":"Wqish6GVYpKy6o9WFHingeTJ1oh+ABc8iALBvwz44yxZP+BKDocaC5VQT5Y6dDdUaBILVjRMz0Ynzow1U/Kkts9AoZh3Ja3DX+DyY27exB+BKpSx2rXJ2vs9svm/EKYIsPF0RG1E+7lBYaD9"}]}`)
	var rxData udp.Data
	err := json.Unmarshal(rx, &rxData)
	a.So(err, should.BeNil)

	translate := translator.New(test.GetLogger(t))
	upstream, err := translate.Upstream(rxData, translator.Metadata{ID: ids})
	a.So(err, should.BeNil)

	a.So(len(upstream.UplinkMessages), should.Equal, 1)

	a.So(upstream.UplinkMessages[0].Settings.CodingRate, should.Equal, "4/5")
	a.So(upstream.UplinkMessages[0].Settings.SpreadingFactor, should.Equal, 7)
	a.So(upstream.UplinkMessages[0].Settings.Bandwidth, should.Equal, 125000)
	a.So(upstream.UplinkMessages[0].Settings.Frequency, should.Equal, uint64(868100000))
	a.So(upstream.UplinkMessages[0].Settings.Modulation, should.Equal, ttnpb.Modulation_LORA)

	a.So(upstream.UplinkMessages[0].RxMetadata[0].Timestamp, should.Equal, uint64(368384825000))
	a.So(len(upstream.UplinkMessages[0].RawPayload), should.Equal, base64.StdEncoding.DecodedLen(len("Wqish6GVYpKy6o9WFHingeTJ1oh+ABc8iALBvwz44yxZP+BKDocaC5VQT5Y6dDdUaBILVjRMz0Ynzow1U/Kkts9AoZh3Ja3DX+DyY27exB+BKpSx2rXJ2vs9svm/EKYIsPF0RG1E+7lBYaD9")))
}

func TestMultiAntennaUplink(t *testing.T) {
	a := assertions.New(t)

	rx := []byte(`{
		"rxpk": [{
			"tmst": 879148780,
			"time": "2017-07-04T13:51:17.997099Z",
			"rfch": 0,
			"freq": 868.500000,
			"stat": 1,
			"modu": "LORA",
			"datr": "SF7BW125",
			"codr": "4/5",
			"size": 24,
			"data": "gM+AMQcAvgQBlohnlJqUGOJKTDuTscQD",
			"rsig": [{
				"ant": 0,
				"chan": 7,
				"rssic": -95,
				"lsnr": 14.0,
				"etime": "42QMzOlYSSPMMeqVPrY0fQ==",
				"rssis": -95,
				"rssisd": 0,
				"ftime": 1255738435,
				"foff": -8898,
				"ft2d": -251,
				"rfbsb": 100,
				"rs2s1": 97
			}, {
				"ant": 1,
				"chan": 23,
				"rssic": -88,
				"lsnr": 14.0,
				"etime": "djGiSzOC+gCT7vRPv7+Asw==",
				"rssis": -88,
				"rssisd": 0,
				"ftime": -1252538435,
				"foff": -8898,
				"ft2d": -187,
				"rfbsb": 100,
				"rs2s1": 104
			}]
		}]
	}`)
	var rxData udp.Data
	err := json.Unmarshal(rx, &rxData)
	a.So(err, should.BeNil)

	translate := translator.New(test.GetLogger(t))
	_, err = translate.Upstream(rxData, translator.Metadata{ID: ids})
	a.So(err, should.BeNil)
}
