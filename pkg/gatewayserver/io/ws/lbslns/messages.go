// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package lbslns

import (
	"encoding/json"

	"go.thethings.network/lorawan-stack/v3/pkg/basicstation"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

// MessageType is the type of the message.
type MessageType string

// Definition of LoRa Basic Station message types.
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

var errNotSupported = errors.DefineFailedPrecondition("not_supported", "not supported")

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
