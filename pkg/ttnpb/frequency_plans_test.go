// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb_test

import (
	"fmt"
	"regexp"
	"testing"

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
- frequency: 867800000
  data-rate:
    index: 7
- frequency: 867900000
- frequency: 868100000
- frequency: 868300000
- frequency: 868300000
  data-rate:
    index: 6
- frequency: 868500000`

	fp := ttnpb.FrequencyPlan{}

	err := yaml.Unmarshal([]byte(yamlDocument), &fp)
	a.So(err, should.BeNil)

	for _, channelIndex := range []int{0, 1, 2, 3, 5, 6, 7, 9} {
		a.So(fp.Channels[channelIndex].GetDataRate(), should.BeNil)
	}

	a.So(len(fp.Channels), should.Equal, 10)
	a.So(fp.Channels[4].GetDataRate(), should.NotBeNil)
	a.So(fp.Channels[4].GetDataRate().Index, should.Equal, 7)
	a.So(fp.LBT, should.BeNil)
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
			&ttnpb.FrequencyPlan_Channel{
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
