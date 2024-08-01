// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package ttigw

import (
	"time"
)

// Config defines the The Things Industries gateway protocol configuration of the Gateway Server.
type Config struct {
	WSPingInterval      time.Duration `name:"ws-ping-interval" description:"Interval to send websocket ping messages"`
	MissedPongThreshold int           `name:"missed-pong-threshold" description:"Number of consecutive missed pongs before disconnection. This value is used only if the gateway sends at least one pong"` //nolint:lll
}

// DefaultConfig is the default configuration for The Things Industries gateway protocol frontend.
var DefaultConfig = Config{
	WSPingInterval:      30 * time.Second,
	MissedPongThreshold: 2,
}
