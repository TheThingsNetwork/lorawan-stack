// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb_test

import (
	"fmt"
	"regexp"
	"testing"
	time "time"

	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
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
lora-std-channel:
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
lora-std-channel:
  frequency: 904600000
  data-rate:
    index: 4
dwell-time: 400ms`

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
	a.So(fp.DwellTime, should.NotBeNil)
	a.So(*fp.DwellTime, should.Equal, time.Millisecond*400)
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
lora-std-channel:
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
dwell-time: 4s
tx-timeoff-air:
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

	a.So(fp.DwellTime, should.NotBeNil)
	a.So(*fp.DwellTime, should.Equal, time.Second*4)
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
