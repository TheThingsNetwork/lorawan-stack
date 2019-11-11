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

// Package shared contains the configuration that is common to various gateway types.
package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
)

// SX1301Config contains the configuration for the SX1301 concentrator.
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

var errInvalidKey = errors.DefineInvalidArgument("invalid_key", "key `{key}` invalid")

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
	m.add("antenna_gain", c.AntennaGain)
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

// UnmarshalJSON implements json.Unmarshaler.
func (c *SX1301Config) UnmarshalJSON(msg []byte) error {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(msg, &root); err != nil {
		return err
	}
	radioMap, txLutMap, chanMap := make(map[int]RFConfig), make(map[int]TxLUTConfig), make(map[int]IFConfig)
	for key, value := range root {
		switch {
		case key == "lorawan_public":
			if err := json.Unmarshal(value, &c.LoRaWANPublic); err != nil {
				return err
			}
		case key == "antenna_gain":
			if err := json.Unmarshal(value, &c.AntennaGain); err != nil {
				return err
			}
		case key == "clksrc":
			if err := json.Unmarshal(value, &c.ClockSource); err != nil {
				return err
			}
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
			var channel IFConfig
			if err := json.Unmarshal(value, &channel); err != nil {
				return err
			}
			var index int
			if _, err := fmt.Sscanf(key, "chan_multiSF_%d", &index); err == nil {
				chanMap[index] = channel
			} else {
				return err
			}
		case strings.HasPrefix(key, "tx_lut_"):
			var txLut TxLUTConfig
			if err := json.Unmarshal(value, &txLut); err != nil {
				return err
			}
			var index int
			if _, err := fmt.Sscanf(key, "tx_lut_%d", &index); err == nil {
				txLutMap[index] = txLut
			} else {
				return err
			}
		case strings.HasPrefix(key, "radio_"):
			var radio RFConfig
			if err := json.Unmarshal(value, &radio); err != nil {
				return err
			}
			var index int
			if _, err := fmt.Sscanf(key, "radio_%d", &index); err == nil {
				radioMap[index] = radio
			} else {
				return err
			}
		}
	}

	c.Radios, c.TxLUTConfigs, c.Channels = make([]RFConfig, len(radioMap)), make([]TxLUTConfig, len(txLutMap)), make([]IFConfig, len(chanMap))
	for key, value := range radioMap {
		c.Radios[key] = value
	}
	for key, value := range txLutMap {
		c.TxLUTConfigs[key] = value
	}
	for key, value := range chanMap {
		c.Channels[key] = value
	}
	return nil
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
	Type        string  `json:"type,omitempty"`
	Frequency   uint64  `json:"freq"`
	RSSIOffset  float32 `json:"rssi_offset"`
	TxEnable    bool    `json:"tx_enable"`
	TxFreqMin   uint64  `json:"tx_freq_min,omitempty"`
	TxFreqMax   uint64  `json:"tx_freq_max,omitempty"`
	TxNotchFreq uint64  `json:"tx_notch_freq,omitempty"`
}

// IFConfig contains the configuration for one of the channels.
type IFConfig struct {
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
		return []byte(`{"enable": false}`), nil
	}
	return json.Marshal(struct {
		Enable       bool   `json:"enable"`
		Radio        uint8  `json:"radio"`
		IFValue      int32  `json:"if"`
		Bandwidth    uint32 `json:"bandwidth,omitempty"`
		SpreadFactor uint8  `json:"spread_factor,omitempty"`
		Datarate     uint32 `json:"datarate,omitempty"`
	}{
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
	PAGain  int `json:"pa_gain"`
	MixGain int `json:"mix_gain"`
	RFPower int `json:"rf_power"`
	DigGain int `json:"dig_gain"`
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

// BuildSX1301Config builds the SX1301 configuration for the given frequency plan.
func BuildSX1301Config(frequencyPlan *frequencyplans.FrequencyPlan) (*SX1301Config, error) {
	band, err := band.GetByID(frequencyPlan.BandID)
	if err != nil {
		return nil, err
	}

	conf := new(SX1301Config)

	conf.LoRaWANPublic = true
	conf.ClockSource = frequencyPlan.ClockSource

	if frequencyPlan.LBT != nil {
		lbtConfig := &LBTConfig{
			Enable:     true,
			RSSITarget: frequencyPlan.LBT.RSSITarget,
			RSSIOffset: frequencyPlan.LBT.RSSIOffset,
		}
		for i, channel := range frequencyPlan.DownlinkChannels {
			if i > 7 {
				break
			}
			lbtConfig.ChannelConfigs = append(lbtConfig.ChannelConfigs, LBTChannelConfig{
				Frequency:            channel.Frequency,
				ScanTimeMicroseconds: uint32(frequencyPlan.LBT.ScanTime / time.Microsecond),
			})
		}
		conf.LBTConfig = lbtConfig
	}

	conf.Radios = make([]RFConfig, len(frequencyPlan.Radios))
	for i, radio := range frequencyPlan.Radios {
		rfConfig := RFConfig{
			Enable:     radio.Enable,
			Type:       radio.ChipType,
			Frequency:  radio.Frequency,
			RSSIOffset: radio.RSSIOffset,
		}
		if radio.TxConfiguration != nil {
			rfConfig.TxEnable = true
			rfConfig.TxFreqMin = radio.TxConfiguration.MinFrequency
			rfConfig.TxFreqMax = radio.TxConfiguration.MaxFrequency
			if radio.TxConfiguration.NotchFrequency != nil {
				rfConfig.TxNotchFreq = *radio.TxConfiguration.NotchFrequency
			}
		}
		conf.Radios[i] = rfConfig
	}

	numChannels := len(frequencyPlan.UplinkChannels)
	if numChannels < 8 {
		numChannels = 8
	}
	conf.Channels = make([]IFConfig, numChannels)
	for i, channel := range frequencyPlan.UplinkChannels {
		ifConfig := IFConfig{
			Enable:  true,
			Radio:   channel.Radio,
			IFValue: int32(int64(channel.Frequency) - int64(conf.Radios[channel.Radio].Frequency)),
		}
		conf.Channels[i] = ifConfig
	}

	conf.LoRaStandardChannel = &IFConfig{Enable: false}
	if channel := frequencyPlan.LoRaStandardChannel; channel != nil {
		if lora := band.DataRates[channel.DataRate].Rate.GetLoRa(); lora != nil {
			conf.LoRaStandardChannel = &IFConfig{
				Enable:       true,
				Radio:        channel.Radio,
				IFValue:      int32(int64(channel.Frequency) - int64(conf.Radios[channel.Radio].Frequency)),
				Bandwidth:    lora.Bandwidth,
				SpreadFactor: uint8(lora.SpreadingFactor),
			}
		}
	}

	conf.FSKChannel = &IFConfig{Enable: false}
	if channel := frequencyPlan.FSKChannel; channel != nil {
		if fsk := band.DataRates[channel.DataRate].Rate.GetFSK(); fsk != nil {
			conf.FSKChannel = &IFConfig{
				Enable:    true,
				Radio:     channel.Radio,
				IFValue:   int32(int64(channel.Frequency) - int64(conf.Radios[channel.Radio].Frequency)),
				Bandwidth: 125000,
				Datarate:  fsk.BitRate,
			}
		}
	}

	conf.TxLUTConfigs = defaultTxLUTConfigs

	return conf, nil
}
