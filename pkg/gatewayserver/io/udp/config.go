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

package udp

import (
	"time"
)

// RateLimitingConfig contains configuration settings for the rate limiting
// capabilities of the UDP gateway frontend firewall.
type RateLimitingConfig struct {
	Enable    bool          `name:"enable" description:"Enable rate limiting for gateways"`
	Messages  int           `name:"messages" description:"Number of past messages to check timestamp for"`
	Threshold time.Duration `name:"threshold" description:"Filter packet if timestamp is not newer than the older timestamps of the previous messages by this threshold"` //nolint:lll
}

// Config contains configuration settings for the UDP gateway frontend.
// Use DefaultConfig for recommended settings.
type Config struct {
	// PacketHandlers defines the number of concurrent packet handlers.
	PacketHandlers int `name:"packet-handlers" description:"Number of concurrent packet handlers"`
	// PacketBuffer defines how many packets are buffered to handlers before it overflows.
	PacketBuffer int `name:"packet-buffer" description:"Buffer size of unhandled packets"`
	// DownlinkPathExpires defines for how long a downlink path is valid. A downlink path is renewed on each pull data and
	// Tx acknowledgment packet.
	// Gateways typically pull data every 5 seconds.
	DownlinkPathExpires time.Duration `name:"downlink-path-expires" description:"Time after which a downlink path to a gateway expires"`
	// ConnectionExpires defines for how long a connection remains valid while no pull data, push data or Tx
	// acknowledgment is received.
	ConnectionExpires time.Duration `name:"connection-expires" description:"Time after which a connection of a gateway expires"`
	// ConnectionErrorExpires defines for how long an existing connection is cached by the Gateway Server when there is a connection error
	// before initiating a new connection. This ensures that packet handlers are not being stalled by gateways which cannot connect but
	// still attempt to do so.
	ConnectionErrorExpires time.Duration `name:"connection-error-expires" description:"Time after which a connection error of a gateway expires"`
	// ScheduleLateTime defines the time in advance to the actual transmission the downlink message should be scheduled to
	// the gateway.
	ScheduleLateTime time.Duration `name:"schedule-late-time" description:"Time in advance to send downlink to the gateway when scheduling late"`
	// AddrChangeBlock defines the time to block traffic when the address changes.
	AddrChangeBlock time.Duration `name:"addr-change-block" description:"Time to block traffic when a gateway's address changes"`
	// RateLimitingConfig is the configuration for the rate limiting firewall capabilities.
	RateLimiting RateLimitingConfig `name:"rate-limiting"`
}

// DefaultConfig contains the default configuration.
// We assume that the gateway sends a PULL_DATA message every 30 seconds, instead of the default of 5 seconds.
// This behavior has been observed in the wild, and is often used by gateways which use metered connections.
var DefaultConfig = Config{
	PacketHandlers:         1024,
	PacketBuffer:           50,
	DownlinkPathExpires:    90 * time.Second, // Expire downlink after missing typically 3 PULL_DATA messages.
	ConnectionExpires:      3 * time.Minute,  // Expire connection after missing typically 2 status messages.
	ConnectionErrorExpires: 5 * time.Minute,
	ScheduleLateTime:       800 * time.Millisecond,
	AddrChangeBlock:        0, // Release address when the connection expires.
	RateLimiting: RateLimitingConfig{
		Enable:    true,
		Messages:  10,
		Threshold: 10 * time.Millisecond,
	},
}
