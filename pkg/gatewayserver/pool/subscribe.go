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
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/scheduling"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

func (p *Pool) Subscribe(gatewayID string, link Subscription, fp ttnpb.FrequencyPlan) (chan *ttnpb.GatewayUp, error) {
	c := connection{
		Logger: p.logger.WithField("gateway_id", gatewayID),

		GatewayID: gatewayID,

		Link:      link,
		StreamErr: &atomicError{},
	}

	upstreamChannel := make(chan *ttnpb.GatewayUp)
	downstreamChannel := make(chan *ttnpb.GatewayDown)

	scheduler, err := scheduling.FrequencyPlanScheduler(link.Context(), fp)
	if err != nil {
		return nil, err
	}

	entry := &gatewayStoreEntry{
		channel: downstreamChannel,

		scheduler:        scheduler,
		observations:     ttnpb.GatewayObservations{},
		observationsLock: sync.RWMutex{},
	}
	p.store.Store(gatewayID, entry)

	wg := &sync.WaitGroup{}
	// Receiving on the stream
	wg.Add(1)
	go p.receivingRoutine(c, entry, upstreamChannel, wg)

	// Sending on the stream
	wg.Add(1)
	go p.sendingRoutine(c, downstreamChannel, wg)

	wg.Wait()
	return upstreamChannel, nil
}

type connection struct {
	Logger log.Interface

	GatewayID string

	Link      Subscription
	StreamErr *atomicError
}

type atomicError struct {
	value atomic.Value
}

func (aerr *atomicError) Store(err error) {
	aerr.value.Store(err)
}

func (aerr *atomicError) Load() error {
	if v := aerr.value.Load(); v != nil {
		err, ok := v.(error)
		if !ok {
			panic("atomicError value is not error type")
		}
		return err
	}
	return nil
}

func (p *Pool) sendingRoutine(c connection, downstreamChannel chan *ttnpb.GatewayDown, wg *sync.WaitGroup) {
	wg.Done()

	ctx := c.Link.Context()
	for {
		if err := c.StreamErr.Load(); err != nil {
			p.logger.WithError(err).Warn("Error encountered on stream, closing sending routine")
			return
		}

		select {
		case <-ctx.Done():
			err := ctx.Err()
			c.Logger.WithError(err).Warn("Link context done, closing sending routine")
			c.StreamErr.Store(err)
			p.store.Remove(c.GatewayID)
			return
		case outgoingMessage, ok := <-downstreamChannel:
			if !ok {
				c.StreamErr.Store(io.EOF)
				return
			}
			err := c.Link.Send(outgoingMessage)
			if err != nil {
				c.StreamErr.Store(err)
				p.store.Remove(c.GatewayID)
				return
			}
			c.Logger.Debug("Sent downlink message to the gateway")
		}
	}
}

func (p *Pool) receivingRoutine(c connection, entry *gatewayStoreEntry, upstreamChannel chan *ttnpb.GatewayUp, wg *sync.WaitGroup) {
	defer close(upstreamChannel)
	wg.Done()

	ctx := c.Link.Context()
	for {
		if err := c.StreamErr.Load(); err != nil {
			p.logger.WithError(err).Warn("Error encountered on stream, closing receiving routine")
			return
		}

		upstreamMessage, err := c.Link.Recv()
		if err != nil {
			return
		}
		c.Logger.Debug("Received incoming message")

		p.addUpstreamObservations(entry, upstreamMessage)

		select {
		case <-ctx.Done():
			p.store.Remove(c.GatewayID)
			err := ctx.Err()
			c.Logger.WithError(err).Warn("Link context done, closing receiving routine")
			c.StreamErr.Store(err)
			return
		case upstreamChannel <- upstreamMessage:
		}
	}
}

func (p *Pool) addUpstreamObservations(entry *gatewayStoreEntry, up *ttnpb.GatewayUp) {
	entry.observationsLock.Lock()
	currentTime := time.Now()

	if up.GatewayStatus != nil {
		entry.observations.LastStatus = up.GatewayStatus
		entry.observations.LastStatusReceivedAt = &currentTime
		entry.observations.StatusCount = entry.observations.StatusCount + 1
	}

	if nbUplinks := len(up.UplinkMessages); nbUplinks > 0 {
		entry.observations.UplinkCount = entry.observations.UplinkCount + uint64(nbUplinks)
		entry.observations.LastUplinkReceivedAt = &currentTime
	}

	entry.observationsLock.Unlock()
}
