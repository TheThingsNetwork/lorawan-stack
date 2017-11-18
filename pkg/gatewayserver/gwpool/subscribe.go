// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gwpool

import (
	"io"
	"sync"
	"sync/atomic"

	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

func (p *pool) Subscribe(gatewayInfo ttnpb.GatewayIdentifier, link PoolSubscription) chan *ttnpb.GatewayUp {
	upstreamChannel := make(chan *ttnpb.GatewayUp)
	downstreamChannel := make(outgoing)

	p.store.Store(gatewayInfo, downstreamChannel)

	var streamErr atomic.Value

	c := connection{
		Logger: p.logger.WithField("gateway_id", gatewayInfo.GatewayID),

		GatewayInfo: gatewayInfo,

		Link:      link,
		StreamErr: &streamErr,
	}

	upstreamChannel := make(chan *ttnpb.GatewayUp)
	downstreamChannel := make(chan *ttnpb.GatewayDown)

	p.store.Store(gatewayInfo, downstreamChannel)

	wg := &sync.WaitGroup{}
	// Receiving on the stream
	wg.Add(1)
	go p.receivingRoutine(c, upstreamChannel, wg)

	// Sending on the stream
	wg.Add(1)
	go p.sendingRoutine(c, downstreamChannel, wg)

	wg.Wait()
	return upstreamChannel
}

type connection struct {
	Logger log.Interface

	GatewayInfo ttnpb.GatewayIdentifier

	Link      PoolSubscription
	StreamErr *atomic.Value
}

func (p *pool) sendingRoutine(c connection, downstreamChannel outgoing, readySignal chan bool) {
	var signaledReady bool
	defer func() {
		if !signaledReady {
			readySignal <- true
		}
	}()
func (p *pool) sendingRoutine(c connection, downstreamChannel chan *ttnpb.GatewayDown, wg *sync.WaitGroup) {
	wg.Done()

	for {
		select {
		case <-c.Link.Context().Done():
			c.StreamErr.Store(c.Link.Context().Err())
			p.store.Remove(c.GatewayInfo)
			return
		case outgoingMessage, ok := <-downstreamChannel:
			if !ok {
				c.StreamErr.Store(io.EOF)
				return
			}
			err := c.Link.Send(outgoingMessage)
			if err != nil {
				c.StreamErr.Store(err)
				p.store.Remove(c.GatewayInfo)
				return
			}
			c.Logger.Debug("Sent outgoing message to the gateway")
		}
	}
}

func (p *pool) receivingRoutine(c connection, upstreamChannel chan *ttnpb.GatewayUp, wg *sync.WaitGroup) {
	defer close(upstreamChannel)
	wg.Done()

	for {
		streamErr := c.StreamErr.Load()
		if streamErr != nil {
			return
		}

		upstreamMessage, err := c.Link.Recv()
		if err != nil {
			return
		}
		c.Logger.Debug("Received incoming message")

		select {
		case upstreamChannel <- upstreamMessage:
		default:
			c.Logger.Debug("No handler for upstream message, dropping message")
		}
	}
}
