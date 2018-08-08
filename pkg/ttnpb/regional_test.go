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

package ttnpb_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	yaml "gopkg.in/yaml.v2"
)

func TestUnmarshalEU(t *testing.T) {
	a := assertions.New(t)

	yamlDocument := `band-id: EU_863_870
channels:
- frequency: 867100000
- frequency: 867300000
- frequency: 867500000
- frequency: 867700000
- frequency: 867900000
- frequency: 868100000
- frequency: 868300000
- frequency: 868500000
lora-standard-channel:
  frequency: 863000000
  data-rate:
    index: 6
fsk-channel:
  frequency: 868800000
  data-rate:
    index: 7`

	fp := ttnpb.FrequencyPlan{}

	err := yaml.Unmarshal([]byte(yamlDocument), &fp)
	a.So(err, should.BeNil)

	for _, channel := range fp.Channels {
		a.So(channel.GetDataRate(), should.BeNil)
	}

	a.So(len(fp.Channels), should.Equal, 8)
	a.So(fp.LoraStandardChannel, should.NotBeNil)
	a.So(fp.LoraStandardChannel.GetDataRate(), should.NotBeNil)
	a.So(fp.LoraStandardChannel.GetDataRate().Index, should.Equal, 6)
	a.So(fp.FSKChannel, should.NotBeNil)
	a.So(fp.FSKChannel.GetDataRate(), should.NotBeNil)
	a.So(fp.FSKChannel.GetDataRate().Index, should.Equal, 7)
	a.So(fp.LBT, should.BeNil)
}

func TestUnmarshalUS(t *testing.T) {
	a := assertions.New(t)

	yamlDocument := `band-id: US_902_928
channels:
  - frequency: 903900000
  - frequency: 904100000
  - frequency: 904300000
  - frequency: 904500000
  - frequency: 904700000
  - frequency: 904900000
  - frequency: 905100000
  - frequency: 905300000
lora-standard-channel:
  frequency: 904600000
  data-rate:
    index: 4
dwell-time:
  uplinks: true
  duration: 400ms`

	fp := ttnpb.FrequencyPlan{}

	err := yaml.Unmarshal([]byte(yamlDocument), &fp)
	a.So(err, should.BeNil)

	for _, channel := range fp.Channels {
		a.So(channel.GetDataRate(), should.BeNil)
	}

	a.So(len(fp.Channels), should.Equal, 8)
	a.So(fp.LoraStandardChannel, should.NotBeNil)
	a.So(fp.LoraStandardChannel.GetDataRate(), should.NotBeNil)
	a.So(fp.LoraStandardChannel.GetDataRate().Index, should.Equal, 4)
	a.So(fp.FSKChannel, should.BeNil)
	a.So(fp.LBT, should.BeNil)
	a.So(fp.DwellTime.Uplinks, should.BeTrue)
	a.So(fp.DwellTime.Downlinks, should.BeFalse)
	a.So(*fp.DwellTime.Duration, should.Equal, 400*time.Millisecond)
}

func TestUnmarshalJP(t *testing.T) {
	a := assertions.New(t)

	yamlDocument := `band-id: AS923
channels:
  - frequency: 922000000
  - frequency: 922200000
  - frequency: 922400000
  - frequency: 922600000
  - frequency: 922800000
  - frequency: 923000000
  - frequency: 923200000
  - frequency: 923400000
lora-standard-channel:
  frequency: 922100000
  data-rate:
    index: 6
fsk-channel:
  frequency: 921800000
  data-rate:
    index: 7
lbt:
  rssi-target: -80
  rssi-offset: -4
  scan-time: 128
dwell-time:
  uplinks: true
  downlinks: true
  duration: 400ms
time-off-air:
  duration: 90ms`

	fp := ttnpb.FrequencyPlan{}

	err := yaml.Unmarshal([]byte(yamlDocument), &fp)
	a.So(err, should.BeNil)

	for _, channel := range fp.Channels {
		a.So(channel.GetDataRate(), should.BeNil)
	}

	a.So(len(fp.Channels), should.Equal, 8)
	a.So(fp.LoraStandardChannel, should.NotBeNil)
	a.So(fp.LoraStandardChannel.GetDataRate(), should.NotBeNil)
	a.So(fp.LoraStandardChannel.GetDataRate().Index, should.Equal, 6)
	a.So(fp.FSKChannel, should.NotBeNil)
	a.So(fp.FSKChannel.GetDataRate(), should.NotBeNil)
	a.So(fp.FSKChannel.GetDataRate().Index, should.Equal, 7)
	a.So(fp.LBT, should.NotBeNil)

	a.So(fp.TimeOffAir, should.NotBeNil)
	a.So(fp.TimeOffAir.Duration, should.NotBeNil)
	a.So(*fp.TimeOffAir.Duration, should.Equal, time.Millisecond*90)

	a.So(fp.DwellTime.Uplinks, should.BeTrue)
	a.So(fp.DwellTime.Downlinks, should.BeTrue)
	a.So(*fp.DwellTime.Duration, should.Equal, time.Millisecond*400)
}

