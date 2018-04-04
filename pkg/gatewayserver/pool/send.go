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

package pool

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/scheduling"
	"github.com/TheThingsNetwork/ttn/pkg/toa"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

var ErrGatewayNotConnected = &errors.ErrDescriptor{
	MessageFormat:  "Gateway `{gateway_id}` not connected",
	Code:           1,
	Type:           errors.NotFound,
	SafeAttributes: []string{"gateway_id"},
}

func init() {
	ErrGatewayNotConnected.Register()
}

func (p *pool) Send(gatewayInfo ttnpb.GatewayIdentifiers, downstream *ttnpb.GatewayDown) (err error) {
	if downstream == nil || downstream.DownlinkMessage == nil {
		return errors.New("No downlink")
	}

	gateway := p.store.Fetch(gatewayInfo)
	if gateway == nil {
		return ErrGatewayNotConnected.New(errors.Attributes{"gateway_id": gatewayInfo.GatewayID})
	}

	span := scheduling.Span{
		Start: scheduling.ConcentratorTime(downstream.DownlinkMessage.TxMetadata.Timestamp),
	}
	span.Duration, err = toa.Compute(downstream.DownlinkMessage.RawPayload, downstream.DownlinkMessage.Settings)
	if err != nil {
		return
	}

	err = gateway.scheduler.ScheduleAt(span, downstream.DownlinkMessage.Settings.Frequency)
	if err != nil {
		return
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
