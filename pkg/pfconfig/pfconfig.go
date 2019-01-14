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

// Package pfconfig implements the JSON configuration of Semtech's reference Packet Forwarder.
package pfconfig

import (
	"bytes"
	"encoding/json"
	"fmt"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

// Config represents the full configuration for Semtech's reference Packet Forwarder.
type Config struct {
	SX1301Conf  SX1301Config `json:"SX1301_conf"`
	GatewayConf GatewayConf  `json:"gateway_conf"`
}

// SX1301Config contains the configuration for the SX1301 concentrator.
// It marshals to compliant JSON, but does currently **not** unmarshal from it.
type SX1301Config struct {
	LoRaWANPublic       bool
	ClockSource         uint8
	AntennaGain         float32
	LBTConfig           *LBTConfig
	Radios              []RFConfig
	Channels            []IFConfig
	LoRaStandardChannel *IFConfig
	FSKChannel          *IFConfig
	TxLUTConfigs        []TxLUTConfig
}

type kv struct {
	key   string
	value interface{}
}

type orderedMap struct {
	kv []kv
}

func (m *orderedMap) add(k string, v interface{}) {
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

// MarshalJSON implements json.Marshaler.
func (c SX1301Config) MarshalJSON() ([]byte, error) {
	var m orderedMap
	m.add("lorawan_public", c.LoRaWANPublic)
	m.add("clksrc", c.ClockSource)
	m.add("clksrc_desc", "radio_1 provides clock to concentrator for most devices except MultiTech. For MultiTech set to 0.")
	m.add("antenna_gain", c.AntennaGain)
	m.add("antenna_gain_desc", "antenna gain, in dBi")
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
	for i, lut := range c.TxLUTConfigs {
		m.add(fmt.Sprintf("tx_lut_%d", i), lut)
	}
	return json.Marshal(m)
}

var errUnmarshalNotImplemented = errors.DefineUnimplemented(
	"unmarshal_not_implemented",
	"unmarshaling SX1301 config is not implemented",
)

// UnmarshalJSON implements json.Unmarshaler. It currently just errors because
// unmarshaling is not supported.
func (c *SX1301Config) UnmarshalJSON(_ []byte) error {
	return errUnmarshalNotImplemented
}

// LBTConfig contains the configuration for listen-before-talk.
type LBTConfig struct {
	Enable         bool               `json:"enable"`
	RSSITarget     float32            `json:"rssi_target"`
	ChannelConfigs []LBTChannelConfig `json:"chan_cfg"`
	RSSIOffset     float32            `json:"sx127x_rssi_offset"`
}

// LBTChannelConfig contains the listen-before-talk configuration for a channel.
type LBTChannelConfig struct {
	Frequency            uint64 `json:"freq_hz"`
	ScanTimeMicroseconds uint32 `json:"scan_time_us"`
}

// RFConfig contains the configuration for one of the radios.
type RFConfig struct {
	Enable      bool    `json:"enable"`
	Type        string  `json:"type"`
	Frequency   uint64  `json:"freq"`
	RSSIOffset  float32 `json:"rssi_offset"`
	TxEnable    bool    `json:"tx_enable"`
	TxFreqMin   uint64  `json:"tx_freq_min,omitempty"`
	TxFreqMax   uint64  `json:"tx_freq_max,omitempty"`
	TxNotchFreq uint64  `json:"tx_notch_freq,omitempty"`
}

// IFConfig contains the configuration for one of the channels.
type IFConfig struct {
	Description  string `json:"desc"`
	Enable       bool   `json:"enable"`
	Radio        uint8  `json:"radio"`
	IFValue      int32  `json:"if"`
	Bandwidth    uint32 `json:"bandwidth,omitempty"`
	SpreadFactor uint8  `json:"spread_factor,omitempty"`
	Datarate     uint32 `json:"datarate,omitempty"`
}

// MarshalJSON implements json.Marshaler
func (c IFConfig) MarshalJSON() ([]byte, error) {
	if !c.Enable {
		return []byte(`{"desc": "disabled","enable": false}`), nil
	}
	return json.Marshal(struct {
		Description  string `json:"desc"`
		Enable       bool   `json:"enable"`
		Radio        uint8  `json:"radio"`
		IFValue      int32  `json:"if"`
		Bandwidth    uint32 `json:"bandwidth,omitempty"`
		SpreadFactor uint8  `json:"spread_factor,omitempty"`
		Datarate     uint32 `json:"datarate,omitempty"`
	}{
		Description:  c.Description,
		Enable:       c.Enable,
		Radio:        c.Radio,
		IFValue:      c.IFValue,
		Bandwidth:    c.Bandwidth,
		SpreadFactor: c.SpreadFactor,
		Datarate:     c.Datarate,
	})
}

// TxLUTConfig contains the configuration for the TX LUT ind
type TxLUTConfig struct {
	Description string `json:"desc"`
	PAGain      int    `json:"pa_gain"`
	MixGain     int    `json:"mix_gain"`
	RFPower     int    `json:"rf_power"`
	DigGain     int    `json:"dig_gain"`
}

var defaultTxLUTConfigs = []TxLUTConfig{
	{PAGain: 0, MixGain: 8, RFPower: -6},
	{PAGain: 0, MixGain: 10, RFPower: -3},
	{PAGain: 0, MixGain: 12, RFPower: 0},
	{PAGain: 1, MixGain: 8, RFPower: 3},
	{PAGain: 1, MixGain: 10, RFPower: 6},
	{PAGain: 1, MixGain: 12, RFPower: 10},
	{PAGain: 1, MixGain: 13, RFPower: 11},
	{PAGain: 2, MixGain: 9, RFPower: 12},
	{PAGain: 1, MixGain: 15, RFPower: 13},
	{PAGain: 2, MixGain: 10, RFPower: 14},
	{PAGain: 2, MixGain: 11, RFPower: 16},
	{PAGain: 3, MixGain: 9, RFPower: 20},
	{PAGain: 3, MixGain: 10, RFPower: 23},
	{PAGain: 3, MixGain: 11, RFPower: 25},
	{PAGain: 3, MixGain: 12, RFPower: 26},
	{PAGain: 3, MixGain: 14, RFPower: 27},
}

func init() {
	for i, config := range defaultTxLUTConfigs {
		config.Description = fmt.Sprintf("TX gain table, index %d", i)
		defaultTxLUTConfigs[i] = config
	}
}

// GatewayConf contains the configuration for the gateway's server connection.
type GatewayConf struct {
	ServerAddress  string        `json:"server_address"`
	ServerPortUp   uint32        `json:"serv_port_up"`
	ServerPortDown uint32        `json:"serv_port_down"`
	Enabled        bool          `json:"serv_enabled,omitempty"` // only used inside servers
	Servers        []GatewayConf `json:"servers,omitempty"`
}
