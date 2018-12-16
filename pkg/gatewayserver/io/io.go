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
	"fmt"
	"sync/atomic"
	"time"

	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

const bufferSize = 10

// Server represents the Gateway Server to gateway frontends.
type Server interface {
	// FillGatewayContext fills the given context and identifiers.
	FillGatewayContext(ctx context.Context, ids ttnpb.GatewayIdentifiers) (context.Context, ttnpb.GatewayIdentifiers, error)
	// Connect connects a gateway by its identifiers to the Gateway Server, and returns a Connection for traffic and
	// control.
	Connect(ctx context.Context, protocol string, ids ttnpb.GatewayIdentifiers, fp *frequencyplans.FrequencyPlan, scheduler *scheduling.Scheduler) (*Connection, error)
	// GetFrequencyPlan gets the specified frequency plan by the gateway identifiers.
	GetFrequencyPlan(ctx context.Context, ids ttnpb.GatewayIdentifiers) (*frequencyplans.FrequencyPlan, error)
	// ClaimDownlink claims the downlink path for the given gateway.
	ClaimDownlink(ctx context.Context, ids ttnpb.GatewayIdentifiers) error
	// UnclaimDownlink releases the claim of the downlink path for the given gateway.
	UnclaimDownlink(ctx context.Context, ids ttnpb.GatewayIdentifiers) error
}

// Connection is a connection to a gateway managed by a frontend.
type Connection struct {
	// Align for sync/atomic.
	uplinks,
	downlinks uint64
	connectTime,
	lastStatusTime,
	lastUplinkTime,
	lastDownlinkTime int64
	lastStatus atomic.Value

	ctx       context.Context
	cancelCtx errorcontext.CancelFunc

	protocol  string
	gateway   *ttnpb.Gateway
	fp        *frequencyplans.FrequencyPlan
	scheduler *scheduling.Scheduler

	upCh     chan *ttnpb.UplinkMessage
	downCh   chan *ttnpb.DownlinkMessage
	statusCh chan *ttnpb.GatewayStatus
	txAckCh  chan *ttnpb.TxAcknowledgment
}

// NewConnection instantiates a new gateway connection.
func NewConnection(ctx context.Context, protocol string, gateway *ttnpb.Gateway, fp *frequencyplans.FrequencyPlan, scheduler *scheduling.Scheduler) *Connection {
	ctx, cancelCtx := errorcontext.New(ctx)
	return &Connection{
		ctx:         ctx,
		cancelCtx:   cancelCtx,
		protocol:    protocol,
		gateway:     gateway,
		fp:          fp,
		scheduler:   scheduler,
		upCh:        make(chan *ttnpb.UplinkMessage, bufferSize),
		downCh:      make(chan *ttnpb.DownlinkMessage, bufferSize),
		statusCh:    make(chan *ttnpb.GatewayStatus, bufferSize),
		txAckCh:     make(chan *ttnpb.TxAcknowledgment, bufferSize),
		connectTime: time.Now().UnixNano(),
	}
}

// Context returns the connection context.
func (c *Connection) Context() context.Context { return c.ctx }

// Disconnect marks the connection as disconnected and cancels the context.
func (c *Connection) Disconnect(err error) {
	c.cancelCtx(err)
}

// Protocol returns the protocol used for the connection, i.e. grpc, mqtt or udp.
func (c *Connection) Protocol() string { return c.protocol }

// Gateway returns the gateway entity.
func (c *Connection) Gateway() *ttnpb.Gateway { return c.gateway }

var errBufferFull = errors.DefineInternal("buffer_full", "buffer is full")

// HandleUp updates the uplink stats and sends the message to the upstream channel.
func (c *Connection) HandleUp(up *ttnpb.UplinkMessage) error {
	if up.Settings.Time != nil {
		c.scheduler.Sync(up.Settings.Timestamp, time.Now(), *up.Settings.Time)
	}

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.upCh <- up:
		atomic.AddUint64(&c.uplinks, 1)
		atomic.StoreInt64(&c.lastUplinkTime, time.Now().UnixNano())
	default:
		return errBufferFull
	}
	return nil
}

