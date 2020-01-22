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

// Package basicstationlns implements the JSON configuration for the Basic Station `router_config` message.
package basicstationlns

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/pfconfig/shared"
)

const (
	configHardwareSpecPrefix = "sx1301"
)

var errFrequencyPlan = errors.DefineInvalidArgument("frequency_plan", "invalid frequency plan `{name}`")

// DataRates encodes the available datarates of the channel plan for the Station in the format below:
// [0] -> SF (Spreading Factor; Range: 7...12 for LoRa, 0 for FSK)
// [1] -> BW (Bandwidth; 125/250/500 for LoRa, ignored for FSK)
// [2] -> DNONLY (Downlink Only; 1 = true, 0 = false)
type DataRates [16][3]int

// RouterConfig contains the router configuration.
// This message is sent by the Gateway Server.
type RouterConfig struct {
	NetID          []int                 `json:"NetID"`
	JoinEUI        [][]int               `json:"JoinEui"`
	Region         string                `json:"region"`
	HardwareSpec   string                `json:"hwspec"`
	FrequencyRange []int                 `json:"freq_range"`
	DataRates      DataRates             `json:"DRs"`
	SX1301Config   []shared.SX1301Config `json:"sx1301_conf"`

	// These are debug options to be unset in production gateways.
	NoCCA       bool `json:"nocca"`
	NoDutyCycle bool `json:"nodc"`
	NoDwellTime bool `json:"nodwell"`

	MuxTime float64 `json:"MuxTime"`
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

// GetRouterConfig returns the routerconfig message to be sent to the gateway.
// Currently as per the basic station docs, all frequency plans have to be from the same region (band) https://doc.sm.tc/station/tcproto.html#router-config-message.
func GetRouterConfig(bandID string, fps map[string]*frequencyplans.FrequencyPlan, isProd bool, dlTime time.Time) (RouterConfig, error) {
	for _, fp := range fps {
		if err := fp.Validate(); err != nil {
			return RouterConfig{}, errFrequencyPlan
		}
	}
	conf := RouterConfig{}
	conf.JoinEUI = nil
	conf.NetID = nil

	band, err := band.GetByID(bandID)
	if err != nil {
		return RouterConfig{}, errFrequencyPlan
	}
	s := strings.Split(band.ID, "_")
	if len(s) < 2 {
		return RouterConfig{}, errFrequencyPlan
	}
	conf.Region = fmt.Sprintf("%s%s", s[0], s[1])

	min, max, err := getMinMaxFrequencies(fps)
	conf.FrequencyRange = []int{
		int(min),
		int(max),
	}

	conf.HardwareSpec = fmt.Sprintf("%s/%d", configHardwareSpecPrefix, len(fps))

	drs, err := getDataRatesFromBandID(bandID)
	if err != nil {
		return RouterConfig{}, errFrequencyPlan
	}
	conf.DataRates = drs

	conf.NoCCA = !isProd
	conf.NoDutyCycle = !isProd
	conf.NoDwellTime = !isProd

	for _, fp := range fps {
		sx1301Conf, err := shared.BuildSX1301Config(fp)
		// These fields are not defined in the v1.5 ref design https://doc.sm.tc/station/gw_v1.5.html#rfconf-object and would cause a parsing error.
		sx1301Conf.Radios[0].TxFreqMin = 0
		sx1301Conf.Radios[0].TxFreqMax = 0
		// Remove hardware specific values that are not necessary.
		sx1301Conf.TxLUTConfigs = nil
		for i := range sx1301Conf.Radios {
			sx1301Conf.Radios[i].Type = ""
		}
		if err != nil {
			return RouterConfig{}, err
		}
		conf.SX1301Config = append(conf.SX1301Config, *sx1301Conf)
	}

	// Add the MuxTime for RTT measurement.
	conf.MuxTime = float64(dlTime.Unix()) + float64(dlTime.Nanosecond())/(1e9)

	return conf, nil
}

// getDataRatesFromBandID parses the available data rates from the band into DataRates.
func getDataRatesFromBandID(id string) (DataRates, error) {
	band, err := band.GetByID(id)
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

	for i, dr := range band.DataRates {
		if loraDR := dr.Rate.GetLoRa(); loraDR != nil {
			loraDR.GetSpreadingFactor()
			drs[i][0] = int(loraDR.GetSpreadingFactor())
			drs[i][1] = int(loraDR.GetBandwidth() / 1000)
		} else if fskDR := dr.Rate.GetFSK(); fskDR != nil {
			drs[i][0] = 0 // must be set to 0 for FSK, the BW field is ignored.
		}
	}
	return drs, nil
}

// getMinMaxFrequencies extract the minimum and maximum frequencies between all the bands.
func getMinMaxFrequencies(fps map[string]*frequencyplans.FrequencyPlan) (uint64, uint64, error) {
	var min, max uint64
	min = math.MaxUint64
	for _, fp := range fps {
		if len(fp.Radios) == 0 {
			return 0, 0, errFrequencyPlan
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
