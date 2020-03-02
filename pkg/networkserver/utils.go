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
	"bytes"
	"context"
	"fmt"
	"math"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/gpstime"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// nsScheduleWindow returns minimum time.Duration between downlink being added to the queue and it being sent to GS for transmission.
func nsScheduleWindow() time.Duration {
	// TODO: Observe this value at runtime https://github.com/TheThingsNetwork/lorawan-stack/issues/1552.
	return 200 * time.Millisecond
}

var timeNow func() time.Time = time.Now

func timeUntil(t time.Time) time.Duration {
	return t.Sub(timeNow())
}

func timeSince(t time.Time) time.Duration {
	return timeNow().Sub(t)
}

// copyEndDevice returns a deep copy of ttnpb.EndDevice pb.
func copyEndDevice(pb *ttnpb.EndDevice) *ttnpb.EndDevice {
	return deepcopy.Copy(pb).(*ttnpb.EndDevice)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func deviceUseADR(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) bool {
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
const DefaultClassBTimeout = time.Minute

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
const DefaultClassCTimeout = 15 * time.Second

func deviceClassCTimeout(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) time.Duration {
	if dev.MACSettings != nil && dev.MACSettings.ClassCTimeout != nil {
		return *dev.MACSettings.ClassCTimeout
	}
	if defaults.ClassCTimeout != nil {
		return *defaults.ClassCTimeout
	}
	return DefaultClassCTimeout
}

func getDeviceBandVersion(dev *ttnpb.EndDevice, fps *frequencyplans.Store) (*frequencyplans.FrequencyPlan, band.Band, error) {
	fp, err := fps.GetByID(dev.FrequencyPlanID)
	if err != nil {
		return nil, band.Band{}, err
	}
	b, err := band.GetByID(fp.BandID)
	if err != nil {
		return nil, band.Band{}, err
	}
	b, err = b.Version(dev.LoRaWANPHYVersion)
	if err != nil {
		return nil, band.Band{}, err
	}
	return fp, b, nil
}

func searchUplinkChannel(freq uint64, macState *ttnpb.MACState) (uint8, error) {
	for i, ch := range macState.CurrentParameters.Channels {
		if ch.UplinkFrequency == freq {
			return uint8(i), nil
		}
	}
	return 0, errUplinkChannelNotFound.WithAttributes("frequency", freq)
}

func partitionDownlinks(p func(down *ttnpb.ApplicationDownlink) bool, downs ...*ttnpb.ApplicationDownlink) (t, f []*ttnpb.ApplicationDownlink) {
	t, f = downs[:0:0], downs[:0:0]
	for _, down := range downs {
		if p(down) {
			t = append(t, down)
		} else {
			f = append(f, down)
		}
	}
	return t, f
}

func paritionDownlinksBySessionKeyID(p func([]byte) bool, downs ...*ttnpb.ApplicationDownlink) (t, f []*ttnpb.ApplicationDownlink) {
	return partitionDownlinks(func(down *ttnpb.ApplicationDownlink) bool { return p(down.SessionKeyID) }, downs...)
}

func partitionDownlinksBySessionKeyIDEquality(id []byte, downs ...*ttnpb.ApplicationDownlink) (t, f []*ttnpb.ApplicationDownlink) {
	return paritionDownlinksBySessionKeyID(func(downID []byte) bool { return bytes.Equal(downID, id) }, downs...)
}

func deviceNeedsMACRequestsAt(ctx context.Context, dev *ttnpb.EndDevice, t time.Time, phy band.Band, defaults ttnpb.MACSettings) bool {
	switch {
	case deviceNeedsADRParamSetupReq(dev, phy),
		deviceNeedsBeaconFreqReq(dev),
		deviceNeedsBeaconTimingReq(dev),
		deviceNeedsDevStatusReq(dev, defaults, t),
		deviceNeedsDLChannelReq(dev),
		deviceNeedsDutyCycleReq(dev),
		deviceNeedsLinkADRReq(dev),
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

func needsClassADataDownlinkAt(ctx context.Context, dev *ttnpb.EndDevice, t time.Time, phy band.Band, defaults ttnpb.MACSettings) bool {
	if dev.Session == nil || dev.MACState == nil || len(dev.MACState.RecentUplinks) == 0 {
		return false
	}
	if len(dev.MACState.QueuedResponses) > 0 {
		return true
	}
	up := lastUplink(dev.MACState.RecentUplinks...)
	switch up.Payload.MHDR.MType {
	case ttnpb.MType_UNCONFIRMED_UP:
		if up.Payload.GetMACPayload().FCtrl.ADRAckReq {
			return true
		}
	case ttnpb.MType_CONFIRMED_UP:
		return true
	}
	for _, down := range dev.Session.QueuedApplicationDownlinks {
		if down.GetClassBC() == nil {
			return true
		}
	}
	return deviceNeedsMACRequestsAt(ctx, dev, t, phy, defaults)
}

func classAWindows(up *ttnpb.UplinkMessage, rxDelay time.Duration) (rx1, rx2 time.Time) {
	rx1 = up.ReceivedAt.Add(-infrastructureDelay/2 + rxDelay)
	return rx1, rx1.Add(time.Second)
}

func nextClassADownlinkAt(dev *ttnpb.EndDevice, earliestAt time.Time) (time.Time, bool) {
	if dev.MACState == nil || !dev.MACState.RxWindowsAvailable || len(dev.MACState.RecentUplinks) == 0 {
		return time.Time{}, false
	}
	rx1, rx2 := classAWindows(lastUplink(dev.MACState.RecentUplinks...), dev.MACState.CurrentParameters.Rx1Delay.Duration())
	switch {
	case !earliestAt.After(rx1):
		return rx1, true
	case !earliestAt.After(rx2):
		return rx2, true
	default:
		return time.Time{}, false
	}
}

func nextUnconfirmedNetworkInitiatedDownlinkAt(ctx context.Context, dev *ttnpb.EndDevice, earliestAt time.Time) (time.Time, bool) {
	if dev.MACState == nil {
		log.FromContext(ctx).Warn("Insufficient data to compute next network-initiated downlink slot")
		return time.Time{}, false
	}
	if dev.MACState.LastNetworkInitiatedDownlinkAt == nil {
		return earliestAt, true
	}
	if len(dev.MACState.RecentUplinks) > 0 {
		recvAt := lastUplink(dev.MACState.RecentUplinks...).ReceivedAt
		if recvAt.After(*dev.MACState.LastNetworkInitiatedDownlinkAt) {
			if earliestAt.After(recvAt) {
				return earliestAt, true
			}
			return recvAt, true
		}
	}
	t := dev.MACState.LastNetworkInitiatedDownlinkAt.Add(networkInitiatedDownlinkInterval)
	if earliestAt.After(t) {
		return earliestAt, true
	}
	return t, true
}

func nextConfirmedNetworkInitiatedDownlinkAt(ctx context.Context, dev *ttnpb.EndDevice, timeout time.Duration, earliestAt time.Time) (time.Time, bool) {
	if dev.MACState == nil {
		log.FromContext(ctx).Warn("Insufficient data to compute next confirmed network-initiated downlink slot")
		return time.Time{}, false
	}
	earliestAt, ok := nextUnconfirmedNetworkInitiatedDownlinkAt(ctx, dev, earliestAt)
	if !ok {
		return time.Time{}, false
	}
	if dev.MACState.LastConfirmedDownlinkAt == nil {
		return earliestAt, true
	}
	if len(dev.MACState.RecentUplinks) > 0 {
		recvAt := lastUplink(dev.MACState.RecentUplinks...).ReceivedAt
		if recvAt.After(*dev.MACState.LastConfirmedDownlinkAt) {
			if earliestAt.After(recvAt) {
				return earliestAt, true
			}
			return recvAt, true
		}
	}
	t := dev.MACState.LastConfirmedDownlinkAt.Add(timeout)
	if earliestAt.After(t) {
		return earliestAt, true
	}
	return t, true
}

const (
	tBeaconDelay   = 1*time.Microsecond + 500*time.Nanosecond
	beaconPeriod   = 128 * time.Second
	beaconReserved = 2*time.Second + 120*time.Millisecond
	pingSlotCount  = 4096
	pingSlotLen    = 30 * time.Millisecond
)

func beaconTimeBefore(t time.Time) time.Duration {
	return gpstime.ToGPS(t) / beaconPeriod * beaconPeriod
}

// nextPingSlotAt returns the transmission time of next available class B ping slot, which will always be after earliestAt.
func nextPingSlotAt(ctx context.Context, dev *ttnpb.EndDevice, earliestAt time.Time) (time.Time, bool) {
	if dev.Session == nil || dev.Session.DevAddr.IsZero() || dev.MACState == nil || dev.MACState.PingSlotPeriodicity == nil {
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

func nextUnconfirmedClassBDownlinkAt(ctx context.Context, dev *ttnpb.EndDevice, earliestAt time.Time) (time.Time, bool) {
	earliestAt, ok := nextUnconfirmedNetworkInitiatedDownlinkAt(ctx, dev, earliestAt)
	if !ok {
		return time.Time{}, false
	}
	return nextPingSlotAt(ctx, dev, earliestAt)
}

func nextConfirmedClassBDownlinkAt(ctx context.Context, dev *ttnpb.EndDevice, defaults ttnpb.MACSettings, earliestAt time.Time) (time.Time, bool) {
	earliestAt, ok := nextConfirmedNetworkInitiatedDownlinkAt(ctx, dev, deviceClassBTimeout(dev, defaults), earliestAt)
	if !ok {
		return time.Time{}, false
	}
	return nextPingSlotAt(ctx, dev, earliestAt)
}

func nextUnconfirmedClassCDownlinkAt(ctx context.Context, dev *ttnpb.EndDevice, earliestAt time.Time) (time.Time, bool) {
	return nextUnconfirmedNetworkInitiatedDownlinkAt(ctx, dev, earliestAt)
}

func nextConfirmedClassCDownlinkAt(ctx context.Context, dev *ttnpb.EndDevice, defaults ttnpb.MACSettings, earliestAt time.Time) (time.Time, bool) {
	return nextConfirmedNetworkInitiatedDownlinkAt(ctx, dev, deviceClassCTimeout(dev, defaults), earliestAt)
}

// nextDataDownlinkAt returns the time.Time after earliestAt when a data downlink
// should be transmitted to the device by the gateway and whether or not there is a data downlink to schedule.
func nextDataDownlinkAt(ctx context.Context, dev *ttnpb.EndDevice, phy band.Band, defaults ttnpb.MACSettings, earliestAt time.Time) (time.Time, ttnpb.Class, bool) {
	if dev.Session == nil || dev.MACState == nil {
		return time.Time{}, ttnpb.CLASS_A, false
	}
	if !dev.Multicast {
		downAt, ok := nextClassADownlinkAt(dev, earliestAt)
		if ok && needsClassADataDownlinkAt(ctx, dev, downAt, phy, defaults) {
			return downAt, ttnpb.CLASS_A, true
		}
	}

	class := dev.MACState.DeviceClass
	if class == ttnpb.CLASS_A {
		return time.Time{}, class, false
	}

	var earliestConfirmedAt time.Time
	switch class {
	case ttnpb.CLASS_B:
		var ok bool
		earliestAt, ok = nextUnconfirmedClassBDownlinkAt(ctx, dev, earliestAt)
		if !ok {
			return time.Time{}, class, false
		}
		earliestConfirmedAt, ok = nextConfirmedClassBDownlinkAt(ctx, dev, defaults, earliestAt)
		if !ok {
			return time.Time{}, class, false
		}

	case ttnpb.CLASS_C:
		var ok bool
		earliestAt, ok = nextUnconfirmedClassCDownlinkAt(ctx, dev, earliestAt)
		if !ok {
			return time.Time{}, class, false
		}
		earliestConfirmedAt, ok = nextConfirmedClassCDownlinkAt(ctx, dev, defaults, earliestAt)
		if !ok {
			return time.Time{}, class, false
		}

	default:
		panic(fmt.Sprintf("unmatched device class: %v", dev.MACState.DeviceClass))
	}

	var absTime time.Time
	for _, down := range dev.Session.QueuedApplicationDownlinks {
		if down.ClassBC == nil || down.ClassBC.AbsoluteTime == nil || down.ClassBC.AbsoluteTime.IsZero() {
			if down.Confirmed || deviceNeedsMACRequestsAt(ctx, dev, earliestAt, phy, defaults) {
				return earliestConfirmedAt, class, true
			}
			return earliestAt, class, true
		}
		t := *down.ClassBC.AbsoluteTime
		if t.Before(earliestAt) || t.Before(earliestConfirmedAt) && (down.Confirmed || deviceNeedsMACRequestsAt(ctx, dev, t, phy, defaults)) {
			// This downlink will never be scheduled, hence continue to next one.
			continue
		}
		if t.After(earliestConfirmedAt) {
			// NOTE: There may be MAC commands available to send earlier than t.
			absTime = t.UTC()
			break
		}
		// t is in (earliestAt;earliestConfirmedAt] range.
		return t.UTC(), class, true
	}
	if deviceNeedsMACRequestsAt(ctx, dev, earliestConfirmedAt, phy, defaults) {
		return earliestConfirmedAt, class, true
	}

	statusAt, ok := deviceNeedsDevStatusReqAt(dev, defaults)
	if !ok {
		if absTime.IsZero() {
			return time.Time{}, class, false
		}
		return absTime, class, true
	}

	// NOTE: statusAt is after earliestConfirmedAt, otherwise deviceNeedsMACRequestsAt call above would evaluate to true.
	if !absTime.IsZero() && statusAt.After(absTime) {
		return absTime, class, true
	}
	if class != ttnpb.CLASS_B {
		return statusAt, class, true
	}

	t, ok := nextConfirmedClassBDownlinkAt(ctx, dev, defaults, statusAt)
	if !ok {
		if absTime.IsZero() {
			return time.Time{}, class, false
		}
		return absTime, class, true
	}
	if !absTime.IsZero() && absTime.Before(t) {
		return absTime, class, true
	}
	return t, class, true
}

func newMACState(dev *ttnpb.EndDevice, fps *frequencyplans.Store, defaults ttnpb.MACSettings) (*ttnpb.MACState, error) {
	fp, phy, err := getDeviceBandVersion(dev, fps)
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
			return nil, errClassAMulticast
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

	if phy.TxParamSetupReqSupport {
		macState.DesiredParameters.UplinkDwellTime = &pbtypes.BoolValue{Value: fp.DwellTime.GetUplinks()}
	}

	if phy.TxParamSetupReqSupport {
		macState.DesiredParameters.DownlinkDwellTime = &pbtypes.BoolValue{Value: fp.DwellTime.GetDownlinks()}
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
		macState.CurrentParameters.PingSlotDataRateIndexValue = &ttnpb.DataRateIndexValue{Value: ttnpb.DataRateIndex(phy.Beacon.DataRateIndex)}
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
		for i, upCh := range phy.UplinkChannels {
			channel := &ttnpb.MACParameters_Channel{
				MinDataRateIndex: upCh.MinDataRate,
				MaxDataRateIndex: upCh.MaxDataRate,
				UplinkFrequency:  upCh.Frequency,
				EnableUplink:     true,
			}
			channel.DownlinkFrequency = phy.DownlinkChannels[i%len(phy.DownlinkChannels)].Frequency
			macState.CurrentParameters.Channels = append(macState.CurrentParameters.Channels, channel)
		}
	}

	macState.DesiredParameters.Channels = make([]*ttnpb.MACParameters_Channel, 0, len(phy.UplinkChannels)+len(fp.UplinkChannels))
	for i, upCh := range phy.UplinkChannels {
		channel := &ttnpb.MACParameters_Channel{
			MinDataRateIndex: upCh.MinDataRate,
			MaxDataRateIndex: upCh.MaxDataRate,
			UplinkFrequency:  upCh.Frequency,
		}
		channel.DownlinkFrequency = phy.DownlinkChannels[i%len(phy.DownlinkChannels)].Frequency
		macState.DesiredParameters.Channels = append(macState.DesiredParameters.Channels, channel)
	}

outerUp:
	for _, upCh := range fp.UplinkChannels {
		for _, ch := range macState.DesiredParameters.Channels {
			if ch.UplinkFrequency == upCh.Frequency {
				ch.MinDataRateIndex = ttnpb.DataRateIndex(upCh.MinDataRate)
				ch.MaxDataRateIndex = ttnpb.DataRateIndex(upCh.MaxDataRate)
				ch.EnableUplink = true
				continue outerUp
			}
		}

		macState.DesiredParameters.Channels = append(macState.DesiredParameters.Channels, &ttnpb.MACParameters_Channel{
			MinDataRateIndex: ttnpb.DataRateIndex(upCh.MinDataRate),
			MaxDataRateIndex: ttnpb.DataRateIndex(upCh.MaxDataRate),
			UplinkFrequency:  upCh.Frequency,
			EnableUplink:     true,
		})
	}

	if len(fp.DownlinkChannels) > 0 {
		for i, ch := range macState.DesiredParameters.Channels {
			downCh := fp.DownlinkChannels[i%len(fp.DownlinkChannels)]
			if downCh.Frequency != 0 {
				ch.DownlinkFrequency = downCh.Frequency
			}
		}
	}

	return macState, nil
}
