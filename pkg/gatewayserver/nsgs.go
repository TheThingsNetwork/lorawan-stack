// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gatewayserver

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/gogo/protobuf/types"
)

func (g *GatewayServer) ScheduleDownlink(ctx context.Context, down *ttnpb.DownlinkMessage) (*types.Empty, error) {
	err := g.gateways.Send(down.TxMetadata.GatewayIdentifier, &ttnpb.GatewayDown{DownlinkMessage: down})
	if err != nil {
		return nil, errors.NewWithCause(err, "Could not send downlink to gateway")
	}

	return nil, nil
}
