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

package ws

import (
	"context"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// ParsedTime abstracts time values parsed from messages.
type ParsedTime struct {
	XTime   int64
	RefTime float64
}

// Formatter abstracts messages to/from websocket based gateways.
type Formatter interface {
	// GetRouterConfig parses version messages, generates router config (for downstream) and a status message (for upstream).
	GetRouterConfig(ctx context.Context, message []byte, bandID string, fps map[string]*frequencyplans.FrequencyPlan, receivedAt time.Time) (context.Context, []byte, *ttnpb.GatewayStatus, error)

	// FromDownlink generates a downlink byte stream that can be sent over the WS connection.
	FromDownlink(dids ttnpb.GatewayIdentifiers, rawPayload []byte, scheduledMsg *ttnpb.TxSettings, dlToken int64, dlTime time.Time, xTime int64) ([]byte, error)

	// ToUplink parses Uplink/JoinRequest messages into ttnpb.UplinkMessage.
	ToUplink(ctx context.Context, raw []byte, ids ttnpb.GatewayIdentifiers, bandID string, receivedAt time.Time, msgType string) (*ttnpb.UplinkMessage, ParsedTime, error)

	// ToTxAck parses fields from the TxConfirmation message and converts it to  ttnpb.TxAcknowledgment message.
	ToTxAck(ctx context.Context, message []byte, tokens io.DownlinkTokens, receivedAt time.Time) (*ttnpb.TxAcknowledgment, ParsedTime, error)
}
