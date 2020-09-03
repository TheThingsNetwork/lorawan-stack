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
	"math"
	"sort"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/gpstime"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
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

// copyUplinkMessage returns a deep copy of *ttnpb.UplinkMessage pb.
func copyUplinkMessage(pb *ttnpb.UplinkMessage) *ttnpb.UplinkMessage {
	return deepcopy.Copy(pb).(*ttnpb.UplinkMessage)
}

// copyEndDevice returns a deep copy of ttnpb.EndDevice pb.
func copyEndDevice(pb *ttnpb.EndDevice) *ttnpb.EndDevice {
	return deepcopy.Copy(pb).(*ttnpb.EndDevice)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func deviceUseADR(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings, phy *band.Band) bool {
	if !phy.EnableADR {
		return false
	}
	if dev.MACSettings != nil && dev.MACSettings.UseADR != nil {
		return dev.MACSettings.UseADR.Value
	}
	if defaults.UseADR != nil {
		return defaults.UseADR.Value
	}
	return true
}

// DefaultClassBTimeout is the default time-out for the device to respond to class B downlink messages.
// When waiting for a response times out, the downlink message is considered lost, and the downlink task triggers again.
const DefaultClassBTimeout = 10 * time.Minute

func deviceClassBTimeout(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) time.Duration {
	if dev.MACSettings != nil && dev.MACSettings.ClassBTimeout != nil {
		return *dev.MACSettings.ClassBTimeout
	}
	if defaults.ClassBTimeout != nil {
		return *defaults.ClassBTimeout
	}
	return DefaultClassBTimeout
}

// DefaultClassCTimeout is the default time-out for the device to respond to class C downlink messages.
// When waiting for a response times out, the downlink message is considered lost, and the downlink task triggers again.
const DefaultClassCTimeout = 5 * time.Minute

func deviceClassCTimeout(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) time.Duration {
	if dev.MACSettings != nil && dev.MACSettings.ClassCTimeout != nil {
		return *dev.MACSettings.ClassCTimeout
	}
	if defaults.ClassCTimeout != nil {
		return *defaults.ClassCTimeout
	}
	return DefaultClassCTimeout
}

var errNoBandVersion = errors.DefineInvalidArgument("no_band_version", "specified version `{ver}` of band `{id}` does not exist")

func deviceFrequencyPlanAndBand(dev *ttnpb.EndDevice, fps *frequencyplans.Store) (*frequencyplans.FrequencyPlan, *band.Band, error) {
	fp, err := fps.GetByID(dev.FrequencyPlanID)
	if err != nil {
		return nil, nil, err
	}
	b, ok := lorawanBands[fp.BandID][dev.LoRaWANPHYVersion]
	if !ok || b == nil {
		return nil, nil, errNoBandVersion.WithAttributes(
			"ver", dev.LoRaWANPHYVersion,
			"id", fp.BandID,
		)
	}
	return fp, b, nil
}

func deviceBand(dev *ttnpb.EndDevice, fps *frequencyplans.Store) (*band.Band, error) {
	_, phy, err := deviceFrequencyPlanAndBand(dev, fps)
	return phy, err
}

func searchUplinkChannel(freq uint64, macState *ttnpb.MACState) (uint8, error) {
	for i, ch := range macState.CurrentParameters.Channels {
		if ch.UplinkFrequency == freq {
			return uint8(i), nil
		}
	}
	return 0, errUplinkChannelNotFound.WithAttributes("frequency", freq)
}

func searchDataRateIndex(v ttnpb.DataRateIndex, vs ...ttnpb.DataRateIndex) int {
	return sort.Search(len(vs), func(i int) bool { return vs[i] >= v })
}

func searchUint32(v uint32, vs ...uint32) int {
	return sort.Search(len(vs), func(i int) bool { return vs[i] >= v })
}

func searchUint64(v uint64, vs ...uint64) int {
	return sort.Search(len(vs), func(i int) bool { return vs[i] >= v })
}

func deviceRejectedADRDataRateIndex(dev *ttnpb.EndDevice, idx ttnpb.DataRateIndex) bool {
	i := searchDataRateIndex(idx, dev.MACState.RejectedADRDataRateIndexes...)
	return i < len(dev.MACState.RejectedADRDataRateIndexes) && dev.MACState.RejectedADRDataRateIndexes[i] == idx
}

func deviceRejectedADRTXPowerIndex(dev *ttnpb.EndDevice, idx uint32) bool {
	i := searchUint32(idx, dev.MACState.RejectedADRTxPowerIndexes...)
	return i < len(dev.MACState.RejectedADRTxPowerIndexes) && dev.MACState.RejectedADRTxPowerIndexes[i] == idx
}

func deviceRejectedFrequency(dev *ttnpb.EndDevice, freq uint64) bool {
	i := searchUint64(freq, dev.MACState.RejectedFrequencies...)
	return i < len(dev.MACState.RejectedFrequencies) && dev.MACState.RejectedFrequencies[i] == freq
}

func lastUplink(ups ...*ttnpb.UplinkMessage) *ttnpb.UplinkMessage {
	return ups[len(ups)-1]
}

func lastDownlink(downs ...*ttnpb.DownlinkMessage) *ttnpb.DownlinkMessage {
	return downs[len(downs)-1]
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
	up := lastUplink(dev.MACState.RecentUplinks...)
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
		len(dev.MACState.RecentUplinks) > 0 && lastUplink(dev.MACState.RecentUplinks...).ReceivedAt.After(*dev.MACState.LastConfirmedDownlinkAt):
		return unconfAt, true
	}

	var timeout time.Duration
	switch dev.MACState.DeviceClass {
	case ttnpb.CLASS_B:
		timeout = deviceClassBTimeout(dev, defaults)

	case ttnpb.CLASS_C:
		timeout = deviceClassCTimeout(dev, defaults)
	default:
		panic(fmt.Errorf("unmatched class: %v", dev.MACState.DeviceClass))
	}
	if t := dev.MACState.LastConfirmedDownlinkAt.Add(timeout); t.After(unconfAt) {
		return t, true
	}
	return unconfAt, true
}