func TestUnmarshalKR(t *testing.T) {
	a := assertions.New(t)

	yamlDocument := `band-id: KR_920_923
channels:
- frequency: 922100000
- frequency: 922300000
- frequency: 922500000
- frequency: 922700000
- frequency: 922900000
- frequency: 923100000
- frequency: 923300000
lbt:
  rssi-target: -80
  scan-time: 128`

	fp := ttnpb.FrequencyPlan{}
	err := yaml.Unmarshal([]byte(yamlDocument), &fp)
	a.So(err, should.BeNil)

	a.So(len(fp.Channels), should.Equal, 7)
	a.So(fp.LBT.RSSITarget, should.Equal, -80)
	a.So(fp.LBT.ScanTime, should.Equal, 128)
}

func TestFailUnmarshal(t *testing.T) {
	a := assertions.New(t)

	wrongYamlDocument2 := `band: UNKNOWN_BAND
channels:
- frequency: 867100000
- frequency: 867300000
- frequency: 867500000
- frequency: 867700000
- frequency: 867800000
datarate-index: 7
- frequency: 867900000
- frequency: 868100000
- frequency: 868300000
- frequency: 868300000
datarate-index: 6
- frequency: 868500000`

	fp2 := ttnpb.FrequencyPlan{}
	err := yaml.Unmarshal([]byte(wrongYamlDocument2), &fp2)
	a.So(err, should.NotBeNil)
}

func TestMarshal(t *testing.T) {
	a := assertions.New(t)

	fp := ttnpb.FrequencyPlan{
		BandID: string(band.EU_863_870),
		LBT:    nil,
		Channels: []*ttnpb.FrequencyPlan_Channel{
			{
				Frequency: 868500000,
			},
		},
	}
	res, err := yaml.Marshal(fp)
	a.So(err, should.BeNil)

	match, err := regexp.Match(`channels:
- frequency: 868500000`, res)
	a.So(match, should.BeTrue)
	a.So(err, should.BeNil)

	match, err = regexp.Match(fmt.Sprintf("band-id: %s", band.EU_863_870), res)
	a.So(match, should.BeTrue)
	a.So(err, should.BeNil)
}

func TestValidFrequencyPlan(t *testing.T) {
	regularDwellTime := 400 * time.Millisecond
	for _, tc := range []struct {
		name string

		fp    ttnpb.FrequencyPlan
		valid bool
	}{
		{
			name:  "no dwell time restrictions",
			fp:    ttnpb.FrequencyPlan{},
			valid: true,
		},
		{
			name: "invalid frequency plan-level dwell time restrictions because no duration is indicated",
			fp: ttnpb.FrequencyPlan{
				DwellTime: &ttnpb.FrequencyPlan_DwellTime{
					Uplinks: true,
				},
			},
			valid: false,
		},
		{
			name: "valid dwell time restrictions",
			fp: ttnpb.FrequencyPlan{
				DwellTime: &ttnpb.FrequencyPlan_DwellTime{
					Uplinks:  true,
					Duration: &regularDwellTime,
				},
				Channels: []*ttnpb.FrequencyPlan_Channel{
					{
						Frequency: 100000,
						DwellTime: &ttnpb.FrequencyPlan_DwellTime{
							Duration: &regularDwellTime,
						},
					},
				},
			},
			valid: true,
		},
		{
			name: "channel dwell time flag but no duration indicated",
			fp: ttnpb.FrequencyPlan{
				Channels: []*ttnpb.FrequencyPlan_Channel{
					{
						Frequency: 200000,
					},
					{
						Frequency: 100000,
						DwellTime: &ttnpb.FrequencyPlan_DwellTime{
							Downlinks: true,
						},
					},
				},
			},
			valid: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := assertions.New(t)
			shouldHaveExpectedResult := should.NotBeNil
			if tc.valid {
				shouldHaveExpectedResult = should.BeNil
			}

			err := tc.fp.Validate()
			a.So(err, shouldHaveExpectedResult)
		})
	}
}

