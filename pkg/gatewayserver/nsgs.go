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

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/gogo/protobuf/types"
)

// ScheduleDownlink on a gateway connected to this gateway server.
func (g *GatewayServer) ScheduleDownlink(ctx context.Context, down *ttnpb.DownlinkMessage) (*types.Empty, error) {
	err := g.gateways.Send(down.TxMetadata.GatewayIdentifiers, &ttnpb.GatewayDown{DownlinkMessage: down})
	if err != nil {
		return nil, errors.NewWithCause(err, "Could not send downlink to gateway")
	}

	return nil, nil
}
