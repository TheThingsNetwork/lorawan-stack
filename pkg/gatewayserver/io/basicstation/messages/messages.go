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

package messages

import "encoding/json"

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

const (
	TypeVersion      = "version"
	TypeRouterConfig = "router_config"
	TypeJoinRequest  = "jreq"
)

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
		Type:  TypeVersion,
		Alias: Alias(v),
	})
}

// SX1301Config contains the concentrator configuration.
type SX1301Config struct {
	/*
		{
			"radio_0": { .. } // same structure as radio_1
			"radio_1": {
				"enable": BOOL,
				"freq"  : INT
			},
			"chan_FSK": {
				"enable": BOOL
			},
			"chan_Lora_std": {
				"enable": BOOL,
				"radio": 0|1,
				"if": INT,
				"bandwidth": INT,
				"spread_factor": INT
			},
			"chan_multiSF_0": { .. }  // _0 .. _7 all have the same structure
			..
			"chan_multiSF_7": {
				"enable": BOOL,
				"radio": 0|1,
				"if": INT
			}
		} */
}

// RouterInfo contains router information.
// This message is sent by the Gateway Server.
type RouterInfo struct {
	NetID          []int        `json:"NetID"`
	JoinEUI        [][]int      `json:"JoinEui"`
	Region         string       `json:"region"`
	HardwareSpec   string       `json:"hwspec"`
	FrequencyRange []int        `json:"freq_range"`
	DataRates      [][]int      `json:"DRs"`
	SX1301Config   SX1301Config `json:"sx1301_conf"`
	NoCCA          bool         `json:"nocca"`
	NoDutyCycle    bool         `json:"nodc"`
	NoDwellTime    bool         `json:"nodwell"`
}
