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
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/udp-translator"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestDownlinks(t *testing.T) {
	a := assertions.New(t)
	var err error

	translate := translator.New(test.GetLogger(t))

	downlink := ttnpb.DownlinkMessage{
		TxMetadata: ttnpb.TxMetadata{
			Timestamp: 1886440700000,
		},
		Settings: ttnpb.TxSettings{
			Frequency:             925700000,
			Modulation:            ttnpb.Modulation_LORA,
			TxPower:               20,
			SpreadingFactor:       10,
			Bandwidth:             500000,
			PolarizationInversion: true,
		},
	}
	downlink.RawPayload, err = base64.StdEncoding.DecodeString("ffOO")
	data, err := translate.Downlink(&ttnpb.GatewayDown{DownlinkMessage: &downlink})
	a.So(err, should.BeNil)

	a.So(data.TxPacket.DatR.LoRa, should.Equal, "SF10BW500")
	a.So(data.TxPacket.Tmst, should.Equal, 1886440700)
	a.So(data.TxPacket.NCRC, should.Equal, true)
}

func TestDummyDownlink(t *testing.T) {
	a := assertions.New(t)

	translate := translator.New(test.GetLogger(t))

	downlink := ttnpb.DownlinkMessage{Settings: ttnpb.TxSettings{Modulation: 3939}} // Dummy modulation set
	_, err := translate.Downlink(&ttnpb.GatewayDown{DownlinkMessage: &downlink})
	a.So(err, should.NotBeNil)
}
