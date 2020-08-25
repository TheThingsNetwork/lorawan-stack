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

func deviceNeedsMACRequestsAt(ctx context.Context, dev *ttnpb.EndDevice, earliestAt time.Time, phy *band.Band, defaults ttnpb.MACSettings) bool {
	if dev.GetMulticast() {
		return false
	}
	now := timeNow()
	if earliestAt.Before(now) {
		earliestAt = now
	}
	switch {
	case deviceNeedsADRParamSetupReq(dev, phy),
		deviceNeedsBeaconFreqReq(dev),
		deviceNeedsBeaconTimingReq(dev),
		deviceNeedsDevStatusReq(dev, defaults, earliestAt),
		deviceNeedsDLChannelReq(dev),
		deviceNeedsDutyCycleReq(dev),
		deviceNeedsLinkADRReq(dev, defaults, phy),
		deviceNeedsNewChannelReq(dev),
		deviceNeedsPingSlotChannelReq(dev),
		deviceNeedsRejoinParamSetupReq(dev),
		deviceNeedsRxParamSetupReq(dev),
		deviceNeedsRxTimingSetupReq(dev),
		deviceNeedsTxParamSetupReq(dev, phy):
		return true
	}
	return false
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
		hasClassA = dev.MACState.RxWindowsAvailable && !rx2.Before(earliestAt) && deviceHasPathForDownlink(ctx, dev, nil)
		if hasClassA && (len(dev.MACState.QueuedResponses) > 0 || deviceNeedsMACRequestsAt(ctx, dev, rx2, phy, defaults)) {
			logger.Debug("MAC commands required, choose class A downlink slot")
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
				!down.Confirmed &&
				!deviceNeedsMACRequestsAt(ctx, dev, nwkUnconf, phy, defaults):
				logger.Debug("Application downlink with no absolute time, choose unconfirmed network-initiated downlink slot")
				return &networkInitiatedDownlinkSlot{
					Time:  nwkUnconf,
					Class: dev.MACState.DeviceClass,
				}, true
			case hasNwkConf:
				logger.Debug("Application downlink with no absolute time, choose confirmed network-initiated downlink slot")
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

		case hasNwkUnconf && !down.Confirmed && !deviceNeedsMACRequestsAt(ctx, dev, nwkUnconf, phy, defaults) && !absTime.Before(nwkUnconf),
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

func frequencyPlanChannels(phy *band.Band, fpUpChs []frequencyplans.Channel, fpDownChs ...frequencyplans.Channel) []*ttnpb.MACParameters_Channel {
	chs := make([]*ttnpb.MACParameters_Channel, 0, len(phy.UplinkChannels)+len(fpUpChs))
	for i, phyUpCh := range phy.UplinkChannels {
		chs = append(chs, &ttnpb.MACParameters_Channel{
			MinDataRateIndex:  phyUpCh.MinDataRate,
			MaxDataRateIndex:  phyUpCh.MaxDataRate,
			UplinkFrequency:   phyUpCh.Frequency,
			DownlinkFrequency: phy.DownlinkChannels[i%len(phy.DownlinkChannels)].Frequency,
		})
	}

outerUp:
	for _, fpUpCh := range fpUpChs {
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
	if len(fpDownChs) > 0 {
		for i, ch := range chs {
			ch.DownlinkFrequency = fpDownChs[i%len(fpDownChs)].Frequency
		}
	}
	return chs
}

func newMACState(dev *ttnpb.EndDevice, fps *frequencyplans.Store, defaults ttnpb.MACSettings) (*ttnpb.MACState, error) {
	fp, phy, err := deviceFrequencyPlanAndBand(dev, fps)
	if err != nil {
		return nil, err
	}

	class := ttnpb.CLASS_A
	if dev.Multicast {
		if dev.SupportsClassC {
			class = ttnpb.CLASS_C
		} else if dev.SupportsClassB {
			class = ttnpb.CLASS_B
		} else {
			return nil, errClassAMulticast.New()
		}
	} else if dev.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 && dev.SupportsClassC {
		class = ttnpb.CLASS_C
	}

	macState := &ttnpb.MACState{
		LoRaWANVersion: dev.LoRaWANVersion,
		DeviceClass:    class,
	}
	if dev.GetMACSettings().GetPingSlotPeriodicity() != nil {
		macState.PingSlotPeriodicity = dev.MACSettings.PingSlotPeriodicity
	} else if defaults.GetPingSlotPeriodicity() != nil {
		macState.PingSlotPeriodicity = defaults.PingSlotPeriodicity
	}

	macState.CurrentParameters.MaxEIRP = phy.DefaultMaxEIRP
	if fp.MaxEIRP != nil && *fp.MaxEIRP > 0 && *fp.MaxEIRP < macState.CurrentParameters.MaxEIRP {
		macState.DesiredParameters.MaxEIRP = *fp.MaxEIRP
	} else {
		macState.DesiredParameters.MaxEIRP = macState.CurrentParameters.MaxEIRP
	}
	if fp.DwellTime.Uplinks != nil {
		macState.DesiredParameters.UplinkDwellTime = &pbtypes.BoolValue{Value: *fp.DwellTime.Uplinks}
	}
	if fp.DwellTime.Downlinks != nil {
		macState.DesiredParameters.DownlinkDwellTime = &pbtypes.BoolValue{Value: *fp.DwellTime.Downlinks}
	}

	macState.CurrentParameters.ADRDataRateIndex = ttnpb.DATA_RATE_0
	macState.DesiredParameters.ADRDataRateIndex = macState.CurrentParameters.ADRDataRateIndex

	macState.CurrentParameters.ADRTxPowerIndex = 0
	macState.DesiredParameters.ADRTxPowerIndex = macState.CurrentParameters.ADRTxPowerIndex

	macState.CurrentParameters.ADRNbTrans = 1
	macState.DesiredParameters.ADRNbTrans = macState.CurrentParameters.ADRNbTrans

	macState.CurrentParameters.ADRAckLimitExponent = &ttnpb.ADRAckLimitExponentValue{Value: phy.ADRAckLimit}
	if dev.GetMACSettings().GetDesiredADRAckLimitExponent() != nil {
		macState.DesiredParameters.ADRAckLimitExponent = &ttnpb.ADRAckLimitExponentValue{Value: dev.MACSettings.DesiredADRAckLimitExponent.Value}
	} else if defaults.DesiredADRAckLimitExponent != nil {
		macState.DesiredParameters.ADRAckLimitExponent = &ttnpb.ADRAckLimitExponentValue{Value: defaults.DesiredADRAckLimitExponent.Value}
	} else {
		macState.DesiredParameters.ADRAckLimitExponent = &ttnpb.ADRAckLimitExponentValue{Value: phy.ADRAckLimit}
	}

	macState.CurrentParameters.ADRAckDelayExponent = &ttnpb.ADRAckDelayExponentValue{Value: phy.ADRAckDelay}
	if dev.GetMACSettings().GetDesiredADRAckDelayExponent() != nil {
		macState.DesiredParameters.ADRAckDelayExponent = &ttnpb.ADRAckDelayExponentValue{Value: dev.MACSettings.DesiredADRAckDelayExponent.Value}
	} else if defaults.DesiredADRAckDelayExponent != nil {
		macState.DesiredParameters.ADRAckDelayExponent = &ttnpb.ADRAckDelayExponentValue{Value: defaults.DesiredADRAckDelayExponent.Value}
	} else {
		macState.DesiredParameters.ADRAckDelayExponent = &ttnpb.ADRAckDelayExponentValue{Value: phy.ADRAckDelay}
	}

	if dev.GetMACSettings().GetRx1Delay() != nil {
		macState.CurrentParameters.Rx1Delay = dev.MACSettings.Rx1Delay.Value
	} else if defaults.Rx1Delay != nil {
		macState.CurrentParameters.Rx1Delay = defaults.Rx1Delay.Value
	} else {
		macState.CurrentParameters.Rx1Delay = ttnpb.RxDelay(phy.ReceiveDelay1.Seconds())
	}
	if dev.GetMACSettings().GetDesiredRx1Delay() != nil {
		macState.DesiredParameters.Rx1Delay = dev.MACSettings.DesiredRx1Delay.Value
	} else if defaults.DesiredRx1Delay != nil {
		macState.DesiredParameters.Rx1Delay = defaults.DesiredRx1Delay.Value
	} else {
		macState.DesiredParameters.Rx1Delay = macState.CurrentParameters.Rx1Delay
	}

	if dev.GetMACSettings().GetRx1DataRateOffset() != nil {
		macState.CurrentParameters.Rx1DataRateOffset = dev.MACSettings.Rx1DataRateOffset.Value
	} else if defaults.Rx1DataRateOffset != nil {
		macState.CurrentParameters.Rx1DataRateOffset = defaults.Rx1DataRateOffset.Value
	}
	if dev.GetMACSettings().GetDesiredRx1DataRateOffset() != nil {
		macState.DesiredParameters.Rx1DataRateOffset = dev.MACSettings.DesiredRx1DataRateOffset.Value
	} else if defaults.DesiredRx1DataRateOffset != nil {
		macState.DesiredParameters.Rx1DataRateOffset = defaults.DesiredRx1DataRateOffset.Value
	} else {
		macState.DesiredParameters.Rx1DataRateOffset = macState.CurrentParameters.Rx1DataRateOffset
	}

	if dev.GetMACSettings().GetRx2DataRateIndex() != nil {
		macState.CurrentParameters.Rx2DataRateIndex = dev.MACSettings.Rx2DataRateIndex.Value
	} else if defaults.Rx2DataRateIndex != nil {
		macState.CurrentParameters.Rx2DataRateIndex = defaults.Rx2DataRateIndex.Value
	} else {
		macState.CurrentParameters.Rx2DataRateIndex = phy.DefaultRx2Parameters.DataRateIndex
	}
	if dev.GetMACSettings().GetDesiredRx2DataRateIndex() != nil {
		macState.DesiredParameters.Rx2DataRateIndex = dev.MACSettings.DesiredRx2DataRateIndex.Value
	} else if fp.DefaultRx2DataRate != nil {
		macState.DesiredParameters.Rx2DataRateIndex = ttnpb.DataRateIndex(*fp.DefaultRx2DataRate)
	} else if defaults.DesiredRx2DataRateIndex != nil {
		macState.DesiredParameters.Rx2DataRateIndex = defaults.DesiredRx2DataRateIndex.Value
	} else {
		macState.DesiredParameters.Rx2DataRateIndex = macState.CurrentParameters.Rx2DataRateIndex
	}

	if dev.GetMACSettings().GetRx2Frequency() != nil && dev.MACSettings.Rx2Frequency.Value != 0 {
		macState.CurrentParameters.Rx2Frequency = dev.MACSettings.Rx2Frequency.Value
	} else if defaults.Rx2Frequency != nil && dev.MACSettings.Rx2Frequency.Value != 0 {
		macState.CurrentParameters.Rx2Frequency = defaults.Rx2Frequency.Value
	} else {
		macState.CurrentParameters.Rx2Frequency = phy.DefaultRx2Parameters.Frequency
	}
	if dev.GetMACSettings().GetDesiredRx2Frequency() != nil && dev.MACSettings.DesiredRx2Frequency.Value != 0 {
		macState.DesiredParameters.Rx2Frequency = dev.MACSettings.DesiredRx2Frequency.Value
	} else if fp.Rx2Channel != nil {
		macState.DesiredParameters.Rx2Frequency = fp.Rx2Channel.Frequency
	} else if defaults.DesiredRx2Frequency != nil && defaults.DesiredRx2Frequency.Value != 0 {
		macState.DesiredParameters.Rx2Frequency = defaults.DesiredRx2Frequency.Value
	} else {
		macState.DesiredParameters.Rx2Frequency = macState.CurrentParameters.Rx2Frequency
	}

	if dev.GetMACSettings().GetMaxDutyCycle() != nil {
		macState.CurrentParameters.MaxDutyCycle = dev.MACSettings.MaxDutyCycle.Value
	} else if defaults.MaxDutyCycle != nil {
		macState.CurrentParameters.MaxDutyCycle = defaults.MaxDutyCycle.Value
	} else {
		macState.CurrentParameters.MaxDutyCycle = ttnpb.DUTY_CYCLE_1
	}
	if dev.GetMACSettings().GetDesiredMaxDutyCycle() != nil {
		macState.DesiredParameters.MaxDutyCycle = dev.MACSettings.DesiredMaxDutyCycle.Value
	} else if defaults.DesiredMaxDutyCycle != nil {
		macState.DesiredParameters.MaxDutyCycle = defaults.DesiredMaxDutyCycle.Value
	} else {
		macState.DesiredParameters.MaxDutyCycle = macState.CurrentParameters.MaxDutyCycle
	}

	// TODO: Support rejoins. (https://github.com/TheThingsNetwork/lorawan-stack/issues/8)
	macState.CurrentParameters.RejoinTimePeriodicity = ttnpb.REJOIN_TIME_0
	macState.DesiredParameters.RejoinTimePeriodicity = macState.CurrentParameters.RejoinTimePeriodicity

	macState.CurrentParameters.RejoinCountPeriodicity = ttnpb.REJOIN_COUNT_16
	macState.DesiredParameters.RejoinCountPeriodicity = macState.CurrentParameters.RejoinCountPeriodicity

	if dev.GetMACSettings().GetPingSlotFrequency() != nil && dev.MACSettings.PingSlotFrequency.Value != 0 {
		macState.CurrentParameters.PingSlotFrequency = dev.MACSettings.PingSlotFrequency.Value
	} else if defaults.PingSlotFrequency != nil && defaults.PingSlotFrequency.Value != 0 {
		macState.CurrentParameters.PingSlotFrequency = defaults.PingSlotFrequency.Value
	} else if phy.PingSlotFrequency != nil {
		macState.CurrentParameters.PingSlotFrequency = *phy.PingSlotFrequency
	}
	if dev.GetMACSettings().GetDesiredPingSlotFrequency() != nil && dev.MACSettings.DesiredPingSlotFrequency.Value != 0 {
		macState.DesiredParameters.PingSlotFrequency = dev.MACSettings.DesiredPingSlotFrequency.Value
	} else if fp.PingSlot != nil && fp.PingSlot.Frequency != 0 {
		macState.DesiredParameters.PingSlotFrequency = fp.PingSlot.Frequency
	} else if defaults.DesiredPingSlotFrequency != nil && defaults.DesiredPingSlotFrequency.Value != 0 {
		macState.DesiredParameters.PingSlotFrequency = defaults.DesiredPingSlotFrequency.Value
	} else {
		macState.DesiredParameters.PingSlotFrequency = macState.CurrentParameters.PingSlotFrequency
	}

	if dev.GetMACSettings().GetPingSlotDataRateIndex() != nil {
		macState.CurrentParameters.PingSlotDataRateIndexValue = dev.MACSettings.PingSlotDataRateIndex
	} else if defaults.PingSlotDataRateIndex != nil {
		macState.CurrentParameters.PingSlotDataRateIndexValue = defaults.PingSlotDataRateIndex
	} else {
		// Default to mbed-os and LoRaMac-node behavior.
		// https://github.com/ARMmbed/mbed-os/blob/131ea2bb243eef898a501576e611ebbf504b079a/features/lorawan/lorastack/phy/LoRaPHY.cpp#L1625-L1630
		// https://github.com/Lora-net/LoRaMac-node/blob/87f19e84ae2fc4af72af9567fe722386de6ce9f4/src/mac/region/RegionCN779.h#L235.
		macState.CurrentParameters.PingSlotDataRateIndexValue = &ttnpb.DataRateIndexValue{Value: phy.Beacon.DataRateIndex}
	}
	if dev.GetMACSettings().GetDesiredPingSlotDataRateIndex() != nil {
		macState.DesiredParameters.PingSlotDataRateIndexValue = dev.MACSettings.DesiredPingSlotDataRateIndex
	} else if fp.DefaultPingSlotDataRate != nil {
		macState.DesiredParameters.PingSlotDataRateIndexValue = &ttnpb.DataRateIndexValue{Value: ttnpb.DataRateIndex(*fp.DefaultPingSlotDataRate)}
	} else if defaults.DesiredPingSlotDataRateIndex != nil {
		macState.DesiredParameters.PingSlotDataRateIndexValue = defaults.DesiredPingSlotDataRateIndex
	} else {
		macState.DesiredParameters.PingSlotDataRateIndexValue = macState.CurrentParameters.PingSlotDataRateIndexValue
	}

	if dev.GetMACSettings().GetBeaconFrequency() != nil && dev.MACSettings.BeaconFrequency.Value != 0 {
		macState.CurrentParameters.BeaconFrequency = dev.MACSettings.BeaconFrequency.Value
	} else if defaults.BeaconFrequency != nil {
		macState.CurrentParameters.BeaconFrequency = defaults.BeaconFrequency.Value
	}
	if dev.GetMACSettings().GetDesiredBeaconFrequency() != nil && dev.MACSettings.DesiredBeaconFrequency.Value != 0 {
		macState.DesiredParameters.BeaconFrequency = dev.MACSettings.DesiredBeaconFrequency.Value
	} else if defaults.DesiredBeaconFrequency != nil && defaults.DesiredBeaconFrequency.Value != 0 {
		macState.DesiredParameters.BeaconFrequency = defaults.DesiredBeaconFrequency.Value
	} else {
		macState.DesiredParameters.BeaconFrequency = macState.CurrentParameters.BeaconFrequency
	}

	if len(phy.DownlinkChannels) > len(phy.UplinkChannels) || len(fp.DownlinkChannels) > len(fp.UplinkChannels) ||
		len(phy.UplinkChannels) > int(phy.MaxUplinkChannels) || len(phy.DownlinkChannels) > int(phy.MaxDownlinkChannels) ||
		len(fp.UplinkChannels) > int(phy.MaxUplinkChannels) || len(fp.DownlinkChannels) > int(phy.MaxDownlinkChannels) {
		// NOTE: In case the spec changes and this assumption is not valid anymore,
		// the implementation of this function won't be valid and has to be changed.
		panic("uplink/downlink channel length is inconsistent")
	}

	// NOTE: FactoryPresetFrequencies does not indicate the data rate ranges allowed for channels.
	// In the latest regional parameters spec(1.1b) the data rate ranges are DR0-DR5 for mandatory channels in all non-fixed channel plans,
	// hence we assume the same range for predefined channels.
	if len(dev.GetMACSettings().GetFactoryPresetFrequencies()) > 0 {
		macState.CurrentParameters.Channels = make([]*ttnpb.MACParameters_Channel, 0, len(dev.MACSettings.FactoryPresetFrequencies))
		for _, freq := range dev.MACSettings.FactoryPresetFrequencies {
			macState.CurrentParameters.Channels = append(macState.CurrentParameters.Channels, &ttnpb.MACParameters_Channel{
				MinDataRateIndex:  0,
				MaxDataRateIndex:  ttnpb.DATA_RATE_5,
				UplinkFrequency:   freq,
				DownlinkFrequency: freq,
				EnableUplink:      true,
			})
		}
	} else if len(defaults.GetFactoryPresetFrequencies()) > 0 {
		macState.CurrentParameters.Channels = make([]*ttnpb.MACParameters_Channel, 0, len(defaults.FactoryPresetFrequencies))
		for _, freq := range defaults.FactoryPresetFrequencies {
			macState.CurrentParameters.Channels = append(macState.CurrentParameters.Channels, &ttnpb.MACParameters_Channel{
				MinDataRateIndex:  0,
				MaxDataRateIndex:  ttnpb.DATA_RATE_5,
				UplinkFrequency:   freq,
				DownlinkFrequency: freq,
				EnableUplink:      true,
			})
		}
	} else {
		macState.CurrentParameters.Channels = make([]*ttnpb.MACParameters_Channel, 0, len(phy.UplinkChannels))
		for i, phyUpCh := range phy.UplinkChannels {
			macState.CurrentParameters.Channels = append(macState.CurrentParameters.Channels, &ttnpb.MACParameters_Channel{
				MinDataRateIndex:  phyUpCh.MinDataRate,
				MaxDataRateIndex:  phyUpCh.MaxDataRate,
				UplinkFrequency:   phyUpCh.Frequency,
				DownlinkFrequency: phy.DownlinkChannels[i%len(phy.DownlinkChannels)].Frequency,
				EnableUplink:      true,
			})
		}
	}
	macState.DesiredParameters.Channels = frequencyPlanChannels(phy, fp.UplinkChannels, fp.DownlinkChannels...)
	return macState, nil
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
