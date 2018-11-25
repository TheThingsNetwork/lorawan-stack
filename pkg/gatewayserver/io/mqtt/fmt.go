// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package mqtt

import (
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/mqtt/topics"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Formatter formats downlink, uplink and status messages.
type Formatter interface {
	topics.Layout

	EncodeDownlink(down *ttnpb.DownlinkMessage) ([]byte, error)
	DecodeUplink(message []byte) (*ttnpb.UplinkMessage, error)
	DecodeStatus(message []byte) (*ttnpb.GatewayStatus, error)
	DecodeTxAck(message []byte) (*ttnpb.TxAcknowledgment, error)
}

var errNotSupported = errors.DefineFailedPrecondition("not_supported", "not supported")
