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

package messages

import (
	"encoding/json"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/basicstation"
)

// Definition of message types.
const (
	// Upstream types for messages from the Gateway.
	TypeUpstreamVersion              = "version"
	TypeUpstreamJoinRequest          = "jreq"
	TypeUpstreamUplinkDataFrame      = "updf"
	TypeUpstreamProprietaryDataFrame = "propdf"
	TypeUpstreamTxConfirmation       = "dntxed"
	TypeUpstreamTimeSync             = "timesync"
	TypeUpstreamRemoteShell          = "rmtsh"

	// Downstream types for messages from the Network
	TypeDownstreamDownlinkMessage           = "dnmsg"
	TypeDownstreamDownlinkMulticastSchedule = "dnsched"
	TypeDownstreamTimeSync                  = "timesync"
	TypeDownstreamRemoteCommand             = "runcmd"
	TypeDownstreamRemoteShell               = "rmtsh"
)

// DiscoverQuery contains the unique identifier of the gateway.
// This message is sent by the gateway.
type DiscoverQuery struct {
	EUI basicstation.EUI `json:"router"`
}

// DiscoverResponse contains the response to the discover query.
// This message is sent by the Gateway Server.
type DiscoverResponse struct {
	EUI   basicstation.EUI `json:"router"`
	Muxs  basicstation.EUI `json:"muxs,omitempty"`
	URI   string           `json:"uri,omitempty"`
	Error string           `json:"error,omitempty"`
}

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
	Station  string `json:"station"`
	Firmware string `json:"firmware"`
	Package  string `json:"package"`
	Model    string `json:"model"`
	Protocol int    `json:"protocol"`
	Features string `json:"features,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (v Version) MarshalJSON() ([]byte, error) {
	type Alias Version
	return json.Marshal(struct {
		Type string `json:"msgtype"`
		Alias
	}{
		Type:  TypeUpstreamVersion,
		Alias: Alias(v),
	})
}

// IsProduction checks the features field for "prod" and returns true if found.
// This is then used to set debug options in the router config.
func (v Version) IsProduction() bool {
	if v.Features == "" {
		return false
	}
	if strings.Contains(v.Features, "prod") {
		return true
	}
	return false
}
