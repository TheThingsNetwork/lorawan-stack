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

/*
The purpose of this file is to create custom marshalers that allow the config
definitions that are now on the ttnpb to be backwards compatible with the old
tags
*/

import (
	"bytes"
	"encoding/json"
	fmt "fmt"
	"strings"
	"text/template"

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
func (c SX1301Config) MarshalJSON() ([]byte, error) {
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

// UnmarshalJSON implements json.Unmarshaler
func (c *SX1301Config) UnmarshalJSON(msg []byte) error {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(msg, &root); err != nil {
		return err
	}
	radioMap := make(map[int]RFConfig)
	txLutMap := make(map[int]TxLUTConfig)
	chanMap := make(map[int]IFConfig)

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

	c.Radios = make([]*RFConfig, len(radioMap))
	c.TxLutConfigs = make([]*TxLUTConfig, len(txLutMap))
	c.Channels = make([]*IFConfig, len(chanMap))

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
func (c *LBTConfig) MarshalJSON() ([]byte, error) {
	if !c.Enable {
		return []byte(`{"enable": false}`), nil
	}
	return json.Marshal(struct {
		Enable         bool                `json:"enable"`
		RSSITarget     float32             `json:"rssi_target"`
		ChannelConfigs []*LBTChannelConfig `json:"chan_cfg"`
		RSSIOffset     float32             `json:"sx127x_rssi_offset"`
	}{
		Enable:         c.Enable,
		RSSITarget:     c.RssiTarget,
		RSSIOffset:     c.RssiOffset,
		ChannelConfigs: c.ChannelConfigs,
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (c *LBTConfig) UnmarshalJSON(msg []byte) error {
	type parseStruct struct {
		Enable         bool                `json:"enable"`
		RSSITarget     float32             `json:"rssi_target"`
		ChannelConfigs []*LBTChannelConfig `json:"chan_cfg"`
		RSSIOffset     float32             `json:"sx127x_rssi_offset"`
	}
	var parser parseStruct
	if err := json.Unmarshal(msg, &parser); err != nil {
		return err
	}

	c.Enable = parser.Enable
	c.RssiTarget = parser.RSSITarget
	c.RssiOffset = parser.RSSIOffset
	c.ChannelConfigs = parser.ChannelConfigs
	return nil
}

// MarshalJSON implements json.Marshaler
func (c *LBTChannelConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Frequency            uint64 `json:"freq_hz"`
		ScanTimeMicroseconds uint32 `json:"scan_time_us"`
	}{
		Frequency:            c.Frequency,
		ScanTimeMicroseconds: c.ScanTimeMicroseconds,
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (c *LBTChannelConfig) UnmarshalJSON(msg []byte) error {
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
func (c RFConfig) MarshalJSON() ([]byte, error) {
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

// UnmarshalJSON implements json.Unmarshaler
func (c *RFConfig) UnmarshalJSON(msg []byte) error {
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
func (c IFConfig) MarshalJSON() ([]byte, error) {
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

// UnmarshalJSON implements json.Unmarshaler
func (c *IFConfig) UnmarshalJSON(msg []byte) error {
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
func (c TxLUTConfig) MarshalJSON() ([]byte, error) {
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

// UnmarshalJSON implements json.Unmarshaler
func (c *TxLUTConfig) UnmarshalJSON(msg []byte) error {
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

/*
	Marshallers of the semtechudp definitions
*/

// MarshalJSON implements json.Marshaler
func (c SemtechUDPConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SX1301Conf  SX1301Config                   `json:"SX1301_conf"`
		GatewayConf SemtechUDPConfig_GatewayConfig `json:"gateway_conf"`
	}{
		SX1301Conf:  *c.Sx1301Config,
		GatewayConf: *c.GatewayConfig,
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (c *SemtechUDPConfig) UnmarshalJSON(msg []byte) error {
	type parseStruct struct {
		SX1301Conf  SX1301Config                   `json:"SX1301_conf"`
		GatewayConf SemtechUDPConfig_GatewayConfig `json:"gateway_conf"`
	}
	var parser parseStruct
	if err := json.Unmarshal(msg, &parser); err != nil {
		return err
	}

	c.Sx1301Config = &parser.SX1301Conf
	c.GatewayConfig = &parser.GatewayConf
	return nil
}

// MarshalJSON implements json.Marshaler
func (c SemtechUDPConfig_GatewayConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		GatewayID      string                            `json:"gateway_ID,omitempty"`
		ServerAddress  string                            `json:"server_address"`
		ServerPortUp   uint32                            `json:"serv_port_up"`
		ServerPortDown uint32                            `json:"serv_port_down"`
		Enabled        bool                              `json:"serv_enabled,omitempty"` // only used inside servers
		Servers        []*SemtechUDPConfig_GatewayConfig `json:"servers,omitempty"`
	}{
		GatewayID:      c.GatewayId.GetGatewayId(),
		ServerAddress:  c.ServerAddress,
		ServerPortUp:   c.ServerPortUp,
		ServerPortDown: c.ServerPortDown,
		Enabled:        c.Enabled,
		Servers:        c.Servers,
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (c *SemtechUDPConfig_GatewayConfig) UnmarshalJSON(msg []byte) error {
	type parseStruct struct {
		GatewayID      string                            `json:"gateway_ID,omitempty"`
		ServerAddress  string                            `json:"server_address"`
		ServerPortUp   uint32                            `json:"serv_port_up"`
		ServerPortDown uint32                            `json:"serv_port_down"`
		Enabled        bool                              `json:"serv_enabled,omitempty"` // only used inside servers
		Servers        []*SemtechUDPConfig_GatewayConfig `json:"servers,omitempty"`
	}
	var parser parseStruct
	if err := json.Unmarshal(msg, &parser); err != nil {
		return err
	}

	c.GatewayId = &GatewayIdentifiers{GatewayId: parser.GatewayID}
	c.ServerAddress = parser.ServerAddress
	c.ServerPortUp = parser.ServerPortUp
	c.ServerPortDown = parser.ServerPortDown
	c.Enabled = parser.Enabled
	c.Servers = parser.Servers

	return nil
}

/*
	Marshallers of the Lorad definitions
*/

// MarshalJSON implements json.Marshaler
func (c LoradConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SX1301Conf  *LoradConfig_LoradSX1301Config `json:"SX1301_conf"`
		GatewayConf *LoradConfig_GatewayConfig     `json:"gateway_conf"`
	}{
		SX1301Conf:  c.Sx1301Config,
		GatewayConf: c.GatewayConfig,
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (c *LoradConfig) UnmarshalJSON(msg []byte) error {
	type parseStruct struct {
		SX1301Conf  *LoradConfig_LoradSX1301Config `json:"SX1301_conf"`
		GatewayConf *LoradConfig_GatewayConfig     `json:"gateway_conf"`
	}
	var parser parseStruct
	if err := json.Unmarshal(msg, &parser); err != nil {
		return err
	}

	c.Sx1301Config = parser.SX1301Conf
	c.GatewayConfig = parser.GatewayConf

	return nil
}

// MarshalJSON implements json.Marshaler
func (c LoradConfig_LoradSX1301Config) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		*SX1301Config
		InsertionLoss     float32 `json:"insertion_loss"`
		InsertionLossDesc string  `json:"insertion_loss_desc,omitempty"`
		AntennaGainDesc   string  `json:"antenna_gain_desc,omitempty"`
	}{
		SX1301Config:      c.GlobalConfig,
		InsertionLoss:     c.InsertionLoss,
		InsertionLossDesc: c.InsertionLossDesc,
		AntennaGainDesc:   c.AntennaGainDesc,
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (c *LoradConfig_LoradSX1301Config) UnmarshalJSON(msg []byte) error {
	type parseStruct struct {
		*SX1301Config
		InsertionLoss     float32 `json:"insertion_loss"`
		InsertionLossDesc string  `json:"insertion_loss_desc,omitempty"`
		AntennaGainDesc   string  `json:"antenna_gain_desc,omitempty"`
	}
	var parser parseStruct
	if err := json.Unmarshal(msg, &parser); err != nil {
		return err
	}

	c.GlobalConfig = parser.SX1301Config
	c.InsertionLoss = parser.InsertionLoss
	c.InsertionLossDesc = parser.InsertionLossDesc
	c.AntennaGainDesc = parser.AntennaGainDesc

	return nil
}

// MarshalJSON implements json.Marshaler
func (c LoradConfig_GatewayConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		BeaconEnable    bool    `json:"beacon_enable"`
		BeaconPeriod    uint32  `json:"beacon_period,omitempty"`
		BeaconFreqHz    uint32  `json:"beacon_freq_hz,omitempty"`
		BeaconFreqNb    uint32  `json:"beacon_freq_nb,omitempty"`
		BeaconStep      uint32  `json:"beacon_step,omitempty"`
		BeaconDatarate  uint32  `json:"beacon_datarate,omitempty"`
		BeaconBwHz      uint32  `json:"beacon_bw_hz,omitempty"`
		BeaconPower     uint32  `json:"beacon_power,omitempty"`
		BeaconInfodesc  []byte  `json:"beacon_infodesc,omitempty"`
		BeaconLatitude  float64 `json:"beacon_latitude,omitempty"`
		BeaconLongitude float64 `json:"beacon_longitude,omitempty"`
	}{
		BeaconEnable:    c.BeaconEnable,
		BeaconPeriod:    c.BeaconPeriod,
		BeaconFreqHz:    c.BeaconFreqHz,
		BeaconFreqNb:    c.BeaconFreqNb,
		BeaconStep:      c.BeaconStep,
		BeaconDatarate:  c.BeaconDatarate,
		BeaconBwHz:      c.BeaconBwHz,
		BeaconPower:     c.BeaconPower,
		BeaconInfodesc:  c.BeaconInfodesc,
		BeaconLatitude:  c.BeaconLatitude,
		BeaconLongitude: c.BeaconLongitude,
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (c *LoradConfig_GatewayConfig) UnmarshalJSON(msg []byte) error {
	type parseStruct struct {
		BeaconEnable    bool    `json:"beacon_enable"`
		BeaconPeriod    uint32  `json:"beacon_period,omitempty"`
		BeaconFreqHz    uint32  `json:"beacon_freq_hz,omitempty"`
		BeaconFreqNb    uint32  `json:"beacon_freq_nb,omitempty"`
		BeaconStep      uint32  `json:"beacon_step,omitempty"`
		BeaconDatarate  uint32  `json:"beacon_datarate,omitempty"`
		BeaconBwHz      uint32  `json:"beacon_bw_hz,omitempty"`
		BeaconPower     uint32  `json:"beacon_power,omitempty"`
		BeaconInfodesc  []byte  `json:"beacon_infodesc,omitempty"`
		BeaconLatitude  float64 `json:"beacon_latitude,omitempty"`
		BeaconLongitude float64 `json:"beacon_longitude,omitempty"`
	}
	var parser parseStruct
	if err := json.Unmarshal(msg, &parser); err != nil {
		return err
	}

	c.BeaconEnable = parser.BeaconEnable
	c.BeaconPeriod = parser.BeaconPeriod
	c.BeaconFreqHz = parser.BeaconFreqHz
	c.BeaconFreqNb = parser.BeaconFreqNb
	c.BeaconStep = parser.BeaconStep
	c.BeaconDatarate = parser.BeaconDatarate
	c.BeaconBwHz = parser.BeaconBwHz
	c.BeaconPower = parser.BeaconPower
	c.BeaconInfodesc = parser.BeaconInfodesc
	c.BeaconLatitude = parser.BeaconLatitude
	c.BeaconLongitude = parser.BeaconLongitude
	return nil
}

/*
	Marshallers of the LoraFwd definitions
*/

func (c LoraFwdConfig) MarshalText() ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := lorafwdConfigTmpl.Execute(buf, c); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// lorafwdConfigTmpl is based on lorafwd.toml present in CPF 1.1.6 Kerlink Wirnet Station DOTA.
var lorafwdConfigTmpl = template.Must(template.New("lorafwd.toml").Parse(`# The LoRa forwarder 1.1.1-1 configuration file.
#
# This configuration file is formatted using the TOML v0.5.0 language:
#  https://github.com/toml-lang/toml/blob/master/versions/en/toml-v0.5.0.md

[ gateway ]

# The gateway identifier. Used to identify the gateway inside the network. This
# identifier is 64 bits long. It could be expressed in hexadecimal for better
# readability.
#
# Type:    integer
# Example: 1194684 or 0x123abc or 0o4435274 or 0b100100011101010111100
# Default: 0
#
{{ if .Gateway.GatewayId }}id = 0x{{ printf "%s" .Gateway.GatewayId.MarshalText }}{{ else }}#id = 0xFFFFFFFFFFFFFFFF{{ end }}

[ filter ]

# Whether or not an uplink message with a valid CRC will be forwarded.
#
# Type:    boolean
# Example: false
# Default: true
#
#crc.valid = true

# Whether or not an uplink message with an invalid CRC will be forwarded.
#
# Type:    boolean
# Example: true
# Default: false
#
#crc.invalid = false

# Whether or not an uplink message without CRC will be forwarded.
#
# Type:    boolean
# Example: true
# Default: false
#
#crc.none = false

# Whether or not a LoRaWAN downlink will be forwarded as an uplink message.
#
# Type:    boolean
# Example: true
# Default: false
#
#lorawan.downlink = false

[ database ]

# Whether or not a persistent database will store the incoming messages until
# they will be sent and acknowledged.
#
# Type:    boolean
# Example: true
# Default: false
#
#enable = true

# The maximum number of messages allowed to be stored in the database. When
# full the newest message will replace the oldest one.
#
# Type:    integer
# Example: 20000
# Default: 200
#
#limit.messages = 200

# The minimum delay between two database fetch. To allow incoming messages
# to be aggregated before to be sent.
#
# Type:    integer (in milliseconds)
# Example: 1000
# Default: 100
#
#delay.cooldown = 1000

[ gwmp ]

# The internet host where the gateway should connect. The node can be either a
# fully qualified domain name or an IP address (IPv4 or IPv6).
#
# Type:    string
# Example: "myhost.example.com" or "123.45.67.89" or "2001:db8::1234"
# Default: "localhost"
#
node = "{{ .Gwmp.Node }}"

# The GWMP services can be a service name (see services(5)) or an integer and,
# in this case, refers to a network port.

# The service where the gateway should push uplink messages.
#
# Type:    string or integer
# Example: "https" or 1234
# Default: 0
#
service.uplink = {{ .Gwmp.ServiceUplink }}

# The service where the gateway should pull downlink messages.
#
# Type:    string or integer
# Example: "https" or 1234
# Default: 0
#
service.downlink = {{ .Gwmp.ServiceDownlink }}

# The heartbeat period. Used to keep the firewall open.
#
# Type:    integer (in seconds)
# Example: 30
# Default: 10
#
#period.heartbeat = 10

# The statistics period. Used to send statistics.
#
# Type:    integer (in seconds)
# Example: 10
# Default: 30
#
#period.statistics = 30

# The number of timed out messages which will automatically trigger a network
# socket restart. Used to monitor the connection.
#
# Type:    boolean or integer (false = 0 = disabled) (true = 10)
# Example: 3
# Default: true
#
#autorestart = false

# The maximum datagram size for uplink messages. The datagram includes the GWMP
# header and payload.
#
# Type:    integer
# Example: 50000
# Default: 20000
#
#limit.datagram = 65507

# The endpoint to control the LoRa daemon. Used to request statistics.
#
# Type:    string
# Example: "tcp://localhost:3333"
# Default: "ipc:///var/run/lora/lorad"
#
#lorad.control = "ipc:///var/run/lora/lorad"

[ api ]

# The API use ZeroMQ as transport layer. More informations about ZeroMQ
# endpoints format can be found here:
#
# http://api.zeromq.org/4-2:zmq-connect

# The endpoints for the uplink channel. Used to receive uplink messages.
#
# Type:    string or array of strings
# Example: "tcp://localhost:1111"
# Default: "ipc:///var/run/lora/uplink"
#
#uplink = [ "ipc:///var/run/lora/uplink", "tcp://localhost:1111" ]

# The endpoints for the downlink channel. Used to send downlink messages.
#
# Type:    string or array of strings
# Example: "tcp://localhost:2222"
# Default: "ipc:///var/run/lora/downlink"
#
#downlink = [ "ipc:///var/run/lora/downlink", "tcp://localhost:2222" ]

# The endpoints for the control channel. Used to receive control request.
#
# Type:    string or array of strings
# Example: "tcp://eth0:4444"
# Default: "ipc:///var/run/lora/lorafwd"
#
#control = [ "ipc:///var/run/lora/lorafwd", "tcp://eth0:4444" ]

# The filters for the uplink channel. Used to subscribe to uplink messages.
#
# The filters can handle raw binary (by using unicode) or keywords. The special
# empty filter ("") subscribe to all incoming messages.
#
# Keywords are case-insensitive and are one of these:
# - "lora"
# - "gfsk" or "fsk"
# - "event" (for ease of use, lorafwd always subscribe to event messages)
#
# Type:    string or array of strings
# Example: [ "\u000A", "keyword" ]
# Default: ""
#
#filters = [ "lora", "gfsk" ]
`))