const (
	tBeaconDelay   = 1*time.Microsecond + 500*time.Nanosecond
	beaconPeriod   = 128 * time.Second
	beaconReserved = 2*time.Second + 120*time.Millisecond
	pingSlotCount  = 4096
	pingSlotLen    = 30 * time.Millisecond
)

// beaconTimeBefore returns the last beacon time at or before t as time.Duration since GPS epoch.
func beaconTimeBefore(t time.Time) time.Duration {
	return gpstime.ToGPS(t) / beaconPeriod * beaconPeriod
}

// nextPingSlotAt returns the exact time instant before or at earliestAt when next ping slot can be open
// given the data known by Network Server and true, if such time instant exists, otherwise it returns time.Time{} and false.
func nextPingSlotAt(ctx context.Context, dev *ttnpb.EndDevice, earliestAt time.Time) (time.Time, bool) {
	if dev.GetSession() == nil || dev.Session.DevAddr.IsZero() || dev.GetMACState() == nil || dev.MACState.PingSlotPeriodicity == nil {
		log.FromContext(ctx).Warn("Insufficient data to compute next ping slot")
		return time.Time{}, false
	}

	pingNb := uint16(1 << (7 - dev.MACState.PingSlotPeriodicity.Value))
	pingPeriod := uint16(1 << (5 + dev.MACState.PingSlotPeriodicity.Value))
	for beaconTime := beaconTimeBefore(earliestAt); beaconTime < math.MaxInt64; beaconTime += beaconPeriod {
		pingOffset, err := crypto.ComputePingOffset(uint32(beaconTime/time.Second), dev.Session.DevAddr, pingPeriod)
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Failed to compute ping offset")
			return time.Time{}, false
		}

		t := gpstime.Parse(beaconTime + tBeaconDelay + beaconReserved + time.Duration(pingOffset)*pingSlotLen).UTC()
		if !earliestAt.After(t) {
			return t, true
		}
		sub := earliestAt.Sub(t)
		if sub >= beaconPeriod {
			panic(fmt.Errorf("difference between earliestAt and first ping slot must be below '%s', got '%s'", beaconPeriod, sub))
		}
		pingPeriodDuration := time.Duration(pingPeriod) * pingSlotLen
		n := sub / pingPeriodDuration
		if int64(n) >= int64(pingNb) {
			continue
		}
		t = t.Add(n * pingPeriodDuration)
		if !earliestAt.After(t) {
			return t, true
		}
		if int64(n+1) == int64(pingNb) {
			continue
		}
		return t.Add(pingPeriodDuration), true
	}
	return time.Time{}, false
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
		case deviceNeedsADRParamSetupReq(dev, phy):
			logger.Debug("Device needs ADRParamSetupReq, choose class A downlink slot")
			return classA, true
		case deviceNeedsBeaconFreqReq(dev):
			logger.Debug("Device needs BeaconFreqReq, choose class A downlink slot")
			return classA, true
		case deviceNeedsBeaconTimingReq(dev):
			logger.Debug("Device needs BeaconTimingReq, choose class A downlink slot")
			return classA, true
		case deviceNeedsDevStatusReq(dev, defaults, rx2):
			logger.Debug("Device needs DevStatusReq, choose class A downlink slot")
			return classA, true
		case deviceNeedsDLChannelReq(dev):
			logger.Debug("Device needs DLChannelReq, choose class A downlink slot")
			return classA, true
		case deviceNeedsDutyCycleReq(dev):
			logger.Debug("Device needs DutyCycleReq, choose class A downlink slot")
			return classA, true
		case deviceNeedsLinkADRReq(dev, defaults, phy):
			logger.Debug("Device needs LinkADRReq, choose class A downlink slot")
			return classA, true
		case deviceNeedsNewChannelReq(dev):
			logger.Debug("Device needs NewChannelReq, choose class A downlink slot")
			return classA, true
		case deviceNeedsPingSlotChannelReq(dev):
			logger.Debug("Device needs PingSlotChannelReq, choose class A downlink slot")
			return classA, true
		case deviceNeedsRejoinParamSetupReq(dev):
			logger.Debug("Device needs RejoinParamSetupReq, choose class A downlink slot")
			return classA, true
		case deviceNeedsRxParamSetupReq(dev):
			logger.Debug("Device needs RxParamSetupReq, choose class A downlink slot")
			return classA, true
		case deviceNeedsRxTimingSetupReq(dev):
			logger.Debug("Device needs RxTimingSetupReq, choose class A downlink slot")
			return classA, true
		case deviceNeedsTxParamSetupReq(dev, phy):
			logger.Debug("Device needs TxParamSetupReq, choose class A downlink slot")
			return classA, true
		}
	}

	nwkUnconf, hasNwkUnconf := nextUnconfirmedNetworkInitiatedDownlinkAt(ctx, dev, phy)
	if hasNwkUnconf && dev.MACState.DeviceClass == ttnpb.CLASS_B {
		nwkUnconf, hasNwkUnconf = nextPingSlotAt(ctx, dev, latestTime(nwkUnconf, earliestAt))
	}

	nwkConf, hasNwkConf := nextConfirmedNetworkInitiatedDownlinkAt(ctx, dev, phy, defaults)
	if hasNwkConf {
		nwkConf = latestTime(nwkConf, nwkUnconf)
	}
	if hasNwkConf && dev.MACState.DeviceClass == ttnpb.CLASS_B {
		nwkConf, hasNwkConf = nextPingSlotAt(ctx, dev, latestTime(nwkConf, earliestAt))
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

func deviceDefaultClass(dev *ttnpb.EndDevice) (ttnpb.Class, error) {
	switch {
	case dev.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 && dev.SupportsClassC:
		return ttnpb.CLASS_C, nil
	case !dev.Multicast:
		return ttnpb.CLASS_A, nil
	case dev.SupportsClassC:
		return ttnpb.CLASS_C, nil
	case dev.SupportsClassB:
		return ttnpb.CLASS_B, nil
	default:
		return ttnpb.CLASS_A, errClassAMulticast.New()
	}
}

func deviceDefaultLoRaWANVersion(dev *ttnpb.EndDevice) ttnpb.MACVersion {
	switch {
	case dev.Multicast:
		return dev.LoRaWANVersion
	case dev.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0:
		return ttnpb.MAC_V1_1
	default:
		return dev.LoRaWANVersion
	}
}

func deviceDefaultPingSlotPeriodicity(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) *ttnpb.PingSlotPeriodValue {
	switch {
	case dev.GetMACSettings().GetPingSlotPeriodicity() != nil:
		return dev.MACSettings.PingSlotPeriodicity
	case defaults.GetPingSlotPeriodicity() != nil:
		return defaults.PingSlotPeriodicity
	default:
		return nil
	}
}

func deviceDesiredMaxEIRP(dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults ttnpb.MACSettings) float32 {
	switch {
	case fp.MaxEIRP != nil && *fp.MaxEIRP > 0 && *fp.MaxEIRP < phy.DefaultMaxEIRP:
		return *fp.MaxEIRP
	default:
		return phy.DefaultMaxEIRP
	}
}

func deviceDesiredUplinkDwellTime(fp *frequencyplans.FrequencyPlan) *pbtypes.BoolValue {
	if fp.DwellTime.Uplinks == nil {
		return nil
	}
	return &pbtypes.BoolValue{Value: *fp.DwellTime.Uplinks}
}

func deviceDesiredDownlinkDwellTime(fp *frequencyplans.FrequencyPlan) *pbtypes.BoolValue {
	if fp.DwellTime.Downlinks == nil {
		return nil
	}
	return &pbtypes.BoolValue{Value: *fp.DwellTime.Downlinks}
}

func deviceDefaultRX1Delay(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) ttnpb.RxDelay {
	switch {
	case dev.GetMACSettings().GetRx1Delay() != nil:
		return dev.MACSettings.Rx1Delay.Value
	case defaults.Rx1Delay != nil:
		return defaults.Rx1Delay.Value
	default:
		return ttnpb.RxDelay(phy.ReceiveDelay1.Seconds())
	}
}

func deviceDesiredRX1Delay(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) ttnpb.RxDelay {
	switch {
	case dev.GetMACSettings().GetDesiredRx1Delay() != nil:
		return dev.MACSettings.DesiredRx1Delay.Value
	case defaults.DesiredRx1Delay != nil:
		return defaults.DesiredRx1Delay.Value
	default:
		return deviceDefaultRX1Delay(dev, phy, defaults)
	}
}

func deviceDesiredADRAckLimitExponent(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) *ttnpb.ADRAckLimitExponentValue {
	switch {
	case dev.GetMACSettings().GetDesiredADRAckLimitExponent() != nil:
		return &ttnpb.ADRAckLimitExponentValue{Value: dev.MACSettings.DesiredADRAckLimitExponent.Value}
	case defaults.DesiredADRAckLimitExponent != nil:
		return &ttnpb.ADRAckLimitExponentValue{Value: defaults.DesiredADRAckLimitExponent.Value}
	default:
		return &ttnpb.ADRAckLimitExponentValue{Value: phy.ADRAckLimit}
	}
}

func deviceDesiredADRAckDelayExponent(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) *ttnpb.ADRAckDelayExponentValue {
	switch {
	case dev.GetMACSettings().GetDesiredADRAckDelayExponent() != nil:
		return &ttnpb.ADRAckDelayExponentValue{Value: dev.MACSettings.DesiredADRAckDelayExponent.Value}
	case defaults.DesiredADRAckDelayExponent != nil:
		return &ttnpb.ADRAckDelayExponentValue{Value: defaults.DesiredADRAckDelayExponent.Value}
	default:
		return &ttnpb.ADRAckDelayExponentValue{Value: phy.ADRAckDelay}
	}
}

func deviceDefaultRX1DataRateOffset(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) uint32 {
	switch {
	case dev.GetMACSettings().GetRx1DataRateOffset() != nil:
		return dev.MACSettings.Rx1DataRateOffset.Value
	case defaults.Rx1DataRateOffset != nil:
		return defaults.Rx1DataRateOffset.Value
	default:
		return 0
	}
}

func deviceDesiredRX1DataRateOffset(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) uint32 {
	switch {
	case dev.GetMACSettings().GetDesiredRx1DataRateOffset() != nil:
		return dev.MACSettings.DesiredRx1DataRateOffset.Value
	case defaults.DesiredRx1DataRateOffset != nil:
		return defaults.DesiredRx1DataRateOffset.Value
	default:
		return deviceDefaultRX1DataRateOffset(dev, defaults)
	}
}

func deviceDefaultRX2DataRateIndex(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) ttnpb.DataRateIndex {
	switch {
	case dev.GetMACSettings().GetRx2DataRateIndex() != nil:
		return dev.MACSettings.Rx2DataRateIndex.Value
	case defaults.Rx2DataRateIndex != nil:
		return defaults.Rx2DataRateIndex.Value
	default:
		return phy.DefaultRx2Parameters.DataRateIndex
	}
}

func deviceDesiredRX2DataRateIndex(dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults ttnpb.MACSettings) ttnpb.DataRateIndex {
	switch {
	case dev.GetMACSettings().GetDesiredRx2DataRateIndex() != nil:
		return dev.MACSettings.DesiredRx2DataRateIndex.Value
	case fp.DefaultRx2DataRate != nil:
		return ttnpb.DataRateIndex(*fp.DefaultRx2DataRate)
	case defaults.DesiredRx2DataRateIndex != nil:
		return defaults.DesiredRx2DataRateIndex.Value
	default:
		return deviceDefaultRX2DataRateIndex(dev, phy, defaults)
	}
}

func deviceDefaultRX2Frequency(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMACSettings().GetRx2Frequency() != nil && dev.MACSettings.Rx2Frequency.Value != 0:
		return dev.MACSettings.Rx2Frequency.Value
	case defaults.Rx2Frequency != nil && dev.MACSettings.Rx2Frequency.Value != 0:
		return defaults.Rx2Frequency.Value
	default:
		return phy.DefaultRx2Parameters.Frequency
	}
}

