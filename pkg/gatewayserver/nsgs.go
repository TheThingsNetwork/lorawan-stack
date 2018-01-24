// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gatewayserver

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	google_protobuf2 "github.com/gogo/protobuf/types"
)

func (g *GatewayServer) ScheduleDownlink(ctx context.Context, down *ttnpb.DownlinkMessage) (*google_protobuf2.Empty, error) {
	err := g.gateways.Send(down.TxMetadata.GatewayIdentifier, &ttnpb.GatewayDown{DownlinkMessage: down})
	if err != nil {
		return nil, errors.NewWithCause("Could not send downlink to gateway", err)
	}

	return nil, nil
}
