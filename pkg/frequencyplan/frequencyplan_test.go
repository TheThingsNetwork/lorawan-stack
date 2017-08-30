// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package frequencyplan

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	yaml "gopkg.in/yaml.v2"
)

func TestUnmarshalEU(t *testing.T) {
	a := assertions.New(t)

	yamlDocument := `band: EU_863_870
channels:
- frequency: 867100000
- frequency: 867300000
- frequency: 867500000
- frequency: 867700000
- frequency: 867800000
  datarate: 7
- frequency: 867900000
- frequency: 868100000
- frequency: 868300000
- frequency: 868300000
  datarate: 6
- frequency: 868500000
radios:
- frequency: 867500000
  tx:
    min_frequency: 863000000
    max_frequency: 870000000
- frequency: 868500000`

	fp := FrequencyPlan{}
	err := yaml.Unmarshal([]byte(yamlDocument), &fp)
	a.So(err, should.BeNil)

	a.So(len(fp.Channels), should.Equal, 10)
	a.So(fp.Channels[4].DataRateIndex, should.NotBeNil)
	a.So(*fp.Channels[4].DataRateIndex, should.Equal, 7)
	a.So(len(fp.Radios), should.Equal, 2)
	a.So(fp.Radios[0].TX.MinFrequency, should.Equal, 863000000)
	a.So(fp.Radios[0].TX.MaxFrequency, should.Equal, 870000000)
	a.So(fp.Radios[1].TX, should.BeNil)
	a.So(fp.LBT, should.BeNil)
}

func TestUnmarshalKR(t *testing.T) {
	a := assertions.New(t)

	yamlDocument := `band: KR_920_923
channels:
- frequency: 922100000
- frequency: 922300000
- frequency: 922500000
- frequency: 922700000
- frequency: 922900000
- frequency: 923100000
- frequency: 923300000
lbt:
  rssi_offset: -4
  rssi_target: -80
  scan_time: 128
radios:
- frequency: 922700000
  tx:
    min_frequency: 920900000
    max_frequency: 923300000
- frequency: 922700000`

	fp := FrequencyPlan{}
	err := yaml.Unmarshal([]byte(yamlDocument), &fp)
	a.So(err, should.BeNil)

	a.So(len(fp.Channels), should.Equal, 7)
	a.So(len(fp.Radios), should.Equal, 2)
	a.So(fp.Radios[0].TX.MinFrequency, should.Equal, 920900000)
	a.So(fp.Radios[0].TX.MaxFrequency, should.Equal, 923300000)
	a.So(fp.Radios[1].TX, should.BeNil)
	a.So(fp.LBT.RSSIOffset, should.Equal, -4)
	a.So(*fp.LBT.RSSITarget, should.Equal, -80)
	a.So(*fp.LBT.ScanTime, should.Equal, 128)
}

func TestFailUnmarshal(t *testing.T) {
	a := assertions.New(t)

	wrongYamlDocument := `band: UNKNOWN_BAND
	- frequency: 867100000
	- frequency: 867300000
	- frequency: 867500000
	- frequency: 867700000
	- frequency: 867800000
	  datarate: 7
	- frequency: 867900000
	- frequency: 868100000
	- frequency: 868300000
	- frequency: 868300000
	  datarate: 6
	- frequency: 868500000
	radios:
	- frequency: 867500000
	  tx:
		min_frequency: 863000000
		max_frequency: 870000000
	- frequency: 868500000`

	fp := FrequencyPlan{}
	err := yaml.Unmarshal([]byte(wrongYamlDocument), &fp)
	a.So(err, should.NotBeNil)

	wrongYamlDocument2 := `band: UNKNOWN_BAND
channels:
- frequency: 867100000
- frequency: 867300000
- frequency: 867500000
- frequency: 867700000
- frequency: 867800000
  datarate: 7
- frequency: 867900000
- frequency: 868100000
- frequency: 868300000
- frequency: 868300000
  datarate: 6
- frequency: 868500000
radios:
- frequency: 867500000
  tx:
    min_frequency: 863000000
    max_frequency: 870000000
- frequency: 868500000`

	fp2 := FrequencyPlan{}
	err = yaml.Unmarshal([]byte(wrongYamlDocument2), &fp2)
	a.So(err, should.NotBeNil)
}

func TestMarshal(t *testing.T) {
	a := assertions.New(t)

	fpBand, err := band.GetByID(band.EU_863_870)
	a.So(err, should.BeNil)

	fp := FrequencyPlan{
		Band: Band(fpBand),
		LBT:  nil,
		Radios: []Radio{
			{
				Frequency: 868300000,
			},
		},
		Channels: []Channel{
			{
				Frequency: 868500000,
			},
		},
	}
	res, err := yaml.Marshal(fp)
	a.So(err, should.BeNil)

	match, err := regexp.Match(`radios:
- frequency: 868300000`, res)
	a.So(match, should.BeTrue)
	a.So(err, should.BeNil)

	match, err = regexp.Match(`channels:
- frequency: 868500000`, res)
	a.So(match, should.BeTrue)
	a.So(err, should.BeNil)

	match, err = regexp.Match(fmt.Sprintf("band: %s", band.EU_863_870), res)
	a.So(match, should.BeTrue)
	a.So(err, should.BeNil)
}