func deviceDesiredRX2Frequency(dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMACSettings().GetDesiredRx2Frequency() != nil && dev.MACSettings.DesiredRx2Frequency.Value != 0:
		return dev.MACSettings.DesiredRx2Frequency.Value
	case fp.Rx2Channel != nil:
		return fp.Rx2Channel.Frequency
	case defaults.DesiredRx2Frequency != nil && defaults.DesiredRx2Frequency.Value != 0:
		return defaults.DesiredRx2Frequency.Value
	default:
		return deviceDefaultRX2Frequency(dev, phy, defaults)
	}
}

func deviceDefaultMaxDutyCycle(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) ttnpb.AggregatedDutyCycle {
	switch {
	case dev.GetMACSettings().GetMaxDutyCycle() != nil:
		return dev.MACSettings.MaxDutyCycle.Value
	case defaults.MaxDutyCycle != nil:
		return defaults.MaxDutyCycle.Value
	default:
		return ttnpb.DUTY_CYCLE_1
	}
}

func deviceDesiredMaxDutyCycle(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) ttnpb.AggregatedDutyCycle {
	switch {
	case dev.GetMACSettings().GetDesiredMaxDutyCycle() != nil:
		return dev.MACSettings.DesiredMaxDutyCycle.Value
	case defaults.DesiredMaxDutyCycle != nil:
		return defaults.DesiredMaxDutyCycle.Value
	default:
		return deviceDefaultMaxDutyCycle(dev, defaults)
	}
}

