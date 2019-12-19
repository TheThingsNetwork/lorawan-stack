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

package cpf_test

import (
	"fmt"
	"testing"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	. "go.thethings.network/lorawan-stack/pkg/pfconfig/cpf"
	"go.thethings.network/lorawan-stack/pkg/pfconfig/shared"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestBuildLorad(t *testing.T) {
	fps := frequencyplans.NewStore(test.FrequencyPlansFetcher)

	sx1301Config := func(fpID string) shared.SX1301Config {
		return *test.Must(shared.BuildSX1301Config(test.Must(fps.GetByID(fpID)).(*frequencyplans.FrequencyPlan))).(*shared.SX1301Config)
	}

	for _, tc := range []struct {
		Name           string
		Gateway        *ttnpb.Gateway
		Config         *LoradConfig
		ErrorAssertion func(t *testing.T, err error) bool
	}{
		{
			Name:    "Empty gateway",
			Gateway: &ttnpb.Gateway{},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeError)
			},
		},
		{
			Name: "EU868/No antennas",
			Gateway: &ttnpb.Gateway{
				FrequencyPlanID: test.EUFrequencyPlanID,
			},
			Config: &LoradConfig{
				SX1301Conf: LoradSX1301Conf{
					SX1301Config:      sx1301Config(test.EUFrequencyPlanID),
					AntennaGainDesc:   "Antenna gain, in dBi",
					InsertionLoss:     0.5,
					InsertionLossDesc: "Insertion loss, in dBi",
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name: "EU868/1 antenna",
			Gateway: &ttnpb.Gateway{
				FrequencyPlanID: test.EUFrequencyPlanID,
				Antennas: []ttnpb.GatewayAntenna{
					{
						Gain: 4,
						Location: ttnpb.Location{
							Latitude:  0.42,
							Longitude: 42.42,
						},
					},
				},
			},
			Config: &LoradConfig{
				SX1301Conf: LoradSX1301Conf{
					SX1301Config: func() shared.SX1301Config {
						conf := sx1301Config(test.EUFrequencyPlanID)
						conf.AntennaGain = 4
						return conf
					}(),
					AntennaGainDesc:   "Antenna gain, in dBi",
					InsertionLoss:     0.5,
					InsertionLossDesc: "Insertion loss, in dBi",
				},
				GatewayConf: LoradGatewayConf{
					BeaconLatitude:  0.42,
					BeaconLongitude: 42.42,
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name: "EU868/3 antennas",
			Gateway: &ttnpb.Gateway{
				FrequencyPlanID: test.EUFrequencyPlanID,
				Antennas: []ttnpb.GatewayAntenna{
					{
						Gain: 4,
						Location: ttnpb.Location{
							Latitude:  0.42,
							Longitude: 42.42,
						},
					},
					{
						Gain: 5,
						Location: ttnpb.Location{
							Latitude:  0.43,
							Longitude: 42.43,
						},
					},
					{
						Gain: 2,
						Location: ttnpb.Location{
							Latitude:  -42,
							Longitude: 42,
						},
					},
				},
			},
			Config: &LoradConfig{
				SX1301Conf: LoradSX1301Conf{
					SX1301Config: func() shared.SX1301Config {
						conf := sx1301Config(test.EUFrequencyPlanID)
						conf.AntennaGain = 4
						return conf
					}(),
					AntennaGainDesc:   "Antenna gain, in dBi",
					InsertionLoss:     0.5,
					InsertionLossDesc: "Insertion loss, in dBi",
				},
				GatewayConf: LoradGatewayConf{
					BeaconLatitude:  0.42,
					BeaconLongitude: 42.42,
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			gtw := deepcopy.Copy(tc.Gateway).(*ttnpb.Gateway)
			conf, err := BuildLorad(gtw, fps)
			a.So(gtw, should.Resemble, tc.Gateway)
			if a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(conf, should.Resemble, tc.Config)
			}
		})
	}
}

func TestBuildLorafwd(t *testing.T) {
	const host = "test.example.com"
	eui := types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}

	for _, tc := range []struct {
		Name           string
		Gateway        *ttnpb.Gateway
		Config         *LorafwdConfig
		ErrorAssertion func(t *testing.T, err error) bool
	}{
		{
			Name:    "Empty gateway",
			Gateway: &ttnpb.Gateway{},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeError)
			},
		},
		{
			Name: "No EUI/address:host",
			Gateway: &ttnpb.Gateway{
				GatewayServerAddress: host,
			},
			Config: &LorafwdConfig{
				GWMP: LorafwdGWMPConfig{
					Node:            host,
					ServiceUplink:   shared.DefaultGatewayServerUDPPort,
					ServiceDownlink: shared.DefaultGatewayServerUDPPort,
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name: "EUI set/address:host",
			Gateway: &ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{
					EUI: &eui,
				},
				GatewayServerAddress: host,
			},
			Config: &LorafwdConfig{
				Gateway: LorafwdGatewayConfig{
					ID: &eui,
				},
				GWMP: LorafwdGWMPConfig{
					Node:            host,
					ServiceUplink:   shared.DefaultGatewayServerUDPPort,
					ServiceDownlink: shared.DefaultGatewayServerUDPPort,
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name: "EUI set/address:'host:port'",
			Gateway: &ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{
					EUI: &eui,
				},
				GatewayServerAddress: fmt.Sprintf("%s:%d", host, 42),
			},
			Config: &LorafwdConfig{
				Gateway: LorafwdGatewayConfig{
					ID: &eui,
				},
				GWMP: LorafwdGWMPConfig{
					Node:            host,
					ServiceUplink:   42,
					ServiceDownlink: 42,
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			gtw := deepcopy.Copy(tc.Gateway).(*ttnpb.Gateway)
			conf, err := BuildLorafwd(gtw)
			a.So(gtw, should.Resemble, tc.Gateway)
			if a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(conf, should.Resemble, tc.Config)
			}
		})
	}
}

func TestLorafwdConfigMarshalText(t *testing.T) {
	for _, tc := range []struct {
		Name           string
		Config         LorafwdConfig
		Text           string
		ErrorAssertion func(t *testing.T, err error) bool
	}{
		{
			Name: "no EUI;node:'thethings.example.com';uplink:1704;downlink:1706",
			Config: LorafwdConfig{
				GWMP: LorafwdGWMPConfig{
					Node:            "thethings.example.com",
					ServiceUplink:   1704,
					ServiceDownlink: 1706,
				},
			},
			Text: `# The LoRa forwarder 1.1.1-1 configuration file.
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
#id = 0xFFFFFFFFFFFFFFFF

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
node = "thethings.example.com"

# The GWMP services can be a service name (see services(5)) or an integer and,
# in this case, refers to a network port.

# The service where the gateway should push uplink messages.
#
# Type:    string or integer
# Example: "https" or 1234
# Default: 0
#
service.uplink = 1704

# The service where the gateway should pull downlink messages.
#
# Type:    string or integer
# Example: "https" or 1234
# Default: 0
#
service.downlink = 1706

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
`,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name: "EUI:0x42FF42FFFFFFFFFF;node:'thethings.example.com';uplink:1704;downlink:1706",
			Config: LorafwdConfig{
				Gateway: LorafwdGatewayConfig{
					ID: &types.EUI64{0x42, 0xff, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				GWMP: LorafwdGWMPConfig{
					Node:            "thethings.example.com",
					ServiceUplink:   1704,
					ServiceDownlink: 1706,
				},
			},
			Text: `# The LoRa forwarder 1.1.1-1 configuration file.
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
id = 0x42FF42FFFFFFFFFF

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
node = "thethings.example.com"

# The GWMP services can be a service name (see services(5)) or an integer and,
# in this case, refers to a network port.

# The service where the gateway should push uplink messages.
#
# Type:    string or integer
# Example: "https" or 1234
# Default: 0
#
service.uplink = 1704

# The service where the gateway should pull downlink messages.
#
# Type:    string or integer
# Example: "https" or 1234
# Default: 0
#
service.downlink = 1706

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
`,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			conf := deepcopy.Copy(tc.Config).(LorafwdConfig)
			b, err := conf.MarshalText()
			if a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(string(b), should.Equal, tc.Text)
			}
		})
	}
}
