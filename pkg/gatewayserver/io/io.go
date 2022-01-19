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
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/mohae/deepcopy"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

const (
	bufferSize = 1 << 4

	maxRTTs = 20
	rttTTL  = 30 * time.Minute
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
	// FromRequestContext decouples the lifetime of the provided context from the values found in the context.
	FromRequestContext(ctx context.Context) context.Context
	// RateLimiter returns the rate limiter instance.
	RateLimiter() ratelimit.Interface
	// ValidateGatewayID validates the ID of the gateway.
	ValidateGatewayID(ctx context.Context, ids ttnpb.GatewayIdentifiers) error
	task.Starter
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

	frontend         Frontend
	gateway          *ttnpb.Gateway
	gatewayPrimaryFP *frequencyplans.FrequencyPlan
	gatewayFPs       map[string]*frequencyplans.FrequencyPlan
	bandID           string
	fps              *frequencyplans.Store
	scheduler        *scheduling.Scheduler
	rtts             *rtts

	upCh     chan *ttnpb.GatewayUplinkMessage
	downCh   chan *ttnpb.DownlinkMessage
	statusCh chan *ttnpb.GatewayStatus
	txAckCh  chan *ttnpb.TxAcknowledgment

	statsChangedCh chan struct{}
	locCh          chan struct{}

	versionInfoCh chan struct{}

	lastUplink            *uplinkMessage
	lastRepeatUpEventTime time.Time
}