func deviceDefaultPingSlotFrequency(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMACSettings().GetPingSlotFrequency() != nil && dev.MACSettings.PingSlotFrequency.Value != 0:
		return dev.MACSettings.PingSlotFrequency.Value
	case defaults.PingSlotFrequency != nil && defaults.PingSlotFrequency.Value != 0:
		return defaults.PingSlotFrequency.Value
	case phy.PingSlotFrequency != nil:
		return *phy.PingSlotFrequency
	default:
		return 0
	}
}

func deviceDesiredPingSlotFrequency(dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMACSettings().GetDesiredPingSlotFrequency() != nil && dev.MACSettings.DesiredPingSlotFrequency.Value != 0:
		return dev.MACSettings.DesiredPingSlotFrequency.Value
	case fp.PingSlot != nil && fp.PingSlot.Frequency != 0:
		return fp.PingSlot.Frequency
	case defaults.DesiredPingSlotFrequency != nil && defaults.DesiredPingSlotFrequency.Value != 0:
		return defaults.DesiredPingSlotFrequency.Value
	default:
		return deviceDefaultPingSlotFrequency(dev, phy, defaults)
	}
}

func deviceDefaultPingSlotDataRateIndexValue(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) *ttnpb.DataRateIndexValue {
	switch {
	case dev.GetMACSettings().GetPingSlotDataRateIndex() != nil:
		return dev.MACSettings.PingSlotDataRateIndex
	case defaults.PingSlotDataRateIndex != nil:
		return defaults.PingSlotDataRateIndex
	default:
		// Default to mbed-os and LoRaMac-node behavior.
		// https://github.com/Lora-net/LoRaMac-node/blob/87f19e84ae2fc4af72af9567fe722386de6ce9f4/src/mac/region/RegionCN779.h#L235.
		return &ttnpb.DataRateIndexValue{Value: phy.Beacon.DataRateIndex}
	}
}

