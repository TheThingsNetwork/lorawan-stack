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

// Package lbslns implements the JSON configuration for the LoRa Basics Station `router_config` message.
package lbslns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/experimental"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/pfconfig/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

const configHardwareSpecPrefix = "sx1301"

// Based on
// https://github.com/lorabasics/basicstation/blob/ba4f85d80a438a5c2b659e568cd2d0f0de08e5a7/src/s2e.c#L973-L1041 .
// Note that versions 2.0.6 or higher will reject any downstream messages such as downlinks if the
// region ID is unknown, while older versions will be more forgiving.
// Non standard names are used for backwards compatibility reasons.
var bandIDToRegionID = map[string]string{
	band.EU_863_870: "EU863", // Non standard name, officially `EU868`.
	band.AS_923_4:   "IL915", // Non standard name, officially `AS923-4`.
	band.KR_920_923: "KR920",
	band.AS_923:     "AS923JP", // Non standard name, officially `AS923-1`.
	band.US_902_928: "US902",   // Non standard name, officially `US915`.
	band.AU_915_928: "AU915",
}

var referenceRegionNamesFeatureFlag = experimental.DefineFeature("gs.lbslns.reference_region_names", false)

var errFrequencyPlan = errors.DefineInvalidArgument("frequency_plan", "invalid frequency plan `{name}`")

type kv struct {
	key   string
	value any
}

type orderedMap struct {
	kv []kv
}

func (m *orderedMap) add(k string, v any) {
	m.kv = append(m.kv, kv{key: k, value: v})
}

func (m orderedMap) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("{")
	for i, kv := range m.kv {
		if i != 0 {
			b.WriteString(",")
		}
		key, err := json.Marshal(kv.key)
		if err != nil {
			return nil, err
		}
		b.Write(key)
		b.WriteString(":")
		val, err := json.Marshal(kv.value)
		if err != nil {
			return nil, err
		}
		b.Write(val)
	}
	b.WriteString("}")
	return b.Bytes(), nil
}

// DataRates encodes the available datarates of the channel plan for the Station in the format below:
// [0] -> SF (Spreading Factor; Range: 7...12 for LoRa, 0 for FSK)
// [1] -> BW (Bandwidth; 125/250/500 for LoRa, ignored for FSK)
// [2] -> DNONLY (Downlink Only; 1 = true, 0 = false).
type DataRates [16][3]int

// LBSRFConfig contains the configuration for one of the radios only fields used for LoRa Basics Station gateways.
// The other fields of RFConfig (in pkg/pfconfig/shared) are hardware specific and are left out here.
// - `type`, `rssi_offset`, `tx_enable` and `tx_notch_freq` are set in the gateway.
// - `tx_freq_min` and `tx_freq_max` are defined in the  `freq_range` parameter of `router_config`.
// - `antenna_gain` is defined by the user for each antenna.
type LBSRFConfig struct {
	Enable      bool   `json:"enable"`
	Frequency   uint64 `json:"freq"`
	AntennaGain int    `json:"antenna_gain,omitempty"`
}

// LBSSX1301Config contains the configuration for the SX1301 concentrator for
// the LoRa Basics Station `router_config` message.
// This structure incorporates modifications for the `v1.5` and `v2` concentrator reference.
// https://doc.sm.tc/station/gw_v1.5.html
// https://doc.sm.tc/station/gw_v2.html
// The fields `lorawan_public` and `clock_source` are omitted as they should be present in the gateway's `station.conf`.
type LBSSX1301Config struct {
	LBTConfig           *shared.LBTConfig
	Radios              []LBSRFConfig
	Channels            []shared.IFConfig
	LoRaStandardChannel *shared.IFConfig
	FSKChannel          *shared.IFConfig
}

// MarshalJSON implements json.Marshaler.
func (c LBSSX1301Config) MarshalJSON() ([]byte, error) {
	var m orderedMap
	if c.LBTConfig != nil {
		m.add("lbt_cfg", *c.LBTConfig)
	}
	for i, radio := range c.Radios {
		m.add(fmt.Sprintf("radio_%d", i), radio)
	}
	for i, channel := range c.Channels {
		m.add(fmt.Sprintf("chan_multiSF_%d", i), channel)
	}
	if c.LoRaStandardChannel != nil {
		m.add("chan_Lora_std", *c.LoRaStandardChannel)
	}
	if c.FSKChannel != nil {
		m.add("chan_FSK", *c.FSKChannel)
	}
	return json.Marshal(m)
}

