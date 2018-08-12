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

package io

import (
	"context"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/pkg/errorcontext"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/pkg/toa"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const bufferSize = 10

// Server represents the Gateway Server to gateway frontends.
type Server interface {
	// FillContext fills the given context and identifiers.
	FillContext(ctx context.Context, ids ttnpb.GatewayIdentifiers) (context.Context, ttnpb.GatewayIdentifiers, error)
	// Connect connects a gateway by its identifiers to the Gateway Server with its assumed rights, and returns a
	// Connection for traffic and control.
	Connect(ctx context.Context, ids ttnpb.GatewayIdentifiers, assumedRights ...ttnpb.Right) (*Connection, error)
	// GetFrequencyPlan gets the specified frequency plan by its identifier.
	GetFrequencyPlan(ctx context.Context, id string) (*ttnpb.FrequencyPlan, error)
	// ClaimDownlink claims the downlink path for the given gateway.
	ClaimDownlink(ctx context.Context, ids ttnpb.GatewayIdentifiers) error
	// UnclaimDownlink releases the claim of the downlink path for the given gateway.
	UnclaimDownlink(ctx context.Context, ids ttnpb.GatewayIdentifiers) error
	// HasDownlinkClaim returns whether the the given gateway has a downlink path claimed.
	HasDownlinkClaim(ctx context.Context, ids ttnpb.GatewayIdentifiers) (bool, error)
}

// Connection is a connection to a gateway managed by a frontend.
type Connection struct {
	ctx       context.Context
	cancelCtx errorcontext.CancelFunc

	gateway   *ttnpb.Gateway
	scheduler scheduling.Scheduler

	upCh     chan *ttnpb.UplinkMessage
	downCh   chan *ttnpb.DownlinkMessage
	statusCh chan *ttnpb.GatewayStatus

	observations   ttnpb.GatewayObservations
	observationsMu sync.RWMutex
}

// NewConnection instantiates a new gateway connection.
func NewConnection(ctx context.Context, gateway *ttnpb.Gateway, scheduler scheduling.Scheduler) *Connection {
	ctx, cancelCtx := errorcontext.New(ctx)
	return &Connection{
		ctx:       ctx,
		cancelCtx: cancelCtx,
		gateway:   gateway,
		scheduler: scheduler,
		upCh:      make(chan *ttnpb.UplinkMessage, bufferSize),
		downCh:    make(chan *ttnpb.DownlinkMessage, bufferSize),
		statusCh:  make(chan *ttnpb.GatewayStatus, bufferSize),
	}
}

// Context returns the connection context.
func (c *Connection) Context() context.Context { return c.ctx }

// Disconnect marks the connection as disconnected and cancels the context.
func (c *Connection) Disconnect(err error) {
	c.cancelCtx(err)
}

// Gateway returns the gateway entity.
func (c *Connection) Gateway() *ttnpb.Gateway { return c.gateway }

var errBufferFull = errors.DefineInternal("buffer_full", "buffer is full")

// HandleUp updates the gateway observation and sends the message to the upstream channel.
func (c *Connection) HandleUp(up *ttnpb.UplinkMessage) error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.upCh <- up:
	default:
		return errBufferFull
	}
	c.addUplinkObservation(up)
	return nil
}

// HandleStatus updates the gateway observation and sends the status to the status channel.
func (c *Connection) HandleStatus(status *ttnpb.GatewayStatus) error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.statusCh <- status:
	default:
		return errBufferFull
	}
	c.addStatusObservation(status)
	return nil
}

var errComputeTOA = errors.Define("compute_toa", "could not compute the time on air")

// SendDown schedules and sends a downlink message.
func (c *Connection) SendDown(down *ttnpb.DownlinkMessage) error {
	duration, err := toa.Compute(down.RawPayload, down.Settings)
	if err != nil {
		return errComputeTOA.WithCause(err)
	}

	span := scheduling.Span{
		Start:    scheduling.ConcentratorTime(down.TxMetadata.Timestamp),
		Duration: duration,
	}

	if err := c.scheduler.ScheduleAt(span, down.Settings.Frequency); err != nil {
		return err
	}

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.downCh <- down:
	default:
		return errBufferFull
	}
	c.addDownlinkObservation(down)
	return nil
}

// Up returns the upstream channel.
func (c *Connection) Up() <-chan *ttnpb.UplinkMessage {
	return c.upCh
}

// Down returns the downstream channel.
func (c *Connection) Down() <-chan *ttnpb.DownlinkMessage {
	return c.downCh
}

// Status returns the status channel.
func (c *Connection) Status() <-chan *ttnpb.GatewayStatus {
	return c.statusCh
}

// GetObservations returns the gateway observations.
func (c *Connection) GetObservations() ttnpb.GatewayObservations {
	c.observationsMu.RLock()
	observations := c.observations
	c.observationsMu.RUnlock()
	return observations
}

func (c *Connection) addUplinkObservation(msg *ttnpb.UplinkMessage) {
	now := time.Now().UTC()
	c.observationsMu.Lock()
	c.observations.LastUplinkReceivedAt = &now
	c.observationsMu.Unlock()
}

func (c *Connection) addStatusObservation(status *ttnpb.GatewayStatus) {
	now := time.Now().UTC()
	c.observationsMu.Lock()
	c.observations.LastStatus = status
	c.observations.LastStatusReceivedAt = &now
	c.observationsMu.Unlock()
}

func (c *Connection) addDownlinkObservation(msg *ttnpb.DownlinkMessage) {
	now := time.Now().UTC()
	c.observationsMu.Lock()
	c.observations.LastDownlinkReceivedAt = &now
	c.observationsMu.Unlock()
}