func deviceDesiredPingSlotDataRateIndexValue(dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults ttnpb.MACSettings) *ttnpb.DataRateIndexValue {
	switch {
	case dev.GetMACSettings().GetDesiredPingSlotDataRateIndex() != nil:
		return dev.MACSettings.DesiredPingSlotDataRateIndex
	case fp.DefaultPingSlotDataRate != nil:
		return &ttnpb.DataRateIndexValue{Value: ttnpb.DataRateIndex(*fp.DefaultPingSlotDataRate)}
	case defaults.DesiredPingSlotDataRateIndex != nil:
		return defaults.DesiredPingSlotDataRateIndex
	default:
		return deviceDefaultPingSlotDataRateIndexValue(dev, phy, defaults)
	}
}

func deviceDefaultBeaconFrequency(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMACSettings().GetBeaconFrequency() != nil && dev.MACSettings.BeaconFrequency.Value != 0:
		return dev.MACSettings.BeaconFrequency.Value
	case defaults.BeaconFrequency != nil:
		return defaults.BeaconFrequency.Value
	default:
		return 0
	}
}

func deviceDesiredBeaconFrequency(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMACSettings().GetDesiredBeaconFrequency() != nil && dev.MACSettings.DesiredBeaconFrequency.Value != 0:
		return dev.MACSettings.DesiredBeaconFrequency.Value
	case defaults.DesiredBeaconFrequency != nil && defaults.DesiredBeaconFrequency.Value != 0:
		return defaults.DesiredBeaconFrequency.Value
	default:
		return deviceDefaultBeaconFrequency(dev, defaults)
	}
}