// fromSX1301Conf updates fields from shared.SX1301Config.
func (c *LBSSX1301Config) fromSX1301Conf(sx1301Conf shared.SX1301Config, antennaGain int) {
	c.LoRaStandardChannel = sx1301Conf.LoRaStandardChannel
	c.FSKChannel = sx1301Conf.FSKChannel
	c.LBTConfig = sx1301Conf.LBTConfig

	for _, radio := range sx1301Conf.Radios {
		c.Radios = append(c.Radios, LBSRFConfig{
			Enable:      radio.Enable,
			Frequency:   radio.Frequency,
			AntennaGain: antennaGain,
		})
	}

	c.Channels = append(c.Channels, sx1301Conf.Channels...)
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *LBSSX1301Config) UnmarshalJSON(msg []byte) error {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(msg, &root); err != nil {
		return err
	}
	radioMap, chanMap := make(map[int]LBSRFConfig), make(map[int]shared.IFConfig)
	for key, value := range root {
		switch {
		case key == "lbt_cfg":
			if err := json.Unmarshal(value, &c.LBTConfig); err != nil {
				return err
			}
		case key == "chan_Lora_std":
			if err := json.Unmarshal(value, &c.LoRaStandardChannel); err != nil {
				return err
			}
		case key == "chan_FSK":
			if err := json.Unmarshal(value, &c.FSKChannel); err != nil {
				return err
			}
		case strings.HasPrefix(key, "chan_multiSF_"):
			var channel shared.IFConfig
			if err := json.Unmarshal(value, &channel); err != nil {
				return err
			}
			var index int
			if _, err := fmt.Sscanf(key, "chan_multiSF_%d", &index); err != nil {
				return err
			}
			chanMap[index] = channel
		case strings.HasPrefix(key, "radio_"):
			var radio LBSRFConfig
			if err := json.Unmarshal(value, &radio); err != nil {
				return err
			}
			var index int
			if _, err := fmt.Sscanf(key, "radio_%d", &index); err != nil {
				return err
			}
			radioMap[index] = radio
		}
	}

	c.Radios, c.Channels = make([]LBSRFConfig, len(radioMap)), make([]shared.IFConfig, len(chanMap))
	for key, value := range radioMap {
		c.Radios[key] = value
	}
	for key, value := range chanMap {
		c.Channels[key] = value
	}
	return nil
}

// BeaconingConfig contains class B beacon configuration.
type BeaconingConfig struct {
	DR     ttnpb.DataRateIndex `json:"DR"`
	Layout [3]int              `json:"layout"`
	Freqs  []uint64            `json:"freqs"`
}

// beaconLayouts contains the beacon layouts depending on the beacon spreading factor.
// This format is in conformance with L2 1.0.4.
var beaconLayouts = map[uint32][3]int{
	8:  {1, 7, 19},
	9:  {2, 8, 17},
	10: {3, 9, 19},
	11: {4, 10, 21},
	12: {5, 11, 23},
}

// RouterConfig contains the router configuration.
// This message is sent by the Gateway Server.
type RouterConfig struct {
	NetID          []int             `json:"NetID"`
	JoinEUI        [][]int           `json:"JoinEui"`
	Region         string            `json:"region"`
	HardwareSpec   string            `json:"hwspec"`
	FrequencyRange []int             `json:"freq_range"`
	DataRates      DataRates         `json:"DRs"`
	SX1301Config   []LBSSX1301Config `json:"sx1301_conf"`

	// These are debug options to be unset in production gateways.
	// The values are ignored for production gateways, as they produce warnings.
	// https://github.com/lorabasics/basicstation/blob/bd17e53ab1137de6abb5ae48d6f3d52f6c268299/src-linux/sys_linux.c#L728
	NoCCA       bool `json:"nocca,omitempty"`
	NoDutyCycle bool `json:"nodc,omitempty"`
	NoDwellTime bool `json:"nodwell,omitempty"`

	MuxTime float64 `json:"MuxTime"`

	Beacon *BeaconingConfig `json:"bcning"`
}

// MarshalJSON implements json.Marshaler.
func (conf RouterConfig) MarshalJSON() ([]byte, error) {
	type Alias RouterConfig
	return json.Marshal(struct {
		Type string `json:"msgtype"`
		Alias
	}{
		Type:  "router_config",
		Alias: Alias(conf),
	})
}

// RouterFeatures contains the features of the LBS router.
type RouterFeatures interface {
	IsProduction() bool
}

