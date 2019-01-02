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

package gatewayserver

import "go.thethings.network/lorawan-stack/pkg/gatewayserver/io/udp"

// MQTTConfig contains MQTT configuration of the Gateway Server.
type MQTTConfig struct {
	Listen    string `name:"listen" description:"Address for the MQTT frontend to listen on"`
	ListenTLS string `name:"listen-tls" description:"Address for the MQTT frontend to listen on (with TLS)"`
}

// UDPConfig defines the UDP configuration of the Gateway Server.
type UDPConfig struct {
	udp.Config `name:",squash"`
	Listeners  map[string]string `name:"listeners" description:"Listen addresses with (optional) fallback frequency plan ID for non-registered gateways"`
}

// BasicStationConfig defines the Basic Station configuration of the Gateway Server.
type BasicStationConfig struct {
	FallbackFrequencyPlanID string `name:"fallback-frequency-plan-id" description:"Fallback frequency plan ID for non-registered gateways"`
}

// Config represents the Gateway Server configuration.
type Config struct {
	RequireRegisteredGateways bool `name:"require-registered-gateways" description:"Require the gateways to be registered in the Identity Server"`

	MQTT         MQTTConfig         `name:"mqtt"`
	MQTTV2       MQTTConfig         `name:"mqtt-v2"`
	UDP          UDPConfig          `name:"udp"`
	BasicStation BasicStationConfig `name:"basic-station"`
}
