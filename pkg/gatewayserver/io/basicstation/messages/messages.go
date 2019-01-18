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

package messages

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
)

var errFrequencyPlan = errors.Define("frequency_plan", "frequency plan")

// Definition of message types
const (
	// Upstream types for messages from the Gateway
	TypeUpstreamVersion              = "version"
	TypeUpstreamJoinRequest          = "jreq"
	TypeUpstreamJoinUplinkDataFrame  = "updf"
	TypeUpstreamProprietaryDataFrame = "propdf"
	TypeUpstreamTxConfirmation       = "dntxed"
	TypeUpstreamTimeSync             = "timesync"
	TypeUpstreamRemoteShell          = "rmtsh"

	// Downstream types for messages from the Network
	TypeDownstreamRouterConfig              = "router_config"
	TypeDownstreamDownlinkMessage           = "dnmsg"
	TypeDownstreamDownlinkMulticastSchedule = "dnsched"
	TypeDownstreamTimeSync                  = "timesync"
	TypeDownstreamRemoteCommand             = "runcmd"
	TypeDownstreamRemoteShell               = "rmtsh"

	configHardwareSpecPrefix            = "sx1301"
	configHardwareSpecNoOfConcentrators = "1"
)

// DataRates encodes the available datarates of the channel plan for the Station in the format below
// [0] -> SF (Spreading Factor; Range: 7...12 for LoRa, 0 for FSK)
// [1] -> BW (Bandwidth; 125/250/500 for LoRa, ignored for FSK)
// [2] -> DNONLY (Downlink Only; 1 = true, 0 = false)
type DataRates [16][3]int

// DiscoverQuery contains the unique identifier of the gateway.
// This message is sent by the gateway.
type DiscoverQuery struct {
	EUI EUI `json:"router"`
}

// DiscoverResponse contains the response to the discover query.
// This message is sent by the Gateway Server.
type DiscoverResponse struct {
	EUI   EUI    `json:"router"`
	Muxs  EUI    `json:"muxs,omitempty"`
	URI   string `json:"uri,omitempty"`
	Error string `json:"error,omitempty"`
}

// Type returns the message type of the given data.
func Type(data []byte) (string, error) {
	msg := struct {
		Type string `json:"msgtype"`
	}{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return "", err
	}
	return msg.Type, nil
}

// Version contains version information.
// This message is sent by the gateway.
type Version struct {
	Station  string   `json:"station"`
	Firmware string   `json:"firmware"`
	Package  string   `json:"package"`
	Model    string   `json:"model"`
	Protocol int      `json:"protocol"`
	Features []string `json:"features,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (v Version) MarshalJSON() ([]byte, error) {
	type Alias Version
	return json.Marshal(struct {
		Type string `json:"msgtype"`
		Alias
	}{
		Type:  TypeUpstreamVersion,
		Alias: Alias(v),
	})
}

// IsProduction checks the features field for "prod" and returns true if found.
// This is then used to set debug options in the router config
func (v Version) IsProduction() bool {
	if len(v.Features) == 0 {
		return false
	}
	for _, feature := range v.Features {
		if feature == "prod" {
			return true
		}
	}
	return false
}

// SX1301Config contains the concentrator configuration.
// TODO: Hamonize this with sx1301_conf from https://github.com/TheThingsIndustries/lorawan-stack/issues/408
type SX1301Config struct{}

// RouterConfig contains the router configuration.
// This message is sent by the Gateway Server.
type RouterConfig struct {
	NetID          []int        `json:"NetID"`
	JoinEUI        [][]int      `json:"JoinEui"`
	Region         string       `json:"region"`
	HardwareSpec   string       `json:"hwspec"`
	FrequencyRange []int        `json:"freq_range"`
	DataRates      DataRates    `json:"DRs"`
	SX1301Config   SX1301Config `json:"sx1301_conf"`

	// These are debug options to be unset in production gateways
	NoCCA       bool `json:"nocca"`
	NoDutyCycle bool `json:"nodc"`
	NoDwellTime bool `json:"nodwell"`
}

// GetRouterConfig returns the routerconfig message to be sent to the gateway
// TODO: Adapt to https://github.com/TheThingsIndustries/lorawan-stack/pull/1402
func GetRouterConfig(fp frequencyplans.FrequencyPlan, isProd bool) (RouterConfig, error) {
	if err := fp.Validate(); err != nil {
		return RouterConfig{}, errFrequencyPlan
	}

	cfg := RouterConfig{}

	// TODO: Set maximumally permissive values
	cfg.JoinEUI = nil
	cfg.NetID = nil

	s := strings.Split(fp.BandID, "_")
	if len(s) < 2 {
		return RouterConfig{}, errFrequencyPlan
	}
	cfg.Region = fmt.Sprintf("%s%s", s[0], s[1])
	if len(fp.Radios) == 0 {
		return RouterConfig{}, errFrequencyPlan
	}
	// TODO: Handle FP with multiple radios if necessary
	cfg.FrequencyRange = []int{int(fp.Radios[0].TxConfiguration.MinFrequency), int(fp.Radios[0].TxConfiguration.MaxFrequency)}

	// TODO: Figure out how to evaluate configHardwareSpecNoOfConcentrators
	cfg.HardwareSpec = fmt.Sprintf("%s/%s", configHardwareSpecPrefix, configHardwareSpecNoOfConcentrators)

	drs, err := getDataRatesFromBandID(fp.BandID)
	if err != nil {
		return RouterConfig{}, errFrequencyPlan
	}
	cfg.DataRates = drs

	if isProd {
		cfg.NoCCA = false
		cfg.NoDutyCycle = false
		cfg.NoDwellTime = false
	} else {
		cfg.NoCCA = true
		cfg.NoDutyCycle = true
		cfg.NoDwellTime = true
	}

	// TODO: Get sx1301 config https://github.com/TheThingsIndustries/lorawan-stack/issues/408
	cfg.SX1301Config = SX1301Config{}
	return cfg, nil
}

var dataRatePattern = regexp.MustCompile("[0-9]+")

// getDataRatesFromBandID parses the available datarates from the Frequency Plam into the LNS Format.
func getDataRatesFromBandID(id string) (DataRates, error) {
	band, err := band.GetByID(id)
	if err != nil {
		return DataRates{}, err
	}

	drs := DataRates{}
	// Set the default values
	for _, dr := range drs {
		dr[0] = -1
		dr[1] = 0
		dr[2] = 0
	}

	var i = 0
	for _, dr := range band.DataRates {
		if dr.Rate.LoRa != "" {
			// Ex: SF12BW125 -> dr[0] = 12, dr[1] = 125, dr[2] = 0
			s := dataRatePattern.FindAllString(dr.Rate.LoRa, -1)
			if len(s) != 2 {
				continue
			}
			drs[i][0], _ = strconv.Atoi(s[0])
			drs[i][1], _ = strconv.Atoi(s[1])
			i++
		} else if dr.Rate.FSK != 0 {
			drs[i][0] = 0 // must be set to 0 for FSK, the BW field is ignored.
			i++
		}
	}
	return drs, nil
}
