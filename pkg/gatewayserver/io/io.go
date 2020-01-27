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
	"go.thethings.network/lorawan-stack/pkg/config"
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
	// SupportsDownlinkClaim returns true if the frontend can itself claim downlinks.
	SupportsDownlinkClaim() bool
}

// Server represents the Gateway Server to gateway frontends.
type Server interface {
	// GetBaseConfig returns the component configuration.
	GetBaseConfig(ctx context.Context) config.ServiceBase
	// FillGatewayContext fills the given context and identifiers.
	// This method should only be used for request contexts.
	FillGatewayContext(ctx context.Context, ids ttnpb.GatewayIdentifiers) (context.Context, ttnpb.GatewayIdentifiers, error)
	// Connect connects a gateway by its identifiers to the Gateway Server, and returns a Connection for traffic and
	// control.
	Connect(ctx context.Context, frontend Frontend, ids ttnpb.GatewayIdentifiers) (*Connection, error)
	// GetFrequencyPlans gets the frequency plans by the gateway identifiers.
	GetFrequencyPlans(ctx context.Context, ids ttnpb.GatewayIdentifiers) (map[string]*frequencyplans.FrequencyPlan, error)
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

	frontend   Frontend
	gateway    *ttnpb.Gateway
	gatewayFPs map[string]*frequencyplans.FrequencyPlan
	bandID     string
	fps        *frequencyplans.Store
	scheduler  *scheduling.Scheduler
	rtts       *rtts

	upCh     chan *ttnpb.GatewayUplinkMessage
	downCh   chan *ttnpb.DownlinkMessage
	statusCh chan *ttnpb.GatewayStatus
	txAckCh  chan *ttnpb.TxAcknowledgment
}

var (
	errInconsistentFrequencyPlans = errors.DefineCorruption(
		"inconsistent_frequency_plans",
		"inconsistent frequency plans configuration",
	)
	errFrequencyPlansNotFromSameBand = errors.DefineInvalidArgument(
		"frequency_plans_not_from_same_band",
		"frequency plans must be from the same band",
	)
)

// NewConnection instantiates a new gateway connection.
func NewConnection(ctx context.Context, frontend Frontend, gateway *ttnpb.Gateway, fps *frequencyplans.Store, enforceDutyCycle bool, scheduleAnytimeDelay *time.Duration) (*Connection, error) {
	gatewayFPs := make(map[string]*frequencyplans.FrequencyPlan, len(gateway.FrequencyPlanIDs))
	fp0ID := gateway.FrequencyPlanID
	fp0, err := fps.GetByID(fp0ID)
	if err != nil {
		return nil, err
	}
	gatewayFPs[fp0ID] = fp0
	bandID := fp0.BandID

	if len(gateway.FrequencyPlanIDs) > 0 {
		if gateway.FrequencyPlanIDs[0] != fp0ID {
			return nil, errInconsistentFrequencyPlans
		}
		for i := 1; i < len(gateway.FrequencyPlanIDs); i++ {
			fpn, err := fps.GetByID(gateway.FrequencyPlanIDs[i])
			if err != nil {
				return nil, err
			}
			if fpn.BandID != fp0.BandID {
				return nil, errFrequencyPlansNotFromSameBand
			}
			gatewayFPs[gateway.FrequencyPlanIDs[i]] = fpn
		}
	}

	ctx, cancelCtx := errorcontext.New(ctx)
	scheduler, err := scheduling.NewScheduler(ctx, gatewayFPs, enforceDutyCycle, scheduleAnytimeDelay, nil)
	if err != nil {
		return nil, err
	}
	return &Connection{
		ctx:       ctx,
		cancelCtx: cancelCtx,

		frontend:    frontend,
		gateway:     gateway,
		gatewayFPs:  gatewayFPs,
		bandID:      bandID,
		fps:         fps,
		scheduler:   scheduler,
		rtts:        newRTTs(maxRTTs),
		upCh:        make(chan *ttnpb.GatewayUplinkMessage, bufferSize),
		downCh:      make(chan *ttnpb.DownlinkMessage, bufferSize),
		statusCh:    make(chan *ttnpb.GatewayStatus, bufferSize),
		txAckCh:     make(chan *ttnpb.TxAcknowledgment, bufferSize),
		connectTime: time.Now().UnixNano(),
	}, nil
}