type uplinkMessage struct {
	payload   []byte
	frequency uint64
	antennas  []uint32
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
	gatewayFPs := make(map[string]*frequencyplans.FrequencyPlan, len(gateway.FrequencyPlanIds))
	fp0ID := gateway.FrequencyPlanId
	fp0, err := fps.GetByID(fp0ID)
	if err != nil {
		return nil, err
	}
	gatewayFPs[fp0ID] = fp0
	bandID := fp0.BandID

	if len(gateway.FrequencyPlanIds) > 0 {
		if gateway.FrequencyPlanIds[0] != fp0ID {
			return nil, errInconsistentFrequencyPlans.New()
		}
		for i := 1; i < len(gateway.FrequencyPlanIds); i++ {
			fpn, err := fps.GetByID(gateway.FrequencyPlanIds[i])
			if err != nil {
				return nil, err
			}
			if fpn.BandID != fp0.BandID {
				return nil, errFrequencyPlansNotFromSameBand.New()
			}
			gatewayFPs[gateway.FrequencyPlanIds[i]] = fpn
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

		frontend:         frontend,
		gateway:          gateway,
		gatewayPrimaryFP: fp0,
		gatewayFPs:       gatewayFPs,
		bandID:           bandID,
		fps:              fps,
		scheduler:        scheduler,
		rtts:             newRTTs(maxRTTs, rttTTL),
		upCh:             make(chan *ttnpb.GatewayUplinkMessage, bufferSize),
		downCh:           make(chan *ttnpb.DownlinkMessage, bufferSize),
		statusCh:         make(chan *ttnpb.GatewayStatus, bufferSize),
		txAckCh:          make(chan *ttnpb.TxAcknowledgment, bufferSize),
		locCh:            make(chan struct{}, 1),
		versionInfoCh:    make(chan struct{}, 1),
		connectTime:      time.Now().UnixNano(),

		statsChangedCh: make(chan struct{}, 1),
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

// Interval between emitting consecutive gs.up.repeat events for the same gateway connection.
const consecutiveRepeatUpEventsInterval = time.Minute

// FrontendClockSynchronization contains the clock synchronization
// timestamps provided by a frontend for manual synchronization.
type FrontendClockSynchronization struct {
	Timestamp        uint32
	ServerTime       time.Time
	GatewayTime      *time.Time
	ConcentratorTime scheduling.ConcentratorTime
}

// HandleUp updates the uplink stats and sends the message to the upstream channel.
func (c *Connection) HandleUp(up *ttnpb.UplinkMessage, frontendSync *FrontendClockSynchronization) (err error) {
	defer func() {
		if err != nil {
			registerDropMessage(c.ctx, c.gateway, "uplink", err)
		}
	}()
	if err := up.ValidateFields(); err != nil {
		return err
	}
	if c.discardRepeatedUplink(up) {
		return nil
	}

	var ct scheduling.ConcentratorTime
	switch {
	case frontendSync != nil:
		ct = c.scheduler.SyncWithGatewayConcentrator(frontendSync.Timestamp, frontendSync.ServerTime, frontendSync.GatewayTime, frontendSync.ConcentratorTime)
		log.FromContext(c.ctx).WithFields(log.Fields(
			"timestamp", frontendSync.Timestamp,
			"concentrator_time", frontendSync.ConcentratorTime,
			"server_time", frontendSync.ServerTime,
			"gateway_time", frontendSync.GatewayTime,
		)).Debug("Gateway clocks have been synchronized by the frontend")
	case up.Settings.Time != nil:
		ct = c.scheduler.SyncWithGatewayAbsolute(up.Settings.Timestamp, *ttnpb.StdTime(up.ReceivedAt), *ttnpb.StdTime(up.Settings.Time))
		log.FromContext(c.ctx).WithFields(log.Fields(
			"timestamp", up.Settings.Timestamp,
			"concentrator_time", ct,
			"server_time", up.ReceivedAt,
			"gateway_time", *up.Settings.Time,
		)).Debug("Synchronized server and gateway absolute time")
	case up.Settings.Time == nil:
		ct = c.scheduler.Sync(up.Settings.Timestamp, *ttnpb.StdTime(up.ReceivedAt))
		log.FromContext(c.ctx).WithFields(log.Fields(
			"timestamp", up.Settings.Timestamp,
			"concentrator_time", ct,
			"server_time", up.ReceivedAt,
		)).Debug("Synchronized server absolute time only")
	default:
		panic("unreachable")
	}

	for _, md := range up.RxMetadata {
		if md.AntennaIndex != 0 {
			// TODO: Support downlink path to multiple antennas (https://github.com/TheThingsNetwork/lorawan-stack/issues/48)
			md.DownlinkPathConstraint = ttnpb.DownlinkPathConstraint_DOWNLINK_PATH_CONSTRAINT_NEVER
			continue
		}
		buf, err := UplinkToken(&ttnpb.GatewayAntennaIdentifiers{
			GatewayIds:   c.gateway.GetIds(),
			AntennaIndex: md.AntennaIndex,
		}, md.Timestamp, ct, *ttnpb.StdTime(up.ReceivedAt), ttnpb.StdTime(up.Settings.Time))
		if err != nil {
			return err
		}
		md.UplinkToken = buf
		md.DownlinkPathConstraint = c.gateway.DownlinkPathConstraint

		if c.gateway.LocationPublic && len(c.gateway.Antennas) > int(md.AntennaIndex) {
			location := c.gateway.Antennas[md.AntennaIndex].Location
			if location != nil && location.Source != ttnpb.SOURCE_UNKNOWN {
				md.Location = location
			}
		} else if !c.gateway.LocationPublic {
			md.Location = nil
		}
	}

	msg := &ttnpb.GatewayUplinkMessage{
		Message: up,
		BandId:  c.bandID,
	}

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.upCh <- msg:
		atomic.AddUint64(&c.uplinks, 1)
		atomic.StoreInt64(&c.lastUplinkTime, ttnpb.StdTime(up.ReceivedAt).UnixNano())
		c.notifyStatsChanged()
	default:
		return errBufferFull.New()
	}
	return nil
}

// HandleStatus updates the status stats and sends the status to the status channel.
func (c *Connection) HandleStatus(status *ttnpb.GatewayStatus) (err error) {
	defer func() {
		if err != nil {
			registerDropMessage(c.ctx, c.gateway, "status", err)
		}
	}()

	if err := status.ValidateFields(); err != nil {
		return err
	}
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.statusCh <- status:
		c.lastStatus.Store(deepcopy.Copy(status))
		atomic.StoreInt64(&c.lastStatusTime, time.Now().UnixNano())
		c.notifyStatsChanged()

		if len(status.AntennaLocations) > 0 && c.gateway.UpdateLocationFromStatus {
			select {
			case c.locCh <- struct{}{}:
			default:
			}
		}

		// The channel is only written to once, after which there is no longer a recipient.
		// For all subsequent status messages, the default branch is chosen.
		select {
		case c.versionInfoCh <- struct{}{}:
		default:
		}

	default:
		return errBufferFull.New()
	}
	return nil
}

// HandleTxAck sends the acknowledgment to the status channel.
func (c *Connection) HandleTxAck(ack *ttnpb.TxAcknowledgment) (err error) {
	defer func() {
		if err != nil {
			registerDropMessage(c.ctx, c.gateway, "txack", err)
		}
	}()
	if err := ack.ValidateFields(); err != nil {
		return err
	}
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.txAckCh <- ack:
		c.notifyStatsChanged()
	default:
		return errBufferFull.New()
	}
	return nil
}