// HandleStatus updates the status stats and sends the status to the status channel.
func (c *Connection) HandleStatus(status *ttnpb.GatewayStatus) error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.statusCh <- status:
		c.lastStatus.Store(status)
		atomic.StoreInt64(&c.lastStatusTime, time.Now().UnixNano())
	default:
		return errBufferFull
	}
	return nil
}

// HandleTxAck sends the acknowledgment to the status channel.
func (c *Connection) HandleTxAck(ack *ttnpb.TxAcknowledgment) error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.txAckCh <- ack:
	default:
		return errBufferFull
	}
	return nil
}

var (
	errNotTxRequest    = errors.DefineInvalidArgument("not_tx_request", "downlink message is not a Tx request")
	errDownlinkPath    = errors.DefineInvalidArgument("downlink_path", "invalid downlink path: need exactly one option")
	errRxEmpty         = errors.DefineNotFound("rx_empty", "settings empty")
	errDataRate        = errors.DefineInvalidArgument("data_rate", "no data rate with index `{index}`")
	errDownlinkChannel = errors.DefineInvalidArgument("downlink_channel", "no downlink channel with frequency `{frequency}` Hz and data rate index `{data_rate_index}`")
	errTxSchedule      = errors.DefineAborted("tx_schedule", "schedule failed", "reason_rx1", "reason_rx2")
)

// SendDown schedules and sends a downlink message and updates the downlink stats.
// This method returns an error if the downlink message is not a Tx request or if the downlink path is not already
// reduced to one.
func (c *Connection) SendDown(down *ttnpb.DownlinkMessage) error {
	request := down.GetRequest()
	if request == nil {
		return errNotTxRequest
	}
	if len(request.DownlinkPaths) != 1 {
		return errDownlinkPath
	}
	downlinkPath := request.DownlinkPaths[0]
	// If the connection has no scheduler, scheduling is done by the gateway scheduler.
	// Otherwise, scheduling is done by the Gateway Server scheduler. This converts TxRequest to TxSettings.
	if c.scheduler != nil {
		band, err := band.GetByID(c.fp.BandID)
		if err != nil {
			return err
		}
		var schedulingErrors []error
		for i, rx := range []struct {
			dataRateIndex ttnpb.DataRateIndex
			frequency     uint64
			delay         time.Duration
		}{
			{
				dataRateIndex: request.Rx1DataRateIndex,
				frequency:     request.Rx1Frequency,
				delay:         time.Duration(request.Rx1Delay) * time.Second,
			},
			{
				dataRateIndex: request.Rx2DataRateIndex,
				frequency:     request.Rx2Frequency,
				delay:         time.Duration(request.Rx1Delay+1) * time.Second,
			},
		} {
			if rx.frequency == 0 {
				schedulingErrors = append(schedulingErrors, errRxEmpty)
				continue
			}
			dataRate := band.DataRates[rx.dataRateIndex].Rate
			if dataRate == types.EmptyDataRate {
				return errDataRate.WithAttributes("index", rx.dataRateIndex)
			}
			var found bool
			var channelIndex int
			for i, ch := range c.fp.DownlinkChannels {
				if ch.Frequency == rx.frequency &&
					rx.dataRateIndex >= ttnpb.DataRateIndex(ch.MinDataRate) &&
					rx.dataRateIndex <= ttnpb.DataRateIndex(ch.MaxDataRate) {
					channelIndex = i
					found = true
					break
				}
			}
			if !found {
				return errDownlinkChannel.WithAttributes(
					"frequency", rx.frequency,
					"data_rate_index", rx.dataRateIndex,
				)
			}
			settings := ttnpb.TxSettings{
				DataRateIndex: rx.dataRateIndex,
				Frequency:     rx.frequency,
				TxPower:       int32(band.DefaultMaxEIRP),
				ChannelIndex:  uint32(channelIndex),
			}
			if int(downlinkPath.AntennaIndex) < len(c.gateway.Antennas) {
				settings.TxPower -= int32(c.gateway.Antennas[downlinkPath.AntennaIndex].Gain)
			}
			if dataRate.LoRa != "" {
				settings.Modulation = ttnpb.Modulation_LORA
				bw, err := dataRate.Bandwidth()
				if err != nil {
					return err
				}
				settings.Bandwidth = bw
				sf, err := dataRate.SpreadingFactor()
				if err != nil {
					return err
				}
				settings.SpreadingFactor = uint32(sf)
				settings.CodingRate = "4/5"
				settings.InvertPolarization = true
			} else {
				settings.Modulation = ttnpb.Modulation_FSK
				settings.BitRate = dataRate.FSK
			}
			var f func(context.Context, int, ttnpb.TxSettings, ttnpb.TxSchedulePriority) (scheduling.Emission, error)
			switch t := request.Time.(type) {
			case *ttnpb.TxRequest_RelativeToUplink:
				f = c.scheduler.ScheduleAt
				settings.Timestamp = downlinkPath.Timestamp + uint32(rx.delay/time.Microsecond)
			case *ttnpb.TxRequest_Absolute:
				f = c.scheduler.ScheduleAt
				abs := *t.Absolute
				settings.Time = &abs
			case *ttnpb.TxRequest_Any:
				f = c.scheduler.ScheduleAnytime
			default:
				panic(fmt.Sprintf("proto: unexpected type %T in oneof", t))
			}
			em, err := f(c.ctx, len(down.RawPayload), settings, request.Priority)
			if err != nil {
				schedulingErrors = append(schedulingErrors, err)
				continue
			}
			schedulingErrors = nil
			down.Settings = &ttnpb.DownlinkMessage_Scheduled{
				Scheduled: &settings,
			}
			log.FromContext(c.ctx).WithFields(log.Fields(
				"rx_window", i+1,
				"frequency", rx.frequency,
				"data_rate", rx.dataRateIndex,
				"starts", em.Starts(),
				"duration", em.Duration(),
			)).Debug("Scheduled downlink")
			break
		}
		if schedulingErrors != nil {
			var kv []interface{}
			for i, err := range schedulingErrors {
				kv = append(kv, fmt.Sprintf("reason_rx%d", i+1), err.Error())
			}
			return errTxSchedule.WithAttributes(kv...)
		}
	}

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.downCh <- down:
		atomic.AddUint64(&c.downlinks, 1)
		atomic.StoreInt64(&c.lastDownlinkTime, time.Now().UnixNano())
	default:
		return errBufferFull
	}
	return nil
}