func deviceDefaultChannels(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) []*ttnpb.MACParameters_Channel {
	// NOTE: FactoryPresetFrequencies does not indicate the data rate ranges allowed for channels.
	// In the latest regional parameters spec(1.1b) the data rate ranges are DR0-DR5 for mandatory channels in all non-fixed channel plans,
	// hence we assume the same range for predefined channels.
	var chs []*ttnpb.MACParameters_Channel
	switch {
	case len(dev.GetMACSettings().GetFactoryPresetFrequencies()) > 0:
		chs = make([]*ttnpb.MACParameters_Channel, 0, len(dev.MACSettings.FactoryPresetFrequencies))
		for _, freq := range dev.MACSettings.FactoryPresetFrequencies {
			chs = append(chs, &ttnpb.MACParameters_Channel{
				MaxDataRateIndex:  ttnpb.DATA_RATE_5,
				UplinkFrequency:   freq,
				DownlinkFrequency: freq,
				EnableUplink:      true,
			})
		}
	case len(defaults.GetFactoryPresetFrequencies()) > 0:
		chs = make([]*ttnpb.MACParameters_Channel, 0, len(defaults.FactoryPresetFrequencies))
		for _, freq := range defaults.FactoryPresetFrequencies {
			chs = append(chs, &ttnpb.MACParameters_Channel{
				MaxDataRateIndex:  ttnpb.DATA_RATE_5,
				UplinkFrequency:   freq,
				DownlinkFrequency: freq,
				EnableUplink:      true,
			})
		}
	default:
		if len(phy.DownlinkChannels) > len(phy.UplinkChannels) ||
			len(phy.UplinkChannels) > int(phy.MaxUplinkChannels) ||
			len(phy.DownlinkChannels) > int(phy.MaxDownlinkChannels) {
			// NOTE: In case the spec changes and this assumption is not valid anymore,
			// the implementation of this function won't be valid and has to be changed.
			panic("uplink/downlink channel length is inconsistent")
		}
		chs = make([]*ttnpb.MACParameters_Channel, 0, len(phy.UplinkChannels))
		for i, phyUpCh := range phy.UplinkChannels {
			chs = append(chs, &ttnpb.MACParameters_Channel{
				MinDataRateIndex:  phyUpCh.MinDataRate,
				MaxDataRateIndex:  phyUpCh.MaxDataRate,
				UplinkFrequency:   phyUpCh.Frequency,
				DownlinkFrequency: phy.DownlinkChannels[i%len(phy.DownlinkChannels)].Frequency,
				EnableUplink:      true,
			})
		}
	}
	return chs
}

