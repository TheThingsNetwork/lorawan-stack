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

package frequencyplans_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/durationpb"
)

func uint64Ptr(v uint64) *uint64 { return &v }

func Example() {
	fetcher, err := fetch.FromHTTP(http.DefaultClient, "https://raw.githubusercontent.com/TheThingsNetwork/lorawan-frequency-plans")
	if err != nil {
		panic(err)
	}
	store := frequencyplans.NewStore(fetcher)

	ids, err := store.GetAllIDs()
	if err != nil {
		panic(err)
	}

	fmt.Println("Frequency plans available:")
	for _, id := range ids {
		fmt.Println("-", id)
	}

	euFP, err := store.GetByID("EU_863_870")
	if err != nil {
		panic(err)
	}

	fmt.Println("Content of the EU frequency plan:")
	fmt.Println(euFP)
}

func TestInvalidStore(t *testing.T) {
	a := assertions.New(t)

	store := frequencyplans.NewStore(fetch.NewMemFetcher(map[string][]byte{
		"frequency-plans.yml": []byte(`invalid-yaml`),
	}))

	_, err := store.GetAllIDs()
	a.So(err, should.NotBeNil)
}

func TestEmptyStore(t *testing.T) {
	a := assertions.New(t)

	store := frequencyplans.NewStore(fetch.NewMemFetcher(map[string][]byte{}))

	_, err := store.GetAllIDs()
	a.So(err, should.NotBeNil)

	_, err = store.GetByID("EU_863_870")
	a.So(err, should.NotBeNil)
}

func TestStore(t *testing.T) {
	a := assertions.New(t)

	store := frequencyplans.NewStore(fetch.NewMemFetcher(map[string][]byte{
		"frequency-plans.yml": []byte(`- id: AS_923
  description: South East Asia
  base-frequency: 915
  file: AS_923.yml
- id: JP
  base-id: AS_923
  description: Japan
  base-frequency: 915
  file: JP.yml
- id: KR
  base-id: AS_923
  description: South Korea
  base-frequency: 915
  file: KR.yml
- id: EU_863_870
  description: European Union
  file: EU.yml
  base-frequency: 868
- id: US_915
  description: United States
  file: US_915.yml
  base-frequency: 915
- id: SA
  base-id: AFRICA
  description: South Africa
  file: AS_923.yml
  base-frequency: 868
- id: CA
  base-id: US_915
  description: Canada
  file: EU.yml
  base-frequency: 915
`),
		"AS_923.yml": []byte(`band-id: AS_923
uplink-channels:
- frequency: 923000000
`),
		"US_915.yml": []byte(`invalid-yaml`),
		"JP.yml": []byte(`sub-bands:
- min-frequency: 923000000
  max-frequency: 923000000
  max-eirp: 42
listen-before-talk:
  rssi-target: 1.1
  rssi-offset: 2.2
  scan-time: 80
dwell-time:
  uplinks: true
  downlinks: true
  duration: 400ms
uplink-channels:
- frequency: 923000000
  dwell-time:
    enabled: true
    duration: 400ms
`),
		"KR.yml": []byte(`dwell-time:
  uplinks: true
  downlinks: true
uplink-channels:
- frequency: 923000000
  dwell-time:
    enabled: true
`),
	}))

	{
		ids, err := store.GetAllIDs()
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		a.So(ids, should.Contain, "AS_923")
		a.So(ids, should.Contain, "JP")
		a.So(ids, should.Contain, "KR")
		a.So(ids, should.Contain, "EU_863_870")
		a.So(ids, should.Contain, "US_915")
		a.So(ids, should.Contain, "SA")
		a.So(ids, should.Contain, "CA")
	}

	assertAS923Content := func(fp *frequencyplans.FrequencyPlan) {
		a.So(fp.UplinkChannels, should.HaveLength, 1)
		a.So(fp.UplinkChannels[0].Frequency, should.Equal, 923000000)
		a.So(fp.BandID, should.Equal, "AS_923")
	}

	// AS923 Frequency plan
	{
		fp, err := store.GetByID("AS_923")
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		assertAS923Content(fp)
	}

	// JP Frequency plan
	{
		fp, err := store.GetByID("JP")
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		assertAS923Content(fp)
		sb, ok := fp.FindSubBand(923000000)
		if !a.So(ok, should.BeTrue) {
			t.FailNow()
		}
		a.So(*sb.MaxEIRP, should.Equal, 42)
		a.So(fp.LBT, should.NotBeNil)
		a.So(fp.LBT.RSSIOffset, should.AlmostEqual, 2.2, 0.00001)
		a.So(fp.LBT.ScanTime, should.Equal, 80)
		a.So(*fp.UplinkChannels[0].DwellTime.Enabled, should.BeTrue)
	}

	// Invalid frequency plan (no dwell time duration)
	{
		_, err := store.GetByID("KR")
		a.So(errors.IsDataLoss(err), should.BeTrue)
	}

	// Unknown frequency plan
	{
		_, err := store.GetByID("Unknown")
		a.So(errors.IsNotFound(err), should.BeTrue)
	}

	// Frequency plan non-existent
	{
		_, err := store.GetByID("EU_863_870")
		a.So(errors.IsNotFound(err), should.BeTrue)
	}

	// Frequency plan with invalid yaml
	{
		_, err := store.GetByID("US_915")
		a.So(err, should.NotBeNil)
	}

	// Frequency plan with non-existent base
	{
		_, err := store.GetByID("SA")
		a.So(errors.IsNotFound(err), should.BeTrue)
	}

	// Frequency plan with base with invalid yaml
	{
		_, err := store.GetByID("CA")
		a.So(err, should.NotBeNil)
	}
}