// RecordRTT records the given round-trip time.
func (c *Connection) RecordRTT(d time.Duration, t time.Time) {
	c.rtts.Record(d, t)
	c.notifyStatsChanged()
}

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
	errDataRateRxWindow = errors.DefineInvalidArgument("data_rate_rx_window", "invalid data rate in Rx window `{window}`")
	errTooLong          = errors.DefineInvalidArgument("too_long", "the payload length `{payload_length}` exceeds maximum `{maximum_length}` at data rate `{data_rate}`")
	errTxSchedule       = errors.DefineAborted("tx_schedule", "failed to schedule")
)

// getDownlinkPath returns the downlink path.
// If the path contains an uplink token, the gateway antenna identifiers are taken from the uplink token, and the uplink token is returned.
// If the path is fixed, the gateway antenna identifiers are taken from the fixed path.
// Class A downlink requires the path to provide an uplink token, while class B and C downlink may use a fixed downlink path.
func getDownlinkPath(path *ttnpb.DownlinkPath, class ttnpb.Class) (ttnpb.GatewayAntennaIdentifiers, *ttnpb.UplinkToken, error) {
	if buf := path.GetUplinkToken(); len(buf) == 0 {
		if class == ttnpb.CLASS_A {
			return ttnpb.GatewayAntennaIdentifiers{}, nil, errNoUplinkToken.New()
		}
	} else {
		token, err := ParseUplinkToken(buf)
		if err != nil {
			return ttnpb.GatewayAntennaIdentifiers{}, nil, err
		}
		return *token.Ids, token, err
	}
	fixed := path.GetFixed()
	if fixed == nil {
		return ttnpb.GatewayAntennaIdentifiers{}, nil, errDownlinkPath.New()
	}
	return *fixed, nil, nil
}

// SendDown sends the downlink message directly on the downlink channel.
func (c *Connection) SendDown(msg *ttnpb.DownlinkMessage) error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.downCh <- msg:
		atomic.AddUint64(&c.downlinks, 1)
		atomic.StoreInt64(&c.lastDownlinkTime, time.Now().UnixNano())

		c.notifyStatsChanged()
	default:
		return errBufferFull.New()
	}
	return nil
}

var (
	errFrequencyPlanNotConfigured   = errors.DefineInvalidArgument("frequency_plan_not_configured", "frequency plan `{id}` is not configured for this gateway")
	errNoFrequencyPlanIDInTxRequest = errors.DefineInvalidArgument("no_frequency_plan_id_in_tx_request", "no frequency plan ID in tx request")
)