// Context returns the connection context.
func (c *Connection) Context() context.Context { return c.ctx }

// Disconnect marks the connection as disconnected and cancels the context.
func (c *Connection) Disconnect(err error) {
	c.cancelCtx(err)
}

// Frontend returns the frontend using this connection.
func (c *Connection) Frontend() Frontend { return c.frontend }

// Gateway returns the gateway entity.
func (c *Connection) Gateway() *ttnpb.Gateway { return c.gateway }

var errBufferFull = errors.DefineInternal("buffer_full", "buffer is full")

// HandleUp updates the uplink stats and sends the message to the upstream channel.
func (c *Connection) HandleUp(up *ttnpb.UplinkMessage) error {
	if up.Settings.Time != nil {
		c.scheduler.SyncWithGatewayAbsolute(up.Settings.Timestamp, up.ReceivedAt, *up.Settings.Time)
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

		if c.gateway.LocationPublic && len(c.gateway.Antennas) > int(md.AntennaIndex) {
			location := c.gateway.Antennas[md.AntennaIndex].Location
			if location.Source != ttnpb.SOURCE_UNKNOWN {
				md.Location = &location
			}
		}
	}

	msg := &ttnpb.GatewayUplinkMessage{
		UplinkMessage: up,
		BandID:        c.bandID,
	}

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.upCh <- msg:
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
	errNoAbsoluteTime   = errors.DefineInvalidArgument("no_absolute_time", "no absolute time provided for class B downlink")
	errNoGPSSync        = errors.DefineFailedPrecondition("no_gps_sync", "gateway time is not GPS synchronized")
	errNoRxDelay        = errors.DefineInvalidArgument("no_rx_delay", "no Rx delay provided for class A downlink")
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

// SendDown sends the downlink message directly on the downlink channel.
func (c *Connection) SendDown(msg *ttnpb.DownlinkMessage) error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.downCh <- msg:
		atomic.AddUint64(&c.downlinks, 1)
		atomic.StoreInt64(&c.lastDownlinkTime, time.Now().UnixNano())
	default:
		return errBufferFull
	}
	return nil
}

var (
	errFrequencyPlanNotConfigured   = errors.DefineInvalidArgument("frequency_plan_not_configured", "frequency plan `{id}` is not configured for this gateway")
	errNoFrequencyPlanIDInTxRequest = errors.DefineInvalidArgument("no_frequency_plan_id_in_tx_request", "no frequency plan ID in tx request")
)