func TestProtoConversion(t *testing.T) {
	for i, tc := range []struct {
		Input  *frequencyplans.FrequencyPlan
		Output *ttnpb.ConcentratorConfig
	}{
		{
			Input: &frequencyplans.FrequencyPlan{
				BandID: "US_902_928",
				UplinkChannels: []frequencyplans.Channel{
					{Frequency: 922100000, Radio: 0},
					{Frequency: 922300000, Radio: 0},
					{Frequency: 922500000, Radio: 0},
				},
				DownlinkChannels: []frequencyplans.Channel{
					{Frequency: 922100000, Radio: 0},
					{Frequency: 922300000, Radio: 0},
					{Frequency: 922500000, Radio: 0},
				},
				Radios: []frequencyplans.Radio{
					{
						Enable:    true,
						ChipType:  "SX1257",
						Frequency: 922300000,
						TxConfiguration: &frequencyplans.RadioTxConfiguration{
							MinFrequency:   909000000,
							MaxFrequency:   925000000,
							NotchFrequency: uint64Ptr(920000000),
						},
					},
					{
						Enable:    false,
						ChipType:  "SX1257",
						Frequency: 923000000,
					},
				},
				ClockSource: 1,
			},
			Output: &ttnpb.ConcentratorConfig{
				Channels: []*ttnpb.ConcentratorConfig_Channel{
					{Frequency: 922100000, Radio: 0},
					{Frequency: 922300000, Radio: 0},
					{Frequency: 922500000, Radio: 0},
				},
				Radios: []*ttnpb.GatewayRadio{
					{
						Enable:    true,
						ChipType:  "SX1257",
						Frequency: 922300000,
						TxConfiguration: &ttnpb.GatewayRadio_TxConfiguration{
							MinFrequency:   909000000,
							MaxFrequency:   925000000,
							NotchFrequency: 920000000,
						},
					},
					{
						Enable:    false,
						ChipType:  "SX1257",
						Frequency: 923000000,
					},
				},
				ClockSource: 1,
			},
		},
		{
			Input: &frequencyplans.FrequencyPlan{
				BandID: "EU_863_870",
				FSKChannel: &frequencyplans.FSKChannel{
					Frequency: 868800000,
					Radio:     1,
					DataRate:  7,
				},
				LoRaStandardChannel: &frequencyplans.LoRaStandardChannel{
					Frequency: 868300000,
					Radio:     1,
					DataRate:  6,
				},
			},
			Output: &ttnpb.ConcentratorConfig{
				FskChannel: &ttnpb.ConcentratorConfig_FSKChannel{
					Frequency: 868800000,
					Radio:     1,
				},
				LoraStandardChannel: &ttnpb.ConcentratorConfig_LoRaStandardChannel{
					Frequency:       868300000,
					Radio:           1,
					SpreadingFactor: 7,
					Bandwidth:       250000,
				},
			},
		},
		{
			Input: &frequencyplans.FrequencyPlan{
				BandID: "AS_923",
				LBT: &frequencyplans.LBT{
					ScanTime: 32 * time.Microsecond,
				},
				PingSlot: &frequencyplans.Channel{
					Frequency: 923000000,
					Radio:     1,
				},
			},
			Output: &ttnpb.ConcentratorConfig{
				Lbt: &ttnpb.ConcentratorConfig_LBTConfiguration{
					ScanTime: durationpb.New(32 * time.Microsecond),
				},
				PingSlot: &ttnpb.ConcentratorConfig_Channel{
					Frequency: 923000000,
					Radio:     1,
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("Proto%d", i), func(t *testing.T) {
			a := assertions.New(t)
			output, err := tc.Input.ToConcentratorConfig()
			a.So(err, should.BeNil)
			a.So(output, should.Resemble, tc.Output)
		})
	}
}

func TestRespectsDwellTime(t *testing.T) {
	a := assertions.New(t)

	store := frequencyplans.NewStore(fetch.NewMemFetcher(map[string][]byte{
		"frequency-plans.yml": []byte(`- id: Test
  description: Test
  base-frequency: 915
  file: test.yml
`),
		"test.yml": []byte(`band-id: AS_923
uplink-channels:
- frequency: 1
- frequency: 2
  dwell-time:
    enabled: true
    duration: 100ms
- frequency: 3
  dwell-time:
    enabled: true
downlink-channels:
- frequency: 1
  dwell-time:
    enabled: false
- frequency: 2
  dwell-time:
    duration: 100ms
- frequency: 3
dwell-time:
  uplinks: false
  downlinks: true
  duration: 400ms
`),
	}))

	fp, err := store.GetByID("Test")
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	for _, tc := range []struct {
		IsDownlink bool
		Frequency  uint64
		Duration   time.Duration
		Expected   bool
	}{
		{
			IsDownlink: false,
			Frequency:  1,
			Duration:   1 * time.Second,
			Expected:   true,
		},
		{
			IsDownlink: false,
			Frequency:  2,
			Duration:   50 * time.Millisecond,
			Expected:   true,
		},
		{
			IsDownlink: false,
			Frequency:  2,
			Duration:   150 * time.Millisecond,
			Expected:   false,
		},
		{
			IsDownlink: false,
			Frequency:  3,
			Duration:   300 * time.Millisecond,
			Expected:   true,
		},
		{
			IsDownlink: false,
			Frequency:  3,
			Duration:   500 * time.Millisecond,
			Expected:   false,
		},
		{
			IsDownlink: false,
			Frequency:  4,
			Duration:   500 * time.Millisecond,
			Expected:   true,
		},
		{
			IsDownlink: true,
			Frequency:  1,
			Duration:   1 * time.Second,
			Expected:   true,
		},
		{
			IsDownlink: true,
			Frequency:  2,
			Duration:   50 * time.Millisecond,
			Expected:   true,
		},
		{
			IsDownlink: true,
			Frequency:  2,
			Duration:   150 * time.Millisecond,
			Expected:   false,
		},
		{
			IsDownlink: true,
			Frequency:  3,
			Duration:   100 * time.Millisecond,
			Expected:   true,
		},
		{
			IsDownlink: true,
			Frequency:  3,
			Duration:   500 * time.Millisecond,
			Expected:   false,
		},
		{
			IsDownlink: true,
			Frequency:  4,
			Duration:   500 * time.Millisecond,
			Expected:   false,
		},
	} {
		dir := "DL"
		if !tc.IsDownlink {
			dir = "UL"
		}
		t.Run(fmt.Sprintf("%v/%v/%v", dir, tc.Frequency, tc.Duration), func(t *testing.T) {
			a := assertions.New(t)
			a.So(fp.RespectsDwellTime(tc.IsDownlink, tc.Frequency, tc.Duration), should.Equal, tc.Expected)
		})
	}
}

func TestGetGatewayFrequencyPlans(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	store := frequencyplans.NewStore(fetch.NewMemFetcher(map[string][]byte{
		"frequency-plans.yml": []byte(`- id: EU_863_870
  band-id: EU_863_870
  name: Europe 863-870 MHz (SF12 for RX2)
  description: Default frequency plan for Europe
  base-frequency: 868
  file: EU_863_870.yml
- id: EU_863_870_TTN
  band-id: EU_863_870
  name: Europe 863-870 MHz (SF9 for RX2 - recommended)
  description: TTN Community Network frequency plan for Europe, using SF9 for RX2
  base-frequency: 868
  base-id: EU_863_870
  file: EU_863_870_TTN.yml
- id: US_902_928_FSB_1
  band-id: US_902_928
  name: United States 902-928 MHz, FSB 1
  description: Default frequency plan for the United States and Canada, using sub-band 1
  base-frequency: 915
  file: US_902_928_FSB_1.yml
- id: AS_923_2
  band-id: AS_923_2
  name: Asia 920-923 MHz (AS923 Group 2) with only default channels
  description: Compatibility frequency plan for Asian countries with common channels in the 921.4-922.0 MHz sub-band
  base-frequency: 915
  file: AS_923_2.yml
- id: AS_923_2_DT
  band-id: AS_923_2
  name: Asia 920-923 MHz (AS923 Group 2) with only default channels and dwell time enabled
  base-frequency: 915
  base-id: AS_923_2
  file: enable_dwell_time_400ms.yml
`),
		"EU_863_870.yml": []byte(`band-id: EU_863_870
max-eirp: 1
gateways: false
`),
		"EU_863_870_TTN.yml": []byte(`max-eirp: 2
gateways: true
`),
		"US_902_928_FSB_1.yml": []byte(`band-id: US_902_928
max-eirp: 3
`),
		"AS_923_2.yml": []byte(`band-id: AS_923_2
max-eirp: 4
gateways: true
`),
		"enable_dwell_time_400ms.yml": []byte(`max-eirp: 5
gateways: false
`),
	}))

	// Frequency plan with gateways
	{
		plans, err := store.GetGatewayFrequencyPlans()
		a.So(err, should.BeNil)
		a.So(plans, should.HaveLength, 2)
		a.So(plans[0].BandID, should.Equal, "EU_863_870")
		a.So(plans[1].BandID, should.Equal, "AS_923_2")
		a.So(*plans[0].MaxEIRP, should.Equal, 2)
		a.So(*plans[1].MaxEIRP, should.Equal, 4)
	}
}
