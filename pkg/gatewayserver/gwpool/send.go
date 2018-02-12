// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gwpool

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/scheduling"
	"github.com/TheThingsNetwork/ttn/pkg/toa"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

func (p *pool) Send(gatewayInfo ttnpb.GatewayIdentifier, downstream *ttnpb.GatewayDown) error {
	if downstream == nil || downstream.DownlinkMessage == nil {
		return errors.New("No downlink")
	}

	gateway, err := p.store.Fetch(gatewayInfo)
	if err != nil {
		return errors.New("No network link to this gateway")
	}

	span := scheduling.Span{
		Start: scheduling.ConcentratorTime(downstream.DownlinkMessage.TxMetadata.Timestamp),
	}
	span.Duration, err = toa.Compute(downstream.DownlinkMessage.RawPayload, downstream.DownlinkMessage.Settings)
	if err != nil {
		return err
	}

	err = gateway.scheduler.ScheduleAt(span, downstream.DownlinkMessage.Settings.Frequency)
	if err != nil {
		return err
	}

	select {
	case gateway.channel <- downstream:
		p.addDownstreamObservations(gateway, downstream)
		return nil
	case <-time.After(p.sendTimeout):
		return errors.Errorf("Downlink could not be picked up by this gateway's sending routine in given time interval(%s)", p.sendTimeout)
	}
}

func (p *pool) addDownstreamObservations(entry *gatewayStoreEntry, down *ttnpb.GatewayDown) {
	entry.observationsLock.Lock()

	currentTime := time.Now()
	entry.observations.LastDownlinkReceivedAt = &currentTime
	entry.observations.DownlinkCount = entry.observations.DownlinkCount + 1

	entry.observationsLock.Unlock()
}
