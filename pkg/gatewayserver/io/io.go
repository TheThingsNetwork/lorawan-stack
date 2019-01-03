// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
)

const (
	bufferSize = 1 << 4
	maxRTTs    = 1 << 5
)

// Frontend provides supported features by the gateway frontend.
type Frontend interface {
	// Protocol returns the protocol used in the frontend.
	Protocol() string
	// HasScheduler indicates whether the gateway has a scheduler.
	// If so, downlink requests are sent to the gateway.
	// If not, the Gateway Server scheduler schedules the request, and the transmission settings are sent to the gateway.
	HasScheduler() bool
}

// Server represents the Gateway Server to gateway frontends.
type Server interface {
	// FillGatewayContext fills the given context and identifiers.
	FillGatewayContext(ctx context.Context, ids ttnpb.GatewayIdentifiers) (context.Context, ttnpb.GatewayIdentifiers, error)
	// Connect connects a gateway by its identifiers to the Gateway Server, and returns a Connection for traffic and
	// control.
	Connect(ctx context.Context, frontend Frontend, ids ttnpb.GatewayIdentifiers) (*Connection, error)
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
	rtts      *rtts

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
		rtts:        newRTTs(maxRTTs),
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

// HasScheduler returns whether the connection has a scheduler.
func (c *Connection) HasScheduler() bool { return c.scheduler != nil }

// Gateway returns the gateway entity.
func (c *Connection) Gateway() *ttnpb.Gateway { return c.gateway }

var errBufferFull = errors.DefineInternal("buffer_full", "buffer is full")

// HandleUp updates the uplink stats and sends the message to the upstream channel.
func (c *Connection) HandleUp(up *ttnpb.UplinkMessage) error {
	if up.Settings.Time != nil {
		c.scheduler.SyncWithGateway(up.Settings.Timestamp, up.ReceivedAt, *up.Settings.Time)
		log.FromContext(c.ctx).WithFields(log.Fields(
			"timestamp", up.Settings.Timestamp,
			"server_time", up.ReceivedAt,
			"gateway_time", *up.Settings.Time,
		)).Debug("Synchronized server and gateway absolute time")
	} else {
		c.scheduler.Sync(up.Settings.Timestamp, up.ReceivedAt)
		log.FromContext(c.ctx).WithFields(log.Fields(
			"timestamp", up.Settings.Timestamp,
			"server_time", up.ReceivedAt,
		)).Debug("Synchronized server absolute time only")
	}
	for _, md := range up.RxMetadata {
		if md.AntennaIndex != 0 {
			// TODO: Support downlink path to multiple antennas (https://github.com/TheThingsNetwork/lorawan-stack/issues/48)
			md.DownlinkPathConstraint = ttnpb.DOWNLINK_PATH_CONSTRAINT_NEVER
			continue
		}
		buf, err := UplinkToken(ttnpb.GatewayAntennaIdentifiers{
			GatewayIdentifiers: c.gateway.GatewayIdentifiers,
			AntennaIndex:       md.AntennaIndex,
		}, md.Timestamp)
		if err != nil {
			return err
		}
		md.UplinkToken = buf
		md.DownlinkPathConstraint = c.gateway.DownlinkPathConstraint
	}

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.upCh <- up:
		atomic.AddUint64(&c.uplinks, 1)
		atomic.StoreInt64(&c.lastUplinkTime, up.ReceivedAt.UnixNano())
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

// RecordRTT records the given round-trip time.
func (c *Connection) RecordRTT(d time.Duration) { c.rtts.Record(d) }

var (
	errNotAllowed       = errors.DefineFailedPrecondition("not_allowed", "downlink not allowed")
	errNotTxRequest     = errors.DefineInvalidArgument("not_tx_request", "downlink message is not a Tx request")
	errNoUplinkToken    = errors.DefineInvalidArgument("no_uplink_token", "no uplink token provided for class A downlink")
	errDownlinkPath     = errors.DefineInvalidArgument("downlink_path", "invalid downlink path")
	errRxEmpty          = errors.DefineFailedPrecondition("rx_empty", "settings empty")
	errRxWindowSchedule = errors.Define("rx_window_schedule", "schedule in Rx window `{window}` failed")
	errDataRate         = errors.DefineInvalidArgument("data_rate", "no data rate with index `{index}`")
	errTooLong          = errors.DefineInvalidArgument("too_long", "the payload length `{payload_length}` exceeds maximum `{maximum_length}` at data rate index `{data_rate_index}`")
	errTxSchedule       = errors.DefineAborted("tx_schedule", "failed to schedule")
)

func getDownlinkPath(path *ttnpb.DownlinkPath, class ttnpb.Class) (ids ttnpb.GatewayAntennaIdentifiers, uplinkTimestamp uint32, err error) {
	buf := path.GetUplinkToken()
	if buf == nil && class == ttnpb.CLASS_A {
		err = errNoUplinkToken
		return
	}
	if buf != nil {
		return ParseUplinkToken(buf)
	}
	fixed := path.GetFixed()
	if fixed == nil {
		err = errDownlinkPath
		return
	}
	return *fixed, 0, nil
}

// SendDown schedules and sends a downlink message by using the given path and updates the downlink stats.
// This method returns an error if the downlink message is not a Tx request.
func (c *Connection) SendDown(path *ttnpb.DownlinkPath, msg *ttnpb.DownlinkMessage) (time.Duration, error) {
	if c.gateway.DownlinkPathConstraint == ttnpb.DOWNLINK_PATH_CONSTRAINT_NEVER {
		return 0, errNotAllowed
	}
	request := msg.GetRequest()
	if request == nil {
		return 0, errNotTxRequest
	}
	var delay time.Duration
	// If the connection has no scheduler, scheduling is done by the gateway scheduler.
	// Otherwise, scheduling is done by the Gateway Server scheduler. This converts TxRequest to TxSettings.
	if c.scheduler != nil {
		logger := log.FromContext(c.ctx).WithField("class", request.Class)
		logger.Debug("Attempt to schedule downlink on gateway")
		ids, uplinkTimestamp, err := getDownlinkPath(path, request.Class)
		if err != nil {
			return 0, err
		}
		phy, err := band.GetByID(c.fp.BandID)
		if err != nil {
			return 0, err
		}
		var errRxDetails []interface{}
		for i, rx := range []struct {
			dataRateIndex ttnpb.DataRateIndex
			frequency     uint64
			delay         time.Duration
		}{
			{
				dataRateIndex: request.Rx1DataRateIndex,
				frequency:     request.Rx1Frequency,
				delay:         0,
			},
			{
				dataRateIndex: request.Rx2DataRateIndex,
				frequency:     request.Rx2Frequency,
				delay:         time.Second,
			},
		} {
			rx1Delay := time.Duration(request.Rx1Delay) * time.Second
			if rx1Delay == 0 {
				rx1Delay = time.Second // RX_DELAY_0 is valid, and 1 second.
			}
			rxDelay := rx1Delay + rx.delay
			if rx.frequency == 0 {
				errRxDetails = append(errRxDetails, errRxEmpty)
				continue
			}
			logger := logger.WithFields(log.Fields(
				"rx_window", i+1,
				"frequency", rx.frequency,
				"data_rate_index", rx.dataRateIndex,
			))
			logger.Debug("Attempt to schedule downlink in receive window")
			dataRate := phy.DataRates[rx.dataRateIndex].Rate
			if dataRate == (ttnpb.DataRate{}) {
				return 0, errDataRate.WithAttributes("index", rx.dataRateIndex)
			}
			// The maximum payload size is MACPayload only; for PHYPayload take MHDR (1 byte) and MIC (4 bytes) into account.
			maxPHYLength := phy.DataRates[rx.dataRateIndex].DefaultMaxSize.PayloadSize(c.fp.DwellTime.GetDownlinks()) + 5
			if len(msg.RawPayload) > int(maxPHYLength) {
				return 0, errTooLong.WithAttributes(
					"payload_length", len(msg.RawPayload),
					"maximum_length", maxPHYLength,
					"data_rate_index", rx.dataRateIndex,
				)
			}
			eirp := phy.DefaultMaxEIRP
			if sb, ok := phy.FindSubBand(rx.frequency); ok {
				eirp = sb.MaxEIRP
			}
			if c.fp.MaxEIRP != nil && *c.fp.MaxEIRP < eirp {
				eirp = *c.fp.MaxEIRP
			}
			settings := ttnpb.TxSettings{
				DataRateIndex: rx.dataRateIndex,
				Frequency:     rx.frequency,
				Downlink: &ttnpb.TxSettings_Downlink{
					TxPower:      eirp,
					AntennaIndex: ids.AntennaIndex,
				},
			}
			if int(ids.AntennaIndex) < len(c.gateway.Antennas) {
				settings.Downlink.TxPower -= c.gateway.Antennas[ids.AntennaIndex].Gain
			}
			settings.DataRate = dataRate
			if dr := dataRate.GetLoRa(); dr != nil {
				settings.CodingRate = phy.LoRaCodingRate
				settings.Downlink.InvertPolarization = true
			}
			var f func(context.Context, int, ttnpb.TxSettings, scheduling.RTTs, ttnpb.TxSchedulePriority) (scheduling.Emission, error)
			switch request.Class {
			case ttnpb.CLASS_A:
				f = c.scheduler.ScheduleAt
				settings.Timestamp = uplinkTimestamp + uint32(rxDelay/time.Microsecond)
			case ttnpb.CLASS_B:
				f = c.scheduler.ScheduleAnytime
			case ttnpb.CLASS_C:
				if request.AbsoluteTime != nil {
					f = c.scheduler.ScheduleAt
					abs := *request.AbsoluteTime
					settings.Time = &abs
				} else {
					f = c.scheduler.ScheduleAnytime
				}
			default:
				panic(fmt.Sprintf("proto: unexpected class %v in oneof", request.Class))
			}
			em, err := f(c.ctx, len(msg.RawPayload), settings, c.rtts, request.Priority)
			if err != nil {
				logger.WithError(err).Debug("Failed to schedule downlink in Rx window")
				errRxDetails = append(errRxDetails, errRxWindowSchedule.WithCause(err).WithAttributes("window", i+1))
				continue
			}
			settings.Time = nil
			settings.Timestamp = uint32(time.Duration(em.Starts()) / time.Microsecond)
			msg.Settings = &ttnpb.DownlinkMessage_Scheduled{
				Scheduled: &settings,
			}
			errRxDetails = nil
			if now, ok := c.scheduler.Now(); ok {
				logger = logger.WithField("now", now)
				delay = time.Duration(em.Starts() - now)
			}
			logger.WithFields(log.Fields(
				"starts", em.Starts(),
				"duration", em.Duration(),
			)).Debug("Scheduled downlink")
			break
		}
		if errRxDetails != nil {
			return 0, errTxSchedule.WithDetails(errRxDetails...)
		}
	}
	select {
	case <-c.ctx.Done():
		return 0, c.ctx.Err()
	case c.downCh <- msg:
		atomic.AddUint64(&c.downlinks, 1)
		atomic.StoreInt64(&c.lastDownlinkTime, time.Now().UnixNano())
	default:
		return 0, errBufferFull
	}
	return delay, nil
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

// RTTStats returns the recorded round-trip time statistics.
func (c *Connection) RTTStats() (min, max, median time.Duration, count int) {
	return c.rtts.Stats()
}
