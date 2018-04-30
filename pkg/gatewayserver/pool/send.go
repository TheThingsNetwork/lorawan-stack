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

var (
	// ErrGatewayNotConnected is returned when a send operation failed because a gateway is not connected.
	ErrGatewayNotConnected = &errors.ErrDescriptor{
		MessageFormat:  "Gateway `{gateway_id}` not connected",
		Code:           1,
		Type:           errors.NotFound,
		SafeAttributes: []string{"gateway_id"},
	}
	// ErrGatewayIDNotSpecified is returned when a send operation failed because no gateway ID was specified.
	ErrGatewayIDNotSpecified = &errors.ErrDescriptor{
		MessageFormat: "No Gateway ID specified",
		Code:          2,
		Type:          errors.InvalidArgument,
	}
)

func init() {
	ErrGatewayNotConnected.Register()
	ErrGatewayIDNotSpecified.Register()
}

func (p *Pool) Send(gatewayID string, downstream *ttnpb.GatewayDown) (err error) {
	if downstream == nil || downstream.DownlinkMessage == nil {
		return errors.New("No downlink")
	}

	gateway := p.store.Fetch(gatewayID)
	if gateway == nil {
		return ErrGatewayNotConnected.New(errors.Attributes{"gateway_id": gatewayID})
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

func (p *Pool) addDownstreamObservations(entry *gatewayStoreEntry, down *ttnpb.GatewayDown) {
	entry.observationsLock.Lock()

	currentTime := time.Now()
	entry.observations.LastDownlinkReceivedAt = &currentTime
	entry.observations.DownlinkCount = entry.observations.DownlinkCount + 1

	entry.observationsLock.Unlock()
}