// ScheduleDown schedules and sends a downlink message by using the given path and updates the downlink stats.
// This method returns an error if the downlink message is not a Tx request.
func (c *Connection) ScheduleDown(path *ttnpb.DownlinkPath, msg *ttnpb.DownlinkMessage) (time.Duration, error) {
	if c.gateway.DownlinkPathConstraint == ttnpb.DOWNLINK_PATH_CONSTRAINT_NEVER {
		return 0, errNotAllowed
	}
	request := msg.GetRequest()
	if request == nil {
		return 0, errNotTxRequest
	}
	var delay time.Duration

	logger := log.FromContext(c.ctx).WithField("class", request.Class)
	logger.Debug("Attempt to schedule downlink on gateway")
	ids, uplinkTimestamp, err := getDownlinkPath(path, request.Class)
	if err != nil {
		return 0, err
	}

	var fp *frequencyplans.FrequencyPlan
	fpID := request.GetFrequencyPlanID()
	if fpID != "" {
		fp = c.gatewayFPs[fpID]
		if fp == nil {
			// The requested frequency plan is not configured for the gateway. Load the plan and enforce that it's in the same band.
			fp, err = c.fps.GetByID(fpID)
			if err != nil {
				return 0, errFrequencyPlanNotConfigured.WithCause(err).WithAttributes("id", request.FrequencyPlanID)
			}
			if fp.BandID != c.bandID {
				return 0, errFrequencyPlansNotFromSameBand
			}
		}
	} else {
		// Backwards compatibility. If there's no FrequencyPlanID in the TxRequest, then there must be only one Frequency Plan configured.
		if len(c.gatewayFPs) != 1 {
			return 0, errNoFrequencyPlanIDInTxRequest
		}
		for _, v := range c.gatewayFPs {
			fp = v
			break
		}
	}
	phy, err := band.GetByID(fp.BandID)
	if err != nil {
		return 0, err
	}
	var rxErrs []errors.ErrorDetails
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
		if rx.frequency == 0 {
			rxErrs = append(rxErrs, errRxEmpty)
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
		maxPHYLength := phy.DataRates[rx.dataRateIndex].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()) + 5
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
		if fp.MaxEIRP != nil {
			eirp = *fp.MaxEIRP
		}
		if sb, ok := fp.FindSubBand(rx.frequency); ok && sb.MaxEIRP != nil {
			eirp = *sb.MaxEIRP
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
			if request.Rx1Delay == ttnpb.RX_DELAY_0 {
				return 0, errNoRxDelay
			}
			settings.Timestamp = uplinkTimestamp + uint32((time.Duration(request.Rx1Delay)*time.Second+rx.delay)/time.Microsecond)
		case ttnpb.CLASS_B:
			if request.AbsoluteTime == nil {
				return 0, errNoAbsoluteTime
			}
			if !c.scheduler.IsGatewayTimeSynced() {
				rxErrs = append(rxErrs, errNoGPSSync)
				continue
			}
			f = c.scheduler.ScheduleAt
			settings.Time = request.AbsoluteTime
		case ttnpb.CLASS_C:
			if request.AbsoluteTime != nil {
				f = c.scheduler.ScheduleAt
				settings.Time = request.AbsoluteTime
			} else {
				f = c.scheduler.ScheduleAnytime
			}
		default:
			panic(fmt.Sprintf("proto: unexpected class %v in oneof", request.Class))
		}
		em, err := f(c.ctx, len(msg.RawPayload), settings, c.rtts, request.Priority)
		if err != nil {
			logger.WithError(err).Debug("Failed to schedule downlink in Rx window")
			rxErrs = append(rxErrs, errRxWindowSchedule.WithCause(err).WithAttributes("window", i+1))
			continue
		}
		if settings.Time == nil || !c.scheduler.IsGatewayTimeSynced() {
			settings.Time = nil
			settings.Timestamp = uint32(time.Duration(em.Starts()) / time.Microsecond)
		} else {
			settings.Timestamp = 0
		}
		msg.Settings = &ttnpb.DownlinkMessage_Scheduled{
			Scheduled: &settings,
		}
		rxErrs = nil
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
	if len(rxErrs) > 0 {
		protoErrs := make([]*ttnpb.ErrorDetails, 0, len(rxErrs))
		for _, rxErr := range rxErrs {
			protoErrs = append(protoErrs, ttnpb.ErrorDetailsToProto(rxErr))
		}
		return 0, errTxSchedule.WithDetails(&ttnpb.ScheduleDownlinkErrorDetails{
			PathErrors: protoErrs,
		})
	}
	err = c.SendDown(msg)
	if err != nil {
		return 0, err
	}
	return delay, nil
}

// Status returns the status channel.
func (c *Connection) Status() <-chan *ttnpb.GatewayStatus {
	return c.statusCh
}

// Up returns the upstream channel.
func (c *Connection) Up() <-chan *ttnpb.GatewayUplinkMessage {
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

// FrequencyPlans returns the frequency plans for the gateway.
func (c *Connection) FrequencyPlans() map[string]*frequencyplans.FrequencyPlan { return c.gatewayFPs }

// BandID returns the common band ID for the frequency plans in this connection.
// TODO: Handle mixed bands (https://github.com/TheThingsNetwork/lorawan-stack/issues/1394)
func (c *Connection) BandID() string { return c.bandID }

// SyncWithGatewayConcentrator synchronizes the clock with the given concentrator timestamp, the server time and the
// relative gateway time that corresponds to the given timestamp.
func (c *Connection) SyncWithGatewayConcentrator(timestamp uint32, server time.Time, concentrator scheduling.ConcentratorTime) {
	c.scheduler.SyncWithGatewayConcentrator(timestamp, server, concentrator)
}

// TimeFromTimestampTime returns the concentrator time by the given timestamp.
// This method returns false if the clock is not synced with the server.
func (c *Connection) TimeFromTimestampTime(timestamp uint32) (scheduling.ConcentratorTime, bool) {
	return c.scheduler.TimeFromTimestampTime(timestamp)
}
