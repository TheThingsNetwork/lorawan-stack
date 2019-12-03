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
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

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

func searchDataRate(dr ttnpb.DataRate, dev *ttnpb.EndDevice, fps *frequencyplans.Store) (ttnpb.DataRateIndex, error) {
	_, phy, err := getDeviceBandVersion(dev, fps)
	if err != nil {
		return 0, err
	}
	for i, bDR := range phy.DataRates {
		if bDR.Rate.Equal(dr) {
			return ttnpb.DataRateIndex(i), nil
		}
	}
	return 0, errDataRateNotFound.WithAttributes("data_rate", dr)
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
	statusAt, ok := deviceNeedsDevStatusReqAt(dev, defaults)
	return ok && t.After(statusAt)
}

func lastUplink(ups ...*ttnpb.UplinkMessage) *ttnpb.UplinkMessage {
	return ups[len(ups)-1]
}

func needsClassADataDownlinkAt(ctx context.Context, dev *ttnpb.EndDevice, t time.Time, phy band.Band, defaults ttnpb.MACSettings) bool {
	if dev.MACState == nil || !dev.MACState.RxWindowsAvailable || len(dev.MACState.RecentUplinks) == 0 {
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
	for _, down := range dev.QueuedApplicationDownlinks {
		if down.GetClassBC() == nil {
			return true
		}
	}
	return deviceNeedsMACRequestsAt(ctx, dev, t, phy, defaults)
}

func nextClassADataDownlinkSlot(dev *ttnpb.EndDevice) (time.Time, bool) {
	if dev.MACState == nil || !dev.MACState.RxWindowsAvailable || len(dev.MACState.RecentUplinks) == 0 {
		return time.Time{}, false
	}
	rx1 := lastUplink(dev.MACState.RecentUplinks...).ReceivedAt.Add(dev.MACState.CurrentParameters.Rx1Delay.Duration())
	rx2 := rx1.Add(time.Second)
	switch {
	case rx2.Before(timeNow()):
		return time.Time{}, false
	case rx1.After(timeNow()):
		return rx1, true
	default:
		return rx2, true
	}
}

func nextConfirmedClassCDownlinkAt(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) time.Time {
	if dev.GetMACState().GetLastConfirmedDownlinkAt() == nil {
		return time.Time{}
	}
	if dev.GetMACState().GetRxWindowsAvailable() {
		return time.Time{}
	}
	if len(dev.MACState.RecentUplinks) > 0 {
		if lastUplink(dev.MACState.RecentUplinks...).ReceivedAt.After(*dev.MACState.LastConfirmedDownlinkAt) {
			return time.Time{}
		}
	}
	return dev.MACState.LastConfirmedDownlinkAt.Add(deviceClassCTimeout(dev, defaults))
}

// nextDataDownlinkAt returns the time.Time when the downlink should be scheduled on the Gateway Server
// and whether or not there is a data downlink to schedule.
func nextDataDownlinkAt(ctx context.Context, dev *ttnpb.EndDevice, phy band.Band, defaults ttnpb.MACSettings) (time.Time, bool) {
	if dev.MACState == nil {
		return time.Time{}, false
	}
	if !dev.Multicast {
		downAt, ok := nextClassADataDownlinkSlot(dev)
		if ok && needsClassADataDownlinkAt(ctx, dev, downAt, phy, defaults) {
			return downAt.Add(-gsScheduleWindow), true
		}
	}
	switch dev.MACState.DeviceClass {
	case ttnpb.CLASS_A, ttnpb.CLASS_B:
		return time.Time{}, false
	case ttnpb.CLASS_C:
		now := timeNow().UTC()
		confirmedAt := nextConfirmedClassCDownlinkAt(dev, defaults)
		if confirmedAt.Before(now) {
			confirmedAt = now
		}
		if deviceNeedsMACRequestsAt(ctx, dev, confirmedAt, phy, defaults) {
			return confirmedAt, true
		}
		var absTime time.Time
		if len(dev.QueuedApplicationDownlinks) > 0 {
			classBC := dev.QueuedApplicationDownlinks[0].ClassBC
			if classBC == nil || classBC.AbsoluteTime == nil || classBC.AbsoluteTime.IsZero() {
				return now, true
			}
			absTime = classBC.AbsoluteTime.UTC()
		}

		statusAt, ok := deviceNeedsDevStatusReqAt(dev, defaults)
		if !ok {
			if absTime.IsZero() {
				return time.Time{}, false
			}
			return absTime.Add(-gsScheduleWindow), true
		}
		// NOTE: statusAt is after confirmedAt, otherwise deviceNeedsMACRequestsAt call above would evaluate to true.
		if statusAt.After(absTime) {
			return absTime.Add(-gsScheduleWindow), true
		}
		return statusAt, true
	default:
		panic(fmt.Sprintf("Unmatched device class: %v", dev.MACState.DeviceClass))
	}
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

	macState.CurrentParameters.MaxEIRP = phy.DefaultMaxEIRP
	macState.DesiredParameters.MaxEIRP = macState.CurrentParameters.MaxEIRP
	if fp.MaxEIRP != nil && *fp.MaxEIRP > 0 && *fp.MaxEIRP < macState.CurrentParameters.MaxEIRP {
		macState.DesiredParameters.MaxEIRP = *fp.MaxEIRP
	}

	macState.CurrentParameters.UplinkDwellTime = nil
	if phy.TxParamSetupReqSupport {
		macState.DesiredParameters.UplinkDwellTime = &pbtypes.BoolValue{Value: fp.DwellTime.GetUplinks()}
	}

	macState.CurrentParameters.DownlinkDwellTime = nil
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
	macState.DesiredParameters.ADRAckLimitExponent = &ttnpb.ADRAckLimitExponentValue{Value: phy.ADRAckLimit}
	if dev.GetMACSettings().GetDesiredADRAckLimitExponent() != nil {
		macState.DesiredParameters.ADRAckLimitExponent = &ttnpb.ADRAckLimitExponentValue{Value: dev.MACSettings.DesiredADRAckLimitExponent.Value}
	} else if defaults.DesiredADRAckLimitExponent != nil {
		macState.DesiredParameters.ADRAckLimitExponent = &ttnpb.ADRAckLimitExponentValue{Value: defaults.DesiredADRAckLimitExponent.Value}
	}

	macState.CurrentParameters.ADRAckDelayExponent = &ttnpb.ADRAckDelayExponentValue{Value: phy.ADRAckDelay}
	macState.DesiredParameters.ADRAckDelayExponent = &ttnpb.ADRAckDelayExponentValue{Value: phy.ADRAckDelay}
	if dev.GetMACSettings().GetDesiredADRAckDelayExponent() != nil {
		macState.DesiredParameters.ADRAckDelayExponent = &ttnpb.ADRAckDelayExponentValue{Value: dev.MACSettings.DesiredADRAckDelayExponent.Value}
	} else if defaults.DesiredADRAckDelayExponent != nil {
		macState.DesiredParameters.ADRAckDelayExponent = &ttnpb.ADRAckDelayExponentValue{Value: defaults.DesiredADRAckDelayExponent.Value}
	}

	macState.CurrentParameters.Rx1Delay = ttnpb.RxDelay(phy.ReceiveDelay1.Seconds())
	if dev.GetMACSettings().GetRx1Delay() != nil {
		macState.CurrentParameters.Rx1Delay = dev.MACSettings.Rx1Delay.Value
	} else if defaults.Rx1Delay != nil {
		macState.CurrentParameters.Rx1Delay = defaults.Rx1Delay.Value
	}
	macState.DesiredParameters.Rx1Delay = macState.CurrentParameters.Rx1Delay
	if dev.GetMACSettings().GetDesiredRx1Delay() != nil {
		macState.DesiredParameters.Rx1Delay = dev.MACSettings.DesiredRx1Delay.Value
	} else if defaults.DesiredRx1Delay != nil {
		macState.DesiredParameters.Rx1Delay = defaults.DesiredRx1Delay.Value
	}

	macState.CurrentParameters.Rx1DataRateOffset = 0
	if dev.GetMACSettings().GetRx1DataRateOffset() != nil {
		macState.CurrentParameters.Rx1DataRateOffset = dev.MACSettings.Rx1DataRateOffset.Value
	} else if defaults.Rx1DataRateOffset != nil {
		macState.CurrentParameters.Rx1DataRateOffset = defaults.Rx1DataRateOffset.Value
	}
	macState.DesiredParameters.Rx1DataRateOffset = macState.CurrentParameters.Rx1DataRateOffset
	if dev.GetMACSettings().GetDesiredRx1DataRateOffset() != nil {
		macState.DesiredParameters.Rx1DataRateOffset = dev.MACSettings.DesiredRx1DataRateOffset.Value
	} else if defaults.DesiredRx1DataRateOffset != nil {
		macState.DesiredParameters.Rx1DataRateOffset = defaults.DesiredRx1DataRateOffset.Value
	}

	macState.CurrentParameters.Rx2DataRateIndex = phy.DefaultRx2Parameters.DataRateIndex
	if dev.GetMACSettings().GetRx2DataRateIndex() != nil {
		macState.CurrentParameters.Rx2DataRateIndex = dev.MACSettings.Rx2DataRateIndex.Value
	} else if defaults.Rx2DataRateIndex != nil {
		macState.CurrentParameters.Rx2DataRateIndex = defaults.Rx2DataRateIndex.Value
	}
	macState.DesiredParameters.Rx2DataRateIndex = macState.CurrentParameters.Rx2DataRateIndex
	if dev.GetMACSettings().GetDesiredRx2DataRateIndex() != nil {
		macState.DesiredParameters.Rx2DataRateIndex = dev.MACSettings.DesiredRx2DataRateIndex.Value
	} else if fp.DefaultRx2DataRate != nil {
		macState.DesiredParameters.Rx2DataRateIndex = ttnpb.DataRateIndex(*fp.DefaultRx2DataRate)
	} else if defaults.DesiredRx2DataRateIndex != nil {
		macState.DesiredParameters.Rx2DataRateIndex = defaults.DesiredRx2DataRateIndex.Value
	}

	macState.CurrentParameters.Rx2Frequency = phy.DefaultRx2Parameters.Frequency
	if dev.GetMACSettings().GetRx2Frequency() != nil && dev.MACSettings.Rx2Frequency.Value != 0 {
		macState.CurrentParameters.Rx2Frequency = dev.MACSettings.Rx2Frequency.Value
	} else if defaults.Rx2Frequency != nil && dev.MACSettings.Rx2Frequency.Value != 0 {
		macState.CurrentParameters.Rx2Frequency = defaults.Rx2Frequency.Value
	}
	macState.DesiredParameters.Rx2Frequency = macState.CurrentParameters.Rx2Frequency
	if dev.GetMACSettings().GetDesiredRx2Frequency() != nil && dev.MACSettings.DesiredRx2Frequency.Value != 0 {
		macState.DesiredParameters.Rx2Frequency = dev.MACSettings.DesiredRx2Frequency.Value
	} else if fp.Rx2Channel != nil {
		macState.DesiredParameters.Rx2Frequency = fp.Rx2Channel.Frequency
	} else if defaults.DesiredRx2Frequency != nil && defaults.DesiredRx2Frequency.Value != 0 {
		macState.DesiredParameters.Rx2Frequency = defaults.DesiredRx2Frequency.Value
	}

	macState.CurrentParameters.MaxDutyCycle = ttnpb.DUTY_CYCLE_1
	if dev.GetMACSettings().GetMaxDutyCycle() != nil {
		macState.CurrentParameters.MaxDutyCycle = dev.MACSettings.MaxDutyCycle.Value
	}
	macState.DesiredParameters.MaxDutyCycle = macState.CurrentParameters.MaxDutyCycle
	if dev.GetMACSettings().GetDesiredMaxDutyCycle() != nil {
		macState.DesiredParameters.MaxDutyCycle = dev.MACSettings.DesiredMaxDutyCycle.Value
	} else if defaults.DesiredMaxDutyCycle != nil {
		macState.DesiredParameters.MaxDutyCycle = defaults.DesiredMaxDutyCycle.Value
	}

	// TODO: Support rejoins. (https://github.com/TheThingsNetwork/lorawan-stack/issues/8)
	macState.CurrentParameters.RejoinTimePeriodicity = ttnpb.REJOIN_TIME_0
	macState.DesiredParameters.RejoinTimePeriodicity = macState.CurrentParameters.RejoinTimePeriodicity

	macState.CurrentParameters.RejoinCountPeriodicity = ttnpb.REJOIN_COUNT_16
	macState.DesiredParameters.RejoinCountPeriodicity = macState.CurrentParameters.RejoinCountPeriodicity

	macState.CurrentParameters.PingSlotFrequency = 0
	if dev.GetMACSettings().GetPingSlotFrequency() != nil && dev.MACSettings.PingSlotFrequency.Value != 0 {
		macState.CurrentParameters.PingSlotFrequency = dev.MACSettings.PingSlotFrequency.Value
	} else if defaults.PingSlotFrequency != nil && defaults.PingSlotFrequency.Value != 0 {
		macState.CurrentParameters.PingSlotFrequency = defaults.PingSlotFrequency.Value
	}
	macState.DesiredParameters.PingSlotFrequency = macState.CurrentParameters.PingSlotFrequency
	if fp.PingSlot != nil && fp.PingSlot.Frequency != 0 {
		macState.DesiredParameters.PingSlotFrequency = fp.PingSlot.Frequency
	}

	macState.CurrentParameters.PingSlotDataRateIndex = ttnpb.DATA_RATE_0
	if dev.GetMACSettings().GetPingSlotDataRateIndex() != nil {
		macState.CurrentParameters.PingSlotDataRateIndex = dev.MACSettings.PingSlotDataRateIndex.Value
	} else if defaults.PingSlotDataRateIndex != nil {
		macState.CurrentParameters.PingSlotDataRateIndex = defaults.PingSlotDataRateIndex.Value
	}
	macState.DesiredParameters.PingSlotDataRateIndex = macState.CurrentParameters.PingSlotDataRateIndex
	if fp.DefaultPingSlotDataRate != nil {
		macState.DesiredParameters.PingSlotDataRateIndex = ttnpb.DataRateIndex(*fp.DefaultPingSlotDataRate)
	}

	// TODO: Support class B. (https://github.com/TheThingsNetwork/lorawan-stack/issues/19)
	macState.CurrentParameters.BeaconFrequency = 0
	macState.DesiredParameters.BeaconFrequency = macState.CurrentParameters.BeaconFrequency

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
