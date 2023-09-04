// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package lorawan_test

import (
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestMarshalRelayForwardDownlinkReq(t *testing.T) {
	t.Parallel()

	a := assertions.New(t)

	_, err := lorawan.MarshalRelayForwardDownlinkReq(&ttnpb.RelayForwardDownlinkReq{})
	a.So(err, should.NotBeNil)

	b, err := lorawan.MarshalRelayForwardDownlinkReq(&ttnpb.RelayForwardDownlinkReq{
		RawPayload: []byte{0x01, 0x02, 0x03},
	})
	if a.So(err, should.BeNil) {
		a.So(b, should.Resemble, []byte{0x01, 0x02, 0x03})
	}
}

func TestUnmarshalRelayForwardDownlinkReq(t *testing.T) {
	t.Parallel()

	a := assertions.New(t)

	err := lorawan.UnmarshalRelayForwardDownlinkReq(nil, &ttnpb.RelayForwardDownlinkReq{})
	a.So(err, should.NotBeNil)

	var req ttnpb.RelayForwardDownlinkReq
	err = lorawan.UnmarshalRelayForwardDownlinkReq([]byte{0x01, 0x02, 0x03}, &req)
	if a.So(err, should.BeNil) {
		a.So(req.RawPayload, should.Resemble, []byte{0x01, 0x02, 0x03})
	}
}

func TestMarshalRelayForwardUplinkReq(t *testing.T) {
	t.Parallel()

	a, phy := assertions.New(t), &band.EU_863_870_RP2_V1_0_4
	for _, tc := range []struct {
		Name    string
		Request *ttnpb.RelayForwardUplinkReq
	}{
		{
			Name: "InvalidWORChannel",
			Request: &ttnpb.RelayForwardUplinkReq{
				WorChannel: 2,
			},
		},
		{
			Name: "InvalidRSSI",
			Request: &ttnpb.RelayForwardUplinkReq{
				WorChannel: ttnpb.RelayWORChannel_RELAY_WOR_CHANNEL_DEFAULT,
				Rssi:       -143,
			},
		},
		{
			Name: "InvalidSNR",
			Request: &ttnpb.RelayForwardUplinkReq{
				WorChannel: ttnpb.RelayWORChannel_RELAY_WOR_CHANNEL_DEFAULT,
				Rssi:       -64,
				Snr:        -21,
			},
		},
		{
			Name: "InvalidDataRateIndex",
			Request: &ttnpb.RelayForwardUplinkReq{
				WorChannel: ttnpb.RelayWORChannel_RELAY_WOR_CHANNEL_DEFAULT,
				Rssi:       -64,
				Snr:        5,
				DataRate:   &ttnpb.DataRate{},
			},
		},
		{
			Name: "InvalidRawPayload",
			Request: &ttnpb.RelayForwardUplinkReq{
				WorChannel: ttnpb.RelayWORChannel_RELAY_WOR_CHANNEL_DEFAULT,
				Rssi:       -64,
				Snr:        5,
				DataRate:   phy.DataRates[ttnpb.DataRateIndex_DATA_RATE_1].Rate,
			},
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a := assertions.New(t)
			_, err := lorawan.MarshalRelayForwardUplinkReq(phy, tc.Request)
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		})
	}

	b, err := lorawan.MarshalRelayForwardUplinkReq(phy, &ttnpb.RelayForwardUplinkReq{
		WorChannel: ttnpb.RelayWORChannel_RELAY_WOR_CHANNEL_SECONDARY,
		Rssi:       -64,
		Snr:        5,
		DataRate:   phy.DataRates[ttnpb.DataRateIndex_DATA_RATE_1].Rate,
		Frequency:  868100000,
		RawPayload: []byte{0x01, 0x02, 0x03},
	})
	if a.So(err, should.BeNil) {
		a.So(b, should.Resemble, []byte{0x91, 0x63, 0x01, 0x28, 0x76, 0x84, 0x01, 0x02, 0x03})
	}
}

func TestUnmarshalRelayForwardUplinkReq(t *testing.T) {
	t.Parallel()

	a, phy := assertions.New(t), &band.EU_863_870_RP2_V1_0_4

	err := lorawan.UnmarshalRelayForwardUplinkReq(phy, nil, &ttnpb.RelayForwardUplinkReq{})
	a.So(err, should.NotBeNil)

	var req ttnpb.RelayForwardUplinkReq
	err = lorawan.UnmarshalRelayForwardUplinkReq(phy, []byte{0x91, 0x63, 0x01, 0x28, 0x76, 0x84, 0x01, 0x02, 0x03}, &req)
	if a.So(err, should.BeNil) {
		a.So(req.WorChannel, should.Equal, ttnpb.RelayWORChannel_RELAY_WOR_CHANNEL_SECONDARY)
		a.So(req.Rssi, should.Equal, -64)
		a.So(req.Snr, should.Equal, 5)
		a.So(req.DataRate, should.Resemble, phy.DataRates[ttnpb.DataRateIndex_DATA_RATE_1].Rate)
		a.So(req.Frequency, should.Equal, 868100000)
		a.So(req.RawPayload, should.Resemble, []byte{0x01, 0x02, 0x03})
	}
}
