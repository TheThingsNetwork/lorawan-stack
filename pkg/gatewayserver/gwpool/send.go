// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gwpool

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

func (p *pool) Send(gatewayInfo ttnpb.GatewayIdentifier, downstream *ttnpb.GatewayDown) error {
	gateway, err := p.store.Fetch(gatewayInfo)
	if err != nil {
		return errors.New("No network link to this gateway")
	}

	select {
	case gateway <- downstream:
		return nil
	case <-time.After(p.sendTimeout):
		return errors.Errorf("Downlink could not be picked up by this gateway's sending routine in given time interval(%s)", p.sendTimeout)
	}
}