func deviceDesiredChannels(phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults ttnpb.MACSettings) []*ttnpb.MACParameters_Channel {
	if len(phy.DownlinkChannels) > len(phy.UplinkChannels) || len(fp.DownlinkChannels) > len(fp.UplinkChannels) ||
		len(phy.UplinkChannels) > int(phy.MaxUplinkChannels) || len(phy.DownlinkChannels) > int(phy.MaxDownlinkChannels) ||
		len(fp.UplinkChannels) > int(phy.MaxUplinkChannels) || len(fp.DownlinkChannels) > int(phy.MaxDownlinkChannels) {
		// NOTE: In case the spec changes and this assumption is not valid anymore,
		// the implementation of this function won't be valid and has to be changed.
		panic("uplink/downlink channel length is inconsistent")
	}

	chs := make([]*ttnpb.MACParameters_Channel, 0, len(phy.UplinkChannels)+len(fp.UplinkChannels))
	for i, phyUpCh := range phy.UplinkChannels {
		chs = append(chs, &ttnpb.MACParameters_Channel{
			MinDataRateIndex:  phyUpCh.MinDataRate,
			MaxDataRateIndex:  phyUpCh.MaxDataRate,
			UplinkFrequency:   phyUpCh.Frequency,
			DownlinkFrequency: phy.DownlinkChannels[i%len(phy.DownlinkChannels)].Frequency,
		})
	}

outerUp:
	for _, fpUpCh := range fp.UplinkChannels {
		for _, ch := range chs {
			if ch.UplinkFrequency == fpUpCh.Frequency {
				ch.MinDataRateIndex = ttnpb.DataRateIndex(fpUpCh.MinDataRate)
				ch.MaxDataRateIndex = ttnpb.DataRateIndex(fpUpCh.MaxDataRate)
				ch.EnableUplink = true
				continue outerUp
			}
		}
		chs = append(chs, &ttnpb.MACParameters_Channel{
			MinDataRateIndex:  ttnpb.DataRateIndex(fpUpCh.MinDataRate),
			MaxDataRateIndex:  ttnpb.DataRateIndex(fpUpCh.MaxDataRate),
			UplinkFrequency:   fpUpCh.Frequency,
			DownlinkFrequency: phy.DownlinkChannels[len(chs)%len(phy.DownlinkChannels)].Frequency,
			EnableUplink:      true,
		})
	}
	if len(fp.DownlinkChannels) > 0 {
		for i, ch := range chs {
			ch.DownlinkFrequency = fp.DownlinkChannels[i%len(fp.DownlinkChannels)].Frequency
		}
	}
	return chs
}

