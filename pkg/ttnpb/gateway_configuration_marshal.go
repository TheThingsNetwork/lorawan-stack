// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package ttnpb

import (
	"bytes"
	"encoding/json"
	fmt "fmt"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

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
func (c GlobalSX1301Config) MarshalJSON() ([]byte, error) {
	var m orderedMap
	m.add("lorawan_public", c.LorawanPublic)
	m.add("clksrc", c.ClockSource)
	m.add("antenna_gain", c.AntennaGain)
	if c.LbtConfig != nil {
		m.add("lbt_cfg", *c.LbtConfig)
	}
	for i, radio := range c.Radios {
		m.add(fmt.Sprintf("radio_%d", i), *radio)
	}
	for i, channel := range c.Channels {
		m.add(fmt.Sprintf("chan_multiSF_%d", i), *channel)
	}
	if c.LoraStandardChannel != nil {
		m.add("chan_Lora_std", *c.LoraStandardChannel)
	}
	if c.FskChannel != nil {
		m.add("chan_FSK", *c.FskChannel)
	}
	for i, lut := range c.TxLutConfigs {
		m.add(fmt.Sprintf("tx_lut_%d", i), *lut)
	}
	return json.Marshal(m)
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *GlobalSX1301Config) UnmarshalJSON(msg []byte) error {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(msg, &root); err != nil {
		return err
	}
	radioMap := make(map[int]GlobalSX1301Config_RFConfig)
	txLutMap := make(map[int]GlobalSX1301Config_TxLUTConfig)
	chanMap := make(map[int]GlobalSX1301Config_IFConfig)

	for key, value := range root {
		switch {
		case key == "lorawan_public":
			if err := json.Unmarshal(value, &c.LorawanPublic); err != nil {
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
			if err := json.Unmarshal(value, &c.LbtConfig); err != nil {
				return err
			}
		case key == "chan_Lora_std":
			if err := json.Unmarshal(value, &c.LoraStandardChannel); err != nil {
				return err
			}
		case key == "chan_FSK":
			if err := json.Unmarshal(value, &c.FskChannel); err != nil {
				return err
			}
		case strings.HasPrefix(key, "chan_multiSF_"):
			var channel GlobalSX1301Config_IFConfig
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
			var txLut GlobalSX1301Config_TxLUTConfig
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
			var radio GlobalSX1301Config_RFConfig
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

	c.Radios = make([]*GlobalSX1301Config_RFConfig, len(radioMap))
	c.TxLutConfigs = make([]*GlobalSX1301Config_TxLUTConfig, len(txLutMap))
	c.Channels = make([]*GlobalSX1301Config_IFConfig, len(chanMap))

	for key, value := range radioMap {
		c.Radios[key] = &value
	}
	for key, value := range txLutMap {
		c.TxLutConfigs[key] = &value
	}
	for key, value := range chanMap {
		c.Channels[key] = &value
	}

	return nil
}

// MarshalJSON implements json.Marshaler
func (c *GlobalSX1301Config_LBTConfig) MarshalJSON() ([]byte, error) {
	if !c.Enable {
		return []byte(`{"enable": false}`), nil
	}
	return json.Marshal(struct {
		Enable         bool                                             `json:"enable"`
		RSSITarget     float32                                          `json:"rssi_target"`
		ChannelConfigs []*GlobalSX1301Config_LBTConfig_LBTChannelConfig `json:"chan_cfg"`
		RSSIOffset     float32                                          `json:"sx127x_rssi_offset"`
	}{
		Enable:         c.Enable,
		RSSITarget:     c.RssiTarget,
		RSSIOffset:     c.RssiOffset,
		ChannelConfigs: c.LbtChannelConfigs,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *GlobalSX1301Config_LBTConfig) UnmarshalJSON(msg []byte) error {
	type parseStruct struct {
		Enable         bool                                             `json:"enable"`
		RSSITarget     float32                                          `json:"rssi_target"`
		ChannelConfigs []*GlobalSX1301Config_LBTConfig_LBTChannelConfig `json:"chan_cfg"`
		RSSIOffset     float32                                          `json:"sx127x_rssi_offset"`
	}
	var parser parseStruct
	if err := json.Unmarshal(msg, &parser); err != nil {
		return err
	}

	c.Enable = parser.Enable
	c.RssiTarget = parser.RSSITarget
	c.RssiOffset = parser.RSSIOffset
	c.LbtChannelConfigs = parser.ChannelConfigs
	return nil
}

// MarshalJSON implements json.Marshaler
func (c *GlobalSX1301Config_LBTConfig_LBTChannelConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Frequency            uint64 `json:"freq_hz"`
		ScanTimeMicroseconds uint32 `json:"scan_time_us"`
	}{
		Frequency:            c.Frequency,
		ScanTimeMicroseconds: c.ScanTimeMicroseconds,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *GlobalSX1301Config_LBTConfig_LBTChannelConfig) UnmarshalJSON(msg []byte) error {
	type parseStruct struct {
		Frequency            uint64 `json:"freq_hz"`
		ScanTimeMicroseconds uint32 `json:"scan_time_us"`
	}
	var parser parseStruct
	if err := json.Unmarshal(msg, &parser); err != nil {
		return err
	}

	c.Frequency = parser.Frequency
	c.ScanTimeMicroseconds = parser.ScanTimeMicroseconds
	return nil
}

// MarshalJSON implements json.Marshaler
func (c GlobalSX1301Config_RFConfig) MarshalJSON() ([]byte, error) {
	if !c.Enable {
		return []byte(`{"enable": false}`), nil
	}
	return json.Marshal(struct {
		Enable      bool    `json:"enable"`
		Type        string  `json:"type,omitempty"`
		Frequency   uint64  `json:"freq"`
		RSSIOffset  float32 `json:"rssi_offset"`
		TxEnable    bool    `json:"tx_enable"`
		TxFreqMin   uint64  `json:"tx_freq_min,omitempty"`
		TxFreqMax   uint64  `json:"tx_freq_max,omitempty"`
		TxNotchFreq uint64  `json:"tx_notch_freq,omitempty"`
	}{
		Enable:      c.Enable,
		Type:        c.Type,
		Frequency:   c.Frequency,
		RSSIOffset:  c.RssiOffset,
		TxEnable:    c.TxEnable,
		TxFreqMin:   c.TxFreqMin,
		TxFreqMax:   c.TxFreqMax,
		TxNotchFreq: c.TxNotchFreq,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *GlobalSX1301Config_RFConfig) UnmarshalJSON(msg []byte) error {
	type parseStruct struct {
		Enable      bool    `json:"enable"`
		Type        string  `json:"type,omitempty"`
		Frequency   uint64  `json:"freq"`
		RSSIOffset  float32 `json:"rssi_offset"`
		TxEnable    bool    `json:"tx_enable"`
		TxFreqMin   uint64  `json:"tx_freq_min,omitempty"`
		TxFreqMax   uint64  `json:"tx_freq_max,omitempty"`
		TxNotchFreq uint64  `json:"tx_notch_freq,omitempty"`
	}
	var parser parseStruct
	if err := json.Unmarshal(msg, &parser); err != nil {
		return err
	}

	c.Enable = parser.Enable
	c.Type = parser.Type
	c.Frequency = parser.Frequency
	c.RssiOffset = parser.RSSIOffset
	c.TxEnable = parser.TxEnable
	c.TxFreqMin = parser.TxFreqMin
	c.TxFreqMax = parser.TxFreqMax
	c.TxNotchFreq = parser.TxNotchFreq

	return nil
}

// MarshalJSON implements json.Marshaler
func (c GlobalSX1301Config_IFConfig) MarshalJSON() ([]byte, error) {
	if !c.Enable {
		return []byte(`{"enable": false}`), nil
	}
	return json.Marshal(struct {
		Enable       bool   `json:"enable"`
		Radio        uint32 `json:"radio"`
		IFValue      int32  `json:"if"`
		Bandwidth    uint32 `json:"bandwidth,omitempty"`
		SpreadFactor uint32 `json:"spread_factor,omitempty"`
		Datarate     uint32 `json:"datarate,omitempty"`
	}{
		Enable:       c.Enable,
		Radio:        c.Radio,
		IFValue:      c.IfValue,
		Bandwidth:    c.Bandwidth,
		SpreadFactor: c.SpreadFactor,
		Datarate:     c.Datarate,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *GlobalSX1301Config_IFConfig) UnmarshalJSON(msg []byte) error {
	type parseStruct struct {
		Enable       bool   `json:"enable"`
		Radio        uint32 `json:"radio"`
		IFValue      int32  `json:"if"`
		Bandwidth    uint32 `json:"bandwidth,omitempty"`
		SpreadFactor uint32 `json:"spread_factor,omitempty"`
		Datarate     uint32 `json:"datarate,omitempty"`
	}
	var parser parseStruct
	if err := json.Unmarshal(msg, &parser); err != nil {
		return err
	}

	c.Enable = parser.Enable
	c.Radio = parser.Radio
	c.IfValue = parser.IFValue
	c.Bandwidth = parser.Bandwidth
	c.SpreadFactor = parser.SpreadFactor
	c.Datarate = parser.Datarate

	return nil
}

// MarshalJSON implements json.Marshaler
func (c GlobalSX1301Config_TxLUTConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		PAGain  int32 `json:"pa_gain"`
		MixGain int32 `json:"mix_gain"`
		RFPower int32 `json:"rf_power"`
		DigGain int32 `json:"dig_gain"`
	}{
		PAGain:  c.PaGain,
		MixGain: c.MixGain,
		RFPower: c.RfPower,
		DigGain: c.DigGain,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *GlobalSX1301Config_TxLUTConfig) UnmarshalJSON(msg []byte) error {
	type parseStruct struct {
		PAGain  int `json:"pa_gain"`
		MixGain int `json:"mix_gain"`
		RFPower int `json:"rf_power"`
		DigGain int `json:"dig_gain"`
	}
	var parser parseStruct
	if err := json.Unmarshal(msg, &parser); err != nil {
		return err
	}

	c.PaGain = int32(parser.PAGain)
	c.MixGain = int32(parser.MixGain)
	c.RfPower = int32(parser.RFPower)
	c.DigGain = int32(parser.DigGain)

	return nil
}
