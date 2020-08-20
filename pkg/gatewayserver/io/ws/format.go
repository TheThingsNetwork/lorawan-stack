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

	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// EndPoint contains information on the WebSocket endpoint.
type EndPoint struct {
	Scheme  string
	Address string
	Prefix  string
}

// Formatter abstracts messages to/from websocket based gateways.
type Formatter interface {
	// HandleConnectionInfo handles connection information requests from web socket based protocols.
	// This function returns a byte stream that contains connection information (ex: scheme, host, port etc) or an error if applicable.
	HandleConnectionInfo(ctx context.Context, raw []byte, server io.Server, endPoint EndPoint, receivedAt time.Time) []byte
	// HandleUp handles upstream messages from web socket based gateways.
	// This function optionally returns a byte stream to be sent as response to the upstream message.
	HandleUp(ctx context.Context, raw []byte, ids ttnpb.GatewayIdentifiers, conn *io.Connection, receivedAt time.Time) ([]byte, error)
	// FromDownlink generates a downlink byte stream that can be sent over the WS connection.
	FromDownlink(ctx context.Context, uid string, down ttnpb.DownlinkMessage, concentratorTime scheduling.ConcentratorTime, dlTime time.Time) ([]byte, error)
}