func TestDwellTime(t *testing.T) {
	regularDwellTime := time.Second

	type setting struct {
		name string

		fp ttnpb.FrequencyPlan

		downlink  bool
		frequency uint64
	}

	for _, tc := range []struct {
		setting

		duration time.Duration
		success  bool
	}{
		{
			setting: setting{
				name: "no dwell time, among registered channels",

				fp: ttnpb.FrequencyPlan{
					Channels: []*ttnpb.FrequencyPlan_Channel{{Frequency: 1000000}},
				},
				frequency: 1000000,
			},
			duration: time.Second,
			success:  true,
		},

		{
			setting: setting{
				name: "no dwell time, not in registered channels",

				fp: ttnpb.FrequencyPlan{
					Channels: []*ttnpb.FrequencyPlan_Channel{{Frequency: 1000000}},
				},
				frequency: 1200000,
			},
			duration: time.Second,
			success:  true,
		},
		{
			setting: setting{
				name: "no channel-level dwell time with frequency plan-level default time",

				fp: ttnpb.FrequencyPlan{
					Channels: []*ttnpb.FrequencyPlan_Channel{{Frequency: 1000000}},
					DwellTime: &ttnpb.FrequencyPlan_DwellTime{
						Duration: &regularDwellTime,
					},
				},
				frequency: 1000000,
			},
			duration: time.Second,
			success:  true,
		},
	} {
		res := t.Run(tc.name, func(t *testing.T) {
			a := assertions.New(t)
			result := tc.fp.RespectsDwellTime(tc.downlink, tc.frequency, tc.duration)
			a.So(result, should.Equal, tc.success)
		})
		if !res {
			t.FailNow()
		}
	}

	for _, success := range []bool{true, false} {
		duration := regularDwellTime * 2
		if success {
			duration = regularDwellTime / 2
		}
		successSuffix := ", unsuccessful"
		if success {
			successSuffix = ", successful"
		}
		for _, tc := range []setting{
			{
				name: "frequency plan-level dwell time",

				fp: ttnpb.FrequencyPlan{
					Channels: []*ttnpb.FrequencyPlan_Channel{{Frequency: 1000000}},
					DwellTime: &ttnpb.FrequencyPlan_DwellTime{
						Uplinks:  true,
						Duration: &regularDwellTime,
					},
				},

				frequency: 1000000,
			},
			{
				name: "channel-level dwell time",

				fp: ttnpb.FrequencyPlan{
					Channels: []*ttnpb.FrequencyPlan_Channel{
						{
							Frequency: 1000000,
							DwellTime: &ttnpb.FrequencyPlan_DwellTime{
								Uplinks:  true,
								Duration: &regularDwellTime,
							},
						},
					},
				},
				frequency: 1000000,
			},
			{
				name: "channel-level dwell time with frequency plan-level default time",

				fp: ttnpb.FrequencyPlan{
					Channels: []*ttnpb.FrequencyPlan_Channel{
						{
							Frequency: 1000000,
							DwellTime: &ttnpb.FrequencyPlan_DwellTime{
								Uplinks: true,
							},
						},
					},
					DwellTime: &ttnpb.FrequencyPlan_DwellTime{Duration: &regularDwellTime},
				},
				frequency: 1000000,
			},
			{
				name: "dwell time, not in registered channels",

				fp: ttnpb.FrequencyPlan{
					Channels: []*ttnpb.FrequencyPlan_Channel{{Frequency: 1000000}},
					DwellTime: &ttnpb.FrequencyPlan_DwellTime{
						Duration: &regularDwellTime,
						Uplinks:  true,
					},
				},
				frequency: 1200000,
			},
		} {
			res := t.Run(tc.name+successSuffix, func(t *testing.T) {
				a := assertions.New(t)
				result := tc.fp.RespectsDwellTime(tc.downlink, tc.frequency, duration)
				a.So(result, should.Equal, success)
			})
			if !res {
				t.FailNow()
			}
		}
	}
}