// GetRouterConfig returns the routerconfig message to be sent to the gateway.
// Currently as per the LBS docs, all frequency plans have to be from the same region (band).
// https://doc.sm.tc/station/tcproto.html#router-config-message.
func GetRouterConfig(
	ctx context.Context,
	bandID string,
	fps []*frequencyplans.FrequencyPlan,
	features RouterFeatures,
	dlTime time.Time,
	antennaGain int,
) (RouterConfig, error) {
	for _, fp := range fps {
		if err := fp.Validate(); err != nil {
			return RouterConfig{}, errFrequencyPlan.New()
		}
	}
	conf := RouterConfig{}
	conf.JoinEUI = nil
	conf.NetID = nil

	phy, err := band.GetLatest(bandID)
	if err != nil {
		return RouterConfig{}, errFrequencyPlan.New()
	}
	if regionID, ok := bandIDToRegionID[phy.ID]; ok && referenceRegionNamesFeatureFlag.GetValue(ctx) {
		conf.Region = regionID
	} else {
		s := strings.Split(phy.ID, "_")
		if len(s) < 2 {
			return RouterConfig{}, errFrequencyPlan.New()
		}
		conf.Region = fmt.Sprintf("%s%s", s[0], s[1])
	}

	min, max, err := getMinMaxFrequencies(fps)
	if err != nil {
		return RouterConfig{}, err
	}
	conf.FrequencyRange = []int{
		int(min),
		int(max),
	}

	conf.HardwareSpec = fmt.Sprintf("%s/%d", configHardwareSpecPrefix, len(fps))

	drs, err := getDataRatesFromBandID(bandID)
	if err != nil {
		return RouterConfig{}, errFrequencyPlan.New()
	}
	conf.DataRates = drs

	production := features.IsProduction()
	conf.NoCCA = !production
	conf.NoDutyCycle = !production
	conf.NoDwellTime = !production

	for _, fp := range fps {
		if len(fp.Radios) == 0 {
			continue
		}
		sx1301Conf, err := shared.BuildSX1301Config(fp)
		if err != nil {
			return RouterConfig{}, err
		}
		var lbsSX1301Config LBSSX1301Config
		lbsSX1301Config.fromSX1301Conf(*sx1301Conf, antennaGain)
		conf.SX1301Config = append(conf.SX1301Config, lbsSX1301Config)
	}

	// Add the MuxTime for RTT measurement.
	conf.MuxTime = float64(dlTime.Unix()) + float64(dlTime.Nanosecond())/(1e9)

	if len(phy.Beacon.Frequencies) > 0 {
		dr, ok := phy.DataRates[phy.Beacon.DataRateIndex]
		if !ok {
			panic("unreachable")
		}
		sf := dr.Rate.GetLora().GetSpreadingFactor()
		if sf == 0 {
			panic("unreachable")
		}
		conf.Beacon = &BeaconingConfig{
			DR:     phy.Beacon.DataRateIndex,
			Layout: beaconLayouts[sf],
			Freqs:  phy.Beacon.Frequencies,
		}
	}

	return conf, nil
}

// getDataRatesFromBandID parses the available data rates from the band into DataRates.
func getDataRatesFromBandID(id string) (DataRates, error) {
	phy, err := band.GetLatest(id)
	if err != nil {
		return DataRates{}, err
	}

	// Set the default values.
	drs := DataRates{}
	for _, dr := range drs {
		dr[0] = -1
		dr[1] = 0
		dr[2] = 0
	}

	for i, dr := range phy.DataRates {
		if loraDR := dr.Rate.GetLora(); loraDR != nil {
			drs[i][0] = int(loraDR.GetSpreadingFactor())
			drs[i][1] = int(loraDR.GetBandwidth() / 1000)
		} else if fskDR := dr.Rate.GetFsk(); fskDR != nil {
			drs[i][0] = 0 // must be set to 0 for FSK, the BW field is ignored.
		}
	}
	return drs, nil
}

// getMinMaxFrequencies extract the minimum and maximum frequencies between all the bands.
func getMinMaxFrequencies(fps []*frequencyplans.FrequencyPlan) (min uint64, max uint64, err error) {
	min = math.MaxUint64
	for _, fp := range fps {
		if len(fp.Radios) == 0 {
			return 0, 0, errFrequencyPlan.New()
		}
		if fp.Radios[0].TxConfiguration.MinFrequency < min {
			min = fp.Radios[0].TxConfiguration.MinFrequency
		}
		if fp.Radios[0].TxConfiguration.MaxFrequency > max {
			max = fp.Radios[0].TxConfiguration.MaxFrequency
		}
	}
	return min, max, nil
}
