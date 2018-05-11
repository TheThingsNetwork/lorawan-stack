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

package gatewayserver

import (
	"context"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/validate"
)

// ScheduleDownlink on a gateway connected to this Gateway Server.
//
// This request requires the GatewayIdentifier to have a GatewayID.
func (g *GatewayServer) ScheduleDownlink(ctx context.Context, down *ttnpb.DownlinkMessage) (*types.Empty, error) {
	id := down.TxMetadata.GatewayIdentifiers
	if err := validate.ID(id.GetGatewayID()); err != nil {
		return nil, err
	}

	g.connectionsMu.Lock()
	connection, ok := g.connections[id.UniqueID(ctx)]
	g.connectionsMu.Unlock()

	if !ok {
		return nil, ErrGatewayNotConnected.New(errors.Attributes{"gateway_id": id.GetGatewayID()})
	}
	downMessage := &ttnpb.GatewayDown{DownlinkMessage: down}
	err := connection.Send(downMessage)
	if err != nil {
		return nil, errors.NewWithCause(err, "Could not send downlink to gateway")
	}

	connection.addDownstreamObservations(downMessage)
	return ttnpb.Empty, nil
}