func newMACState(dev *ttnpb.EndDevice, fps *frequencyplans.Store, defaults ttnpb.MACSettings) (*ttnpb.MACState, error) {
	fp, phy, err := deviceFrequencyPlanAndBand(dev, fps)
	if err != nil {
		return nil, err
	}
	class, err := deviceDefaultClass(dev)
	if err != nil {
		return nil, err
	}
	// TODO: Support rejoins. (https://github.com/TheThingsNetwork/lorawan-stack/issues/8)
	return &ttnpb.MACState{
		LoRaWANVersion:      deviceDefaultLoRaWANVersion(dev),
		DeviceClass:         class,
		PingSlotPeriodicity: deviceDefaultPingSlotPeriodicity(dev, defaults),
		CurrentParameters: ttnpb.MACParameters{
			MaxEIRP:                    phy.DefaultMaxEIRP,
			ADRDataRateIndex:           ttnpb.DATA_RATE_0,
			ADRNbTrans:                 1,
			Rx1Delay:                   deviceDefaultRX1Delay(dev, phy, defaults),
			Rx1DataRateOffset:          deviceDefaultRX1DataRateOffset(dev, defaults),
			Rx2DataRateIndex:           deviceDefaultRX2DataRateIndex(dev, phy, defaults),
			Rx2Frequency:               deviceDefaultRX2Frequency(dev, phy, defaults),
			MaxDutyCycle:               deviceDefaultMaxDutyCycle(dev, defaults),
			RejoinTimePeriodicity:      ttnpb.REJOIN_TIME_0,
			RejoinCountPeriodicity:     ttnpb.REJOIN_COUNT_16,
			PingSlotFrequency:          deviceDefaultPingSlotFrequency(dev, phy, defaults),
			BeaconFrequency:            deviceDefaultBeaconFrequency(dev, defaults),
			Channels:                   deviceDefaultChannels(dev, phy, defaults),
			ADRAckLimitExponent:        &ttnpb.ADRAckLimitExponentValue{Value: phy.ADRAckLimit},
			ADRAckDelayExponent:        &ttnpb.ADRAckDelayExponentValue{Value: phy.ADRAckDelay},
			PingSlotDataRateIndexValue: deviceDefaultPingSlotDataRateIndexValue(dev, phy, defaults),
		},
		DesiredParameters: ttnpb.MACParameters{
			MaxEIRP:                    deviceDesiredMaxEIRP(dev, phy, fp, defaults),
			ADRDataRateIndex:           ttnpb.DATA_RATE_0,
			ADRNbTrans:                 1,
			Rx1Delay:                   deviceDesiredRX1Delay(dev, phy, defaults),
			Rx1DataRateOffset:          deviceDesiredRX1DataRateOffset(dev, defaults),
			Rx2DataRateIndex:           deviceDesiredRX2DataRateIndex(dev, phy, fp, defaults),
			Rx2Frequency:               deviceDesiredRX2Frequency(dev, phy, fp, defaults),
			MaxDutyCycle:               deviceDesiredMaxDutyCycle(dev, defaults),
			RejoinTimePeriodicity:      ttnpb.REJOIN_TIME_0,
			RejoinCountPeriodicity:     ttnpb.REJOIN_COUNT_16,
			PingSlotFrequency:          deviceDesiredPingSlotFrequency(dev, phy, fp, defaults),
			BeaconFrequency:            deviceDesiredBeaconFrequency(dev, defaults),
			Channels:                   deviceDesiredChannels(phy, fp, defaults),
			UplinkDwellTime:            deviceDesiredUplinkDwellTime(fp),
			DownlinkDwellTime:          deviceDesiredDownlinkDwellTime(fp),
			ADRAckLimitExponent:        deviceDesiredADRAckLimitExponent(dev, phy, defaults),
			ADRAckDelayExponent:        deviceDesiredADRAckDelayExponent(dev, phy, defaults),
			PingSlotDataRateIndexValue: deviceDesiredPingSlotDataRateIndexValue(dev, phy, fp, defaults),
		},
	}, nil
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

func rxMetadataStats(ctx context.Context, mds []*ttnpb.RxMetadata) (gateways int, maxSNR float32) {
	if len(mds) == 0 {
		return 0, 0
	}
	gtws := make(map[string]struct{}, len(mds))
	maxSNR = mds[0].SNR
	for _, md := range mds {
		if md.PacketBroker != nil {
			gtws[fmt.Sprintf("%s@%s/%s", md.PacketBroker.ForwarderID, md.PacketBroker.ForwarderNetID, md.PacketBroker.ForwarderTenantID)] = struct{}{}
		} else {
			gtws[unique.ID(ctx, md.GatewayIdentifiers)] = struct{}{}
		}
		if md.SNR > maxSNR {
			maxSNR = md.SNR
		}
	}
	return len(gtws), maxSNR
}