// ScheduleDown schedules and sends a downlink message by using the given path and updates the downlink stats.
// This method returns an error if the downlink message is not a Tx request.
func (c *Connection) ScheduleDown(path *ttnpb.DownlinkPath, msg *ttnpb.DownlinkMessage) (rx1, rx2 bool, delay time.Duration, err error) {
	if c.gateway.DownlinkPathConstraint == ttnpb.DownlinkPathConstraint_DOWNLINK_PATH_CONSTRAINT_NEVER {
		return false, false, 0, errNotAllowed.New()
	}
	request := msg.GetRequest()
	if request == nil {
		return false, false, 0, errNotTxRequest.New()
	}

	logger := log.FromContext(c.ctx).WithField("class", request.Class)
	logger.Debug("Attempt to schedule downlink on gateway")
	ids, uplinkToken, err := getDownlinkPath(path, request.Class)
	if err != nil {
		return false, false, 0, err
	}

	var fp *frequencyplans.FrequencyPlan
	fpID := request.GetFrequencyPlanId()
	if fpID != "" {
		fp = c.gatewayFPs[fpID]
		if fp == nil {
			// The requested frequency plan is not configured for the gateway. Load the plan and enforce that it's in the same band.
			fp, err = c.fps.GetByID(fpID)
			if err != nil {
				return false, false, 0, errFrequencyPlanNotConfigured.WithCause(err).WithAttributes("id", request.FrequencyPlanId)
			}
			if fp.BandID != c.bandID {
				return false, false, 0, errFrequencyPlansNotFromSameBand.New()
			}
		}
	} else {
		// Backwards compatibility. If there's no FrequencyPlanID in the TxRequest, then there must be only one Frequency
		// Plan configured.
		// When implementing https://github.com/TheThingsNetwork/lorawan-stack/issues/1394, having multiple frequency plans
		// or even bands should not error. Instead, the minimum MaxEIRP in any frequency plan for the given frequency should
		// be used below to make sure that regional regulations are respected.
		if len(c.gatewayFPs) != 1 {
			return false, false, 0, errNoFrequencyPlanIDInTxRequest.New()
		}
		for _, v := range c.gatewayFPs {
			fp = v
			break
		}
	}

	// Gateway Server does not take the LoRaWAN Regional Parameters version into account as it is a transparent forwarder
	// between the Network Server and the end device. However, Gateway Server does enforce spectrum regulations that are
	// defined in Regional Parameters. These include maximum EIRP and maximum payload length. These are taken from the
	// last known LoRaWAN Regional Parameters version.
	phy, err := band.GetLatest(fp.BandID)
	if err != nil {
		return false, false, 0, err
	}

	var rxErrs []errors.ErrorDetails
	for i, rx := range []struct {
		dataRate  *ttnpb.DataRate
		frequency uint64
		delay     time.Duration
	}{
		{
			dataRate:  request.Rx1DataRate,
			frequency: request.Rx1Frequency,
			delay:     0,
		},
		{
			dataRate:  request.Rx2DataRate,
			frequency: request.Rx2Frequency,
			delay:     time.Second,
		},
	} {
		if rx.frequency == 0 {
			rxErrs = append(rxErrs, errRxEmpty.New())
			continue
		}
		if rx.dataRate == nil {
			rxErrs = append(rxErrs, errDataRateRxWindow.WithAttributes("window", i+1))
			continue
		}
		_, bandDR, ok := phy.FindDownlinkDataRate(rx.dataRate)
		if !ok {
			rxErrs = append(rxErrs, errDataRateRxWindow.WithAttributes("window", i+1))
			continue
		}

		logger := logger.WithFields(log.Fields(
			"rx_window", i+1,
			"frequency", rx.frequency,
			"data_rate", rx.dataRate,
		))
		logger.Debug("Attempt to schedule downlink in receive window")
		// The maximum payload size is MACPayload only; for PHYPayload take MHDR (1 byte) and MIC (4 bytes) into account.
		maxPHYLength := bandDR.MaxMACPayloadSize(fp.DwellTime.GetDownlinks()) + 5
		if len(msg.RawPayload) > int(maxPHYLength) {
			return false, false, 0, errTooLong.WithAttributes(
				"payload_length", len(msg.RawPayload),
				"maximum_length", maxPHYLength,
				"data_rate", rx.dataRate,
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
			DataRate:  rx.dataRate,
			Frequency: rx.frequency,
			Downlink: &ttnpb.TxSettings_Downlink{
				TxPower:      eirp,
				AntennaIndex: ids.AntennaIndex,
			},
		}
		if int(ids.AntennaIndex) < len(c.gateway.Antennas) {
			settings.Downlink.TxPower -= c.gateway.Antennas[ids.AntennaIndex].Gain
		}
		switch mod := rx.dataRate.Modulation.(type) {
		case *ttnpb.DataRate_Lora:
			// TODO: Set coding rate from data rate (https://github.com/TheThingsNetwork/lorawan-stack/issues/4466).
			settings.CodingRate = phy.LoRaCodingRate
			settings.Downlink.InvertPolarization = true
		case *ttnpb.DataRate_Lrfhss:
			settings.CodingRate = mod.Lrfhss.CodingRate
		}
		var f func(context.Context, scheduling.Options) (scheduling.Emission, error)
		switch request.Class {
		case ttnpb.CLASS_A:
			f = c.scheduler.ScheduleAt
			if request.Rx1Delay == ttnpb.RX_DELAY_0 {
				return false, false, 0, errNoRxDelay.New()
			}
			settings.Timestamp = uplinkToken.Timestamp + uint32((time.Duration(request.Rx1Delay)*time.Second+rx.delay)/time.Microsecond)
		case ttnpb.CLASS_B:
			if request.AbsoluteTime == nil {
				return false, false, 0, errNoAbsoluteTime.New()
			}
			if !c.scheduler.IsGatewayTimeSynced() {
				rxErrs = append(rxErrs, errNoGPSSync.New())
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
		em, err := f(c.ctx, scheduling.Options{
			PayloadSize: len(msg.RawPayload),
			TxSettings:  settings,
			RTTs:        c.rtts,
			Priority:    request.Priority,
			UplinkToken: uplinkToken, // uplinkToken is always present with class A downlink, but may be nil otherwise.
		})
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
		rx1 = i == 0
		rx2 = i == 1
		rxErrs = nil
		if now, ok := c.scheduler.Now(); ok {
			logger = logger.WithField("now", now)
			delay = time.Duration(em.Starts() - now)
		}
		logger.WithFields(log.Fields(
			"rx_window", i+1,
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
		return false, false, 0, errTxSchedule.WithDetails(&ttnpb.ScheduleDownlinkErrorDetails{
			PathErrors: protoErrs,
		})
	}
	err = c.SendDown(msg)
	if err != nil {
		return false, false, 0, err
	}
	return
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

// StatsChanged returns the stats changed channel.
func (c *Connection) StatsChanged() <-chan struct{} {
	return c.statsChangedCh
}

// LocationChanged returns the location updates channel.
func (c *Connection) LocationChanged() <-chan struct{} {
	return c.locCh
}

// VersionInfoChanged returns the version info updates channel.
func (c *Connection) VersionInfoChanged() <-chan struct{} {
	return c.versionInfoCh
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
func (c *Connection) RTTStats(percentile int, t time.Time) (min, max, median, np time.Duration, count int) {
	return c.rtts.Stats(percentile, t)
}

// Stats collects and returns the gateway connection statistics and the field mask paths.
func (c *Connection) Stats() (*ttnpb.GatewayConnectionStats, []string) {
	ct := c.ConnectTime()
	stats := &ttnpb.GatewayConnectionStats{
		ConnectedAt: ttnpb.ProtoTimePtr(ct),
		Protocol:    c.Frontend().Protocol(),
	}
	paths := make([]string, 0, len(ttnpb.GatewayConnectionStatsFieldPathsTopLevel))
	paths = append(paths, "connected_at", "disconnected_at", "protocol")

	if s, t, ok := c.StatusStats(); ok {
		stats.LastStatusReceivedAt = ttnpb.ProtoTimePtr(t)
		stats.LastStatus = s
		paths = append(paths, "last_status_received_at", "last_status")
	}
	if count, t, ok := c.UpStats(); ok {
		stats.LastUplinkReceivedAt = ttnpb.ProtoTimePtr(t)
		stats.UplinkCount = count
		paths = append(paths, "last_uplink_received_at", "uplink_count")
	}
	if count, t, ok := c.DownStats(); ok {
		stats.LastDownlinkReceivedAt = ttnpb.ProtoTimePtr(t)
		stats.DownlinkCount = count
		paths = append(paths, "last_downlink_received_at", "downlink_count")
		if c.scheduler != nil {
			// Usage statistics are only available for downlink.
			stats.SubBands = c.scheduler.SubBandStats()
			paths = append(paths, "sub_bands")
		}
	}
	if min, max, median, _, count := c.RTTStats(100, time.Now()); count > 0 {
		stats.RoundTripTimes = &ttnpb.GatewayConnectionStats_RoundTripTimes{
			Min:    ttnpb.ProtoDurationPtr(min),
			Max:    ttnpb.ProtoDurationPtr(max),
			Median: ttnpb.ProtoDurationPtr(median),
			Count:  uint32(count),
		}
		paths = append(paths, "round_trip_times")
	}
	return stats, paths
}

// FrequencyPlans returns the frequency plans for the gateway.
func (c *Connection) FrequencyPlans() map[string]*frequencyplans.FrequencyPlan { return c.gatewayFPs }

// PrimaryFrequencyPlan returns the primary frequency plan of the gateway.
func (c *Connection) PrimaryFrequencyPlan() *frequencyplans.FrequencyPlan { return c.gatewayPrimaryFP }

// BandID returns the common band ID for the frequency plans in this connection.
// TODO: Handle mixed bands (https://github.com/TheThingsNetwork/lorawan-stack/issues/1394)
func (c *Connection) BandID() string { return c.bandID }

// SyncWithGatewayConcentrator synchronizes the clock with the given concentrator timestamp, the server time and the
// relative gateway time that corresponds to the given timestamp.
func (c *Connection) SyncWithGatewayConcentrator(timestamp uint32, server time.Time, gateway *time.Time, concentrator scheduling.ConcentratorTime) scheduling.ConcentratorTime {
	return c.scheduler.SyncWithGatewayConcentrator(timestamp, server, gateway, concentrator)
}

// TimeFromTimestampTime returns the concentrator time by the given timestamp.
// This method returns false if the clock is not synced with the server.
func (c *Connection) TimeFromTimestampTime(timestamp uint32) (scheduling.ConcentratorTime, bool) {
	return c.scheduler.TimeFromTimestampTime(timestamp)
}

// TimeFromServerTime returns the concentrator time by the given server time.
// This method returns false if the clock is not synced with the server.
func (c *Connection) TimeFromServerTime(t time.Time) (scheduling.ConcentratorTime, bool) {
	return c.scheduler.TimeFromServerTime(t)
}

func (c *Connection) notifyStatsChanged() {
	select {
	case c.statsChangedCh <- struct{}{}:
	default:
	}
}

func uplinkMessageFromProto(pb *ttnpb.UplinkMessage) *uplinkMessage {
	up := &uplinkMessage{
		payload:   pb.GetRawPayload(),
		frequency: pb.GetSettings().Frequency,
		antennas:  make([]uint32, 0, len(pb.GetRxMetadata())),
	}
	for _, md := range pb.GetRxMetadata() {
		up.antennas = append(up.antennas, md.GetAntennaIndex())
	}
	return up
}

func isRepeatedUplink(this *uplinkMessage, that *uplinkMessage) bool {
	if this == nil || that == nil || this.frequency != that.frequency || len(this.antennas) != len(that.antennas) || !bytes.Equal(this.payload, that.payload) {
		return false
	}
	for idx, antenna := range this.antennas {
		if that.antennas[idx] != antenna {
			return false
		}
	}
	return true
}

// discardRepeatedUplink will discard repeated uplinks from faulty gateway
// implementations. It returns true if the uplink message is the same as the
// last uplink message that was received by the connection.
//
// discardRepeatedUplink is not goroutine safe.
func (c *Connection) discardRepeatedUplink(up *ttnpb.UplinkMessage) bool {
	uplink := uplinkMessageFromProto(up)
	shouldDiscard := isRepeatedUplink(c.lastUplink, uplink)
	c.lastUplink = uplink
	if shouldDiscard {
		shouldEmitEvent := false
		if time.Since(c.lastRepeatUpEventTime) >= consecutiveRepeatUpEventsInterval {
			log.FromContext(c.ctx).Debug("Dropped repeated gateway uplink")
			shouldEmitEvent = true
			c.lastRepeatUpEventTime = time.Now()
		}
		registerRepeatUp(c.ctx, shouldEmitEvent, c.gateway, c.frontend.Protocol())
	}
	return shouldDiscard
}

type rssiAndIndex struct {
	rssi  float32
	index int
}

// UniqueUplinkMessagesByRSSI returns the given list of gateway uplink messages after discarding
// duplicates by RSSI. Two gateway uplink messages are considered duplicates if the RawPayload
// is identical, and the RSSI values differ. In these cases, only the gateway uplink message
// with the highest RSSI will be included in the result.
//
// UniqueUplinkMessagesByRSSI will allocate a new list of uplink messages, but will not copy the uplink
// messages themselves.
func UniqueUplinkMessagesByRSSI(uplinks []*ttnpb.UplinkMessage) []*ttnpb.UplinkMessage {
	if len(uplinks) < 2 {
		return uplinks
	}

	maxRSSI := make(map[string]rssiAndIndex, len(uplinks))
	deduplicated := make([]*ttnpb.UplinkMessage, 0, len(uplinks))
	for _, uplink := range uplinks {
		md := uplink.GetRxMetadata()
		if len(md) == 0 {
			deduplicated = append(deduplicated, uplink)
			continue
		}
		key := base64.StdEncoding.EncodeToString(uplink.GetRawPayload())
		if s, ok := maxRSSI[key]; ok && s.rssi < md[0].Rssi {
			deduplicated[s.index] = uplink
			maxRSSI[key] = rssiAndIndex{md[0].Rssi, s.index}
		} else if !ok {
			deduplicated = append(deduplicated, uplink)
			maxRSSI[key] = rssiAndIndex{md[0].Rssi, len(deduplicated) - 1}
		}
	}
	return deduplicated
}
