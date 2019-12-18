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

// Package cpf implements the JSON configuration for the Common Packet Forwarder.
package cpf

import (
	"text/template"
)

// lorafwdTmpl is based on lorafwd.toml present in CPF 1.1.6 Kerlink Wirnet Station DOTA.
var lorafwdTmpl = template.Must(template.New("lorafwd.toml").Parse(`# The LoRa forwarder 1.1.1-1 configuration file.
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
{{ if .Gateway.ID }}id = 0x{{ printf "%s" .Gateway.ID.MarshalText }}{{ else }}#id = 0xFFFFFFFFFFFFFFFF{{ end }}

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
node = "{{ .GWMP.Node }}"

# The GWMP services can be a service name (see services(5)) or an integer and,
# in this case, refers to a network port.

# The service where the gateway should push uplink messages.
#
# Type:    string or integer
# Example: "https" or 1234
# Default: 0
#
service.uplink = {{ .GWMP.ServiceUplink }}

# The service where the gateway should pull downlink messages.
#
# Type:    string or integer
# Example: "https" or 1234
# Default: 0
#
service.downlink = {{ .GWMP.ServiceDownlink }}

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
