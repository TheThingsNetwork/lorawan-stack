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

// ServerInfo contains information on the WebSocket server.
type ServerInfo struct {
	Scheme  string
	Address string
}

// Endpoints contains the connection endpoints for the protocol
type Endpoints struct {
	ConnectionInfo string
	Traffic        string
}

// Formatter abstracts messages to/from websocket based gateways.
type Formatter interface {
	// Endpoints fetches the connection endpoints for the protocol.
	Endpoints() Endpoints
	// HandleConnectionInfo handles connection information requests from web socket based protocols.
	// This function returns a byte stream that contains connection information (ex: scheme, host, port etc) or an error if applicable.
	HandleConnectionInfo(
		ctx context.Context,
		raw []byte,
		server io.Server,
		serverInfo ServerInfo,
		assertAuth func(context.Context, *ttnpb.GatewayIdentifiers) error,
	) []byte
	// HandleUp handles upstream messages from web socket based gateways.
	// This function optionally returns a byte stream to be sent as response to the upstream message.
	HandleUp(
		ctx context.Context, raw []byte, ids *ttnpb.GatewayIdentifiers, conn *io.Connection, receivedAt time.Time,
	) ([]byte, error)
	// FromDownlink generates a downlink byte stream that can be sent over the WS connection.
	FromDownlink(ctx context.Context, down *ttnpb.DownlinkMessage, bandID string, dlTime time.Time) ([]byte, error)
	// TransferTime generates a spurious time transfer message for a particular server time.
	TransferTime(
		ctx context.Context, serverTime time.Time, gpsTime *time.Time, concentratorTime *scheduling.ConcentratorTime,
	) ([]byte, error)
}
