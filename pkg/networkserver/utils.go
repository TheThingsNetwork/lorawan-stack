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

package networkserver

import (
	"context"
	"fmt"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// nsScheduleWindow returns minimum time.Duration between downlink being added to the queue and it being sent to GS for transmission.
func nsScheduleWindow() time.Duration {
	// TODO: Observe this value at runtime https://github.com/TheThingsNetwork/lorawan-stack/issues/1552.
	return 200 * time.Millisecond
}

var (
	timeNow   func() time.Time                       = time.Now
	timeAfter func(d time.Duration) <-chan time.Time = time.After
)

func timeUntil(t time.Time) time.Duration {
	return t.Sub(timeNow())
}

func searchUplinkChannel(freq uint64, macState *ttnpb.MACState) (uint8, error) {
	for i, ch := range macState.CurrentParameters.Channels {
		if ch.UplinkFrequency == freq {
			return uint8(i), nil
		}
	}
	return 0, errUplinkChannelNotFound.WithAttributes("frequency", freq)
}

type downlinkSlot interface {
	From() time.Time
	IsContinuous() bool
}

type classADownlinkSlot struct {
	Uplink  *ttnpb.UplinkMessage
	RxDelay time.Duration
}

func (s classADownlinkSlot) From() time.Time {
	return time.Time{}
}

func (s classADownlinkSlot) RX1() time.Time {
	return s.Uplink.ReceivedAt.Add(s.RxDelay)
}

func (s classADownlinkSlot) RX2() time.Time {
	return s.RX1().Add(time.Second)
}

func (s classADownlinkSlot) IsContinuous() bool {
	return false
}

type networkInitiatedDownlinkSlot struct {
	Time              time.Time
	Class             ttnpb.Class
	IsApplicationTime bool
}

func (s networkInitiatedDownlinkSlot) From() time.Time {
	return s.Time
}

func (s networkInitiatedDownlinkSlot) IsContinuous() bool {
	return !s.IsApplicationTime && s.Class == ttnpb.CLASS_C
}

// lastClassADataDownlinkSlot returns the latest class A downlink slot in current session
// if such exists and true, otherwise it returns nil and false.
func lastClassADataDownlinkSlot(dev *ttnpb.EndDevice, phy *band.Band) (*classADownlinkSlot, bool) {
	if dev.GetMACState() == nil || len(dev.MACState.RecentUplinks) == 0 || dev.Multicast {
		return nil, false
	}
	var rxDelay time.Duration
	up := LastUplink(dev.MACState.RecentUplinks...)
	switch up.Payload.MHDR.MType {
	case ttnpb.MType_CONFIRMED_UP, ttnpb.MType_UNCONFIRMED_UP:
		rxDelay = dev.MACState.CurrentParameters.Rx1Delay.Duration()

	case ttnpb.MType_REJOIN_REQUEST:
		rxDelay = phy.JoinAcceptDelay1

	default:
		return nil, false
	}
	return &classADownlinkSlot{
		RxDelay: rxDelay,
		Uplink:  up,
	}, true
}

// nextUnconfirmedNetworkInitiatedDownlinkAt returns the earliest possible time instant when next unconfirmed
// network-initiated data downlink can be transmitted to the device given the data known by Network Server and true,
// if such time instant exists, otherwise it returns time.Time{} and false.
func nextUnconfirmedNetworkInitiatedDownlinkAt(ctx context.Context, dev *ttnpb.EndDevice, phy *band.Band) (time.Time, bool) {
	switch {
	case dev.GetMACState() == nil:
		log.FromContext(ctx).Warn("Insufficient data to compute next network-initiated unconfirmed downlink slot")
		return time.Time{}, false

	case dev.MACState.DeviceClass == ttnpb.CLASS_A:
		return time.Time{}, false

	case dev.MACState.LastDownlinkAt == nil:
		classA, hasClassA := lastClassADataDownlinkSlot(dev, phy)
		if !hasClassA {
			return time.Time{}, true
		}
		return classA.RX2(), true

	case dev.MACState.LastNetworkInitiatedDownlinkAt == nil:
		classA, hasClassA := lastClassADataDownlinkSlot(dev, phy)
		if !hasClassA {
			return *dev.MACState.LastDownlinkAt, true
		}
		return latestTime(classA.RX2(), *dev.MACState.LastDownlinkAt), true
	}
	classA, hasClassA := lastClassADataDownlinkSlot(dev, phy)
	if !hasClassA {
		return dev.MACState.LastNetworkInitiatedDownlinkAt.Add(networkInitiatedDownlinkInterval), true
	}
	if classA.Uplink.ReceivedAt.After(*dev.MACState.LastNetworkInitiatedDownlinkAt) {
		return classA.RX2(), true
	}
	return latestTime(classA.RX2(), dev.MACState.LastNetworkInitiatedDownlinkAt.Add(networkInitiatedDownlinkInterval)), true
}

// nextConfirmedNetworkInitiatedDownlinkAt returns the earliest possible time instant when a confirmed
// network-initiated data downlink can be transmitted to the device given the data known by Network Server and true,
// if such time instant exists, otherwise it returns time.Time{} and false.
func nextConfirmedNetworkInitiatedDownlinkAt(ctx context.Context, dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) (time.Time, bool) {
	if dev.GetMACState() == nil {
		log.FromContext(ctx).Warn("Insufficient data to compute next network-initiated confirmed downlink slot")
		return time.Time{}, false
	}
	if dev.Multicast {
		return time.Time{}, false
	}

	unconfAt, ok := nextUnconfirmedNetworkInitiatedDownlinkAt(ctx, dev, phy)
	switch {
	case !ok:
		return time.Time{}, false

	case dev.MACState.LastConfirmedDownlinkAt == nil,
		len(dev.MACState.RecentUplinks) > 0 && LastUplink(dev.MACState.RecentUplinks...).ReceivedAt.After(*dev.MACState.LastConfirmedDownlinkAt):
		return unconfAt, true
	}

	var timeout time.Duration
	switch dev.MACState.DeviceClass {
	case ttnpb.CLASS_B:
		timeout = mac.DeviceClassBTimeout(dev, defaults)

	case ttnpb.CLASS_C:
		timeout = mac.DeviceClassCTimeout(dev, defaults)
	default:
		panic(fmt.Errorf("unmatched class: %v", dev.MACState.DeviceClass))
	}
	if t := dev.MACState.LastConfirmedDownlinkAt.Add(timeout); t.After(unconfAt) {
		return t, true
	}
	return unconfAt, true
}

func latestTime(ts ...time.Time) time.Time {
	if len(ts) == 0 {
		return time.Time{}
	}
	max := ts[0]
	for _, t := range ts {
		if t.After(max) {
			max = t
		}
	}
	return max
}

func deviceHasPathForDownlink(ctx context.Context, dev *ttnpb.EndDevice, down *ttnpb.ApplicationDownlink) bool {
	if dev.GetMulticast() || dev.GetMACState() == nil {
		return len(down.GetClassBC().GetGateways()) > 0
	}
	switch dev.MACState.DeviceClass {
	case ttnpb.CLASS_A:
		return down.GetClassBC() == nil && len(downlinkPathsFromRecentUplinks(dev.GetMACState().GetRecentUplinks()...)) > 0
	case ttnpb.CLASS_B, ttnpb.CLASS_C:
		return len(downlinkPathsFromRecentUplinks(dev.GetMACState().GetRecentUplinks()...)) > 0 || len(down.GetClassBC().GetGateways()) > 0
	default:
		panic(fmt.Errorf("unmatched class: %v", dev.MACState.DeviceClass))
	}
}

// nextDataDownlinkSlot returns the next downlinkSlot before or at earliestAt when next data downlink can be transmitted to the device
// given the data known by Network Server and true, if such downlinkSlot and downlink exist, otherwise it returns nil and false.
func nextDataDownlinkSlot(ctx context.Context, dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings, earliestAt time.Time) (downlinkSlot, bool) {
	if dev.GetMACState() == nil {
		return nil, false
	}
	earliestAt = latestTime(earliestAt, timeNow())
	if dev.MACState.LastDownlinkAt != nil {
		earliestAt = latestTime(earliestAt, *dev.MACState.LastDownlinkAt)
	}
	logger := log.FromContext(ctx).WithField("earliest_at", earliestAt)

	var needsAck bool
	classA, hasClassA := lastClassADataDownlinkSlot(dev, phy)
	if hasClassA {
		switch classA.Uplink.Payload.MHDR.MType {
		case ttnpb.MType_UNCONFIRMED_UP:
			if classA.Uplink.Payload.GetMACPayload().FCtrl.ADRAckReq {
				logger.Debug("Acknowledgement required for ADRAckReq")
				needsAck = dev.MACState.LastDownlinkAt == nil || dev.MACState.LastDownlinkAt.Before(classA.Uplink.ReceivedAt)
			}
		case ttnpb.MType_CONFIRMED_UP:
			logger.Debug("Acknowledgement required for confirmed uplink")
			needsAck = dev.MACState.LastDownlinkAt == nil || dev.MACState.LastDownlinkAt.Before(classA.Uplink.ReceivedAt)
		}
		rx2 := classA.RX2()
		switch hasClassA = dev.MACState.RxWindowsAvailable && !rx2.Before(earliestAt) && deviceHasPathForDownlink(ctx, dev, nil); {
		case !hasClassA:
		case len(dev.MACState.QueuedResponses) > 0:
			logger.Debug("MAC responses enqueued, choose class A downlink slot")
			return classA, true
		case mac.DeviceNeedsADRParamSetupReq(dev, phy):
			logger.Debug("Device needs ADRParamSetupReq, choose class A downlink slot")
			return classA, true
		case mac.DeviceNeedsBeaconFreqReq(dev):
			logger.Debug("Device needs BeaconFreqReq, choose class A downlink slot")
			return classA, true
		case mac.DeviceNeedsBeaconTimingReq(dev):
			logger.Debug("Device needs BeaconTimingReq, choose class A downlink slot")
			return classA, true
		case mac.DeviceNeedsDevStatusReq(dev, defaults, rx2):
			logger.Debug("Device needs DevStatusReq, choose class A downlink slot")
			return classA, true
		case mac.DeviceNeedsDLChannelReq(dev):
			logger.Debug("Device needs DLChannelReq, choose class A downlink slot")
			return classA, true
		case mac.DeviceNeedsDutyCycleReq(dev):
			logger.Debug("Device needs DutyCycleReq, choose class A downlink slot")
			return classA, true
		case mac.DeviceNeedsLinkADRReq(dev, defaults, phy):
			logger.Debug("Device needs LinkADRReq, choose class A downlink slot")
			return classA, true
		case mac.DeviceNeedsNewChannelReq(dev):
			logger.Debug("Device needs NewChannelReq, choose class A downlink slot")
			return classA, true
		case mac.DeviceNeedsPingSlotChannelReq(dev):
			logger.Debug("Device needs PingSlotChannelReq, choose class A downlink slot")
			return classA, true
		case mac.DeviceNeedsRejoinParamSetupReq(dev):
			logger.Debug("Device needs RejoinParamSetupReq, choose class A downlink slot")
			return classA, true
		case mac.DeviceNeedsRxParamSetupReq(dev):
			logger.Debug("Device needs RxParamSetupReq, choose class A downlink slot")
			return classA, true
		case mac.DeviceNeedsRxTimingSetupReq(dev):
			logger.Debug("Device needs RxTimingSetupReq, choose class A downlink slot")
			return classA, true
		case mac.DeviceNeedsTxParamSetupReq(dev, phy):
			logger.Debug("Device needs TxParamSetupReq, choose class A downlink slot")
			return classA, true
		}
	}

	nwkUnconf, hasNwkUnconf := nextUnconfirmedNetworkInitiatedDownlinkAt(ctx, dev, phy)
	if hasNwkUnconf && dev.MACState.DeviceClass == ttnpb.CLASS_B {
		nwkUnconf, hasNwkUnconf = mac.NextPingSlotAt(ctx, dev, latestTime(nwkUnconf, earliestAt))
	}

	nwkConf, hasNwkConf := nextConfirmedNetworkInitiatedDownlinkAt(ctx, dev, phy, defaults)
	if hasNwkConf {
		nwkConf = latestTime(nwkConf, nwkUnconf)
	}
	if hasNwkConf && dev.MACState.DeviceClass == ttnpb.CLASS_B {
		nwkConf, hasNwkConf = mac.NextPingSlotAt(ctx, dev, latestTime(nwkConf, earliestAt))
	}

	if !hasClassA && !hasNwkUnconf && !hasNwkConf {
		logger.Debug("No downlink slot available, skip downlink slot")
		return nil, false
	}
	if needsAck && deviceHasPathForDownlink(ctx, dev, nil) {
		switch {
		case hasClassA:
			return classA, true
		case hasNwkUnconf:
			return &networkInitiatedDownlinkSlot{
				Time:  nwkUnconf,
				Class: dev.MACState.DeviceClass,
			}, true
		case hasNwkConf:
			return &networkInitiatedDownlinkSlot{
				Time:  nwkConf,
				Class: dev.MACState.DeviceClass,
			}, true
		}
	}
	for _, down := range dev.GetSession().GetQueuedApplicationDownlinks() {
		if !deviceHasPathForDownlink(ctx, dev, down) {
			logger.Debug("Skip downlink, for which no path is available")
			continue
		}
		// NOTE: In case at time t, where t is before earliestConfirmedAt, device requires MAC requests,
		// Network Server will have to wait until earliestConfirmedAt, since MAC commands have priority.
		switch absTime := down.GetClassBC().GetAbsoluteTime(); {
		case absTime == nil:
			switch {
			case hasClassA && down.ClassBC == nil:
				logger.Debug("Non-constrained application downlink, choose class A downlink slot")
				return classA, true

			case hasNwkUnconf &&
				!down.Confirmed:
				logger.Debug("Application downlink with no absolute time, choose unconfirmed network-initiated downlink slot")
				return &networkInitiatedDownlinkSlot{
					Time:  nwkUnconf,
					Class: dev.MACState.DeviceClass,
				}, true
			case hasNwkConf:
				return &networkInitiatedDownlinkSlot{
					Time:  nwkConf,
					Class: dev.MACState.DeviceClass,
				}, true

			default:
				logger.Debug("Skip application with no absolute time and no available downlink slot")
				continue
			}

		case absTime.Before(earliestAt):
			logger.WithField("absolute_time", absTime).Debug("Skip application downlink with expired absolute time")
			continue

		case hasNwkUnconf && !down.Confirmed && !absTime.Before(nwkUnconf),
			hasNwkConf && !absTime.Before(nwkConf):
			logger.WithField("absolute_time", absTime).Debug("Application downlink with absolute time, choose absolute time downlink slot")
			return &networkInitiatedDownlinkSlot{
				Time:              absTime.UTC(),
				Class:             dev.MACState.DeviceClass,
				IsApplicationTime: true,
			}, true

		default:
			logger.WithField("absolute_time", absTime).Debug("Skip application with absolute time and no available downlink slot")
			continue
		}
	}
	logger.Debug("No available downlink to send, skip downlink slot")
	return nil, false
}

func publishEvents(ctx context.Context, evs ...events.Event) {
	n := len(evs)
	if n == 0 {
		return
	}
	log.FromContext(ctx).WithField("event_count", n).Debug("Publish events")
	events.Publish(evs...)
}

func (ns *NetworkServer) enqueueApplicationUplinks(ctx context.Context, ups ...*ttnpb.ApplicationUp) {
	n := len(ups)
	if n == 0 {
		return
	}
	logger := log.FromContext(ctx).WithField("uplink_count", n)
	logger.Debug("Enqueue application uplinks for sending to Application Server")
	if err := ns.applicationUplinks.Add(ctx, ups...); err != nil {
		logger.WithError(err).Warn("Failed to enqueue application uplinks for sending to Application Server")
	}
}