// Status returns the status channel.
func (c *Connection) Status() <-chan *ttnpb.GatewayStatus {
	return c.statusCh
}

// Up returns the upstream channel.
func (c *Connection) Up() <-chan *ttnpb.UplinkMessage {
	return c.upCh
}

// Down returns the downstream channel.
func (c *Connection) Down() <-chan *ttnpb.DownlinkMessage {
	return c.downCh
}

// TxAck returns the downlink acknowledgments channel.
func (c *Connection) TxAck() <-chan *ttnpb.TxAcknowledgment {
	return c.txAckCh
}

// ConnectTime returns the time the gateway connected.
func (c *Connection) ConnectTime() time.Time { return time.Unix(0, c.connectTime) }

// StatusStats returns the status statistics.
func (c *Connection) StatusStats() (last *ttnpb.GatewayStatus, t time.Time, ok bool) {
	if last, ok = c.lastStatus.Load().(*ttnpb.GatewayStatus); ok {
		t = time.Unix(0, atomic.LoadInt64(&c.lastStatusTime))
	}
	return
}

// UpStats returns the upstream statistics.
func (c *Connection) UpStats() (total uint64, t time.Time, ok bool) {
	total = atomic.LoadUint64(&c.uplinks)
	if ok = total > 0; ok {
		t = time.Unix(0, atomic.LoadInt64(&c.lastUplinkTime))
	}
	return
}

// DownStats returns the downstream statistics.
func (c *Connection) DownStats() (total uint64, t time.Time, ok bool) {
	total = atomic.LoadUint64(&c.downlinks)
	if ok = total > 0; ok {
		t = time.Unix(0, atomic.LoadInt64(&c.lastDownlinkTime))
	}
	return
}
