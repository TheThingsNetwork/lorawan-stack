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

package mac

import (
	"context"
	"fmt"
	"math"

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/gpstime"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/specification/macspec"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"golang.org/x/exp/slices"
)

func channelDataRateRange(chs ...*ttnpb.MACParameters_Channel) (min, max ttnpb.DataRateIndex, ok bool) {
	i := 0
	for {
		if i >= len(chs) {
			return 0, 0, false
		}
		if chs[i].GetEnableUplink() {
			break
		}
		i++
	}

	min = chs[i].MinDataRateIndex
	max = chs[i].MaxDataRateIndex
	for _, ch := range chs[i+1:] {
		if !ch.GetEnableUplink() {
			continue
		}
		if ch.MaxDataRateIndex > max {
			max = ch.MaxDataRateIndex
		}
		if ch.MinDataRateIndex < min {
			min = ch.MinDataRateIndex
		}
	}
	if min > max {
		return 0, 0, false
	}
	return min, max, true
}

// DefaultClassBTimeout is the default time-out for the device to respond to class B downlink messages.
// When waiting for a response times out, the downlink message is considered lost, and the downlink task triggers again.
const DefaultClassBTimeout = 10 * time.Minute

func DeviceClassBTimeout(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings) time.Duration {
	if t := dev.GetMacSettings().GetClassBTimeout(); t != nil {
		return ttnpb.StdDurationOrZero(t)
	}
	if defaults.ClassBTimeout != nil {
		return ttnpb.StdDurationOrZero(defaults.ClassBTimeout)
	}
	return DefaultClassBTimeout
}

// DefaultClassCTimeout is the default time-out for the device to respond to class C downlink messages.
// When waiting for a response times out, the downlink message is considered lost, and the downlink task triggers again.
const DefaultClassCTimeout = 5 * time.Minute

func DeviceClassCTimeout(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings) time.Duration {
	if t := dev.GetMacSettings().GetClassCTimeout(); t != nil {
		return ttnpb.StdDurationOrZero(t)
	}
	if defaults.ClassCTimeout != nil {
		return ttnpb.StdDurationOrZero(defaults.ClassCTimeout)
	}
	return DefaultClassCTimeout
}

const (
	tBeaconDelay   = 1*time.Microsecond + 500*time.Nanosecond
	BeaconPeriod   = 128 * time.Second
	beaconReserved = 2*time.Second + 120*time.Millisecond
	pingSlotCount  = 4096
	pingSlotLen    = 30 * time.Millisecond
)

// beaconTimeBefore returns the last beacon time at or before t as time.Duration since GPS epoch.
func beaconTimeBefore(t time.Time) time.Duration {
	return gpstime.ToGPS(t) / BeaconPeriod * BeaconPeriod
}

// NextPingSlotAt returns the exact time instant before or at earliestAt when next ping slot can be open
// given the data known by Network Server and true, if such time instant exists, otherwise it returns time.Time{} and false.
func NextPingSlotAt(ctx context.Context, dev *ttnpb.EndDevice, earliestAt time.Time) (time.Time, bool) {
	if dev.GetSession() == nil || types.MustDevAddr(dev.Session.DevAddr).OrZero().IsZero() || dev.GetMacState() == nil || dev.MacState.PingSlotPeriodicity == nil {
		log.FromContext(ctx).Warn("Insufficient data to compute next ping slot")
		return time.Time{}, false
	}

	pingNb := uint16(1 << (7 - dev.MacState.PingSlotPeriodicity.Value))
	pingPeriod := uint16(1 << (5 + dev.MacState.PingSlotPeriodicity.Value))
	for beaconTime := beaconTimeBefore(earliestAt); beaconTime < math.MaxInt64; beaconTime += BeaconPeriod {
		pingOffset, err := crypto.ComputePingOffset(uint32(beaconTime/time.Second), types.MustDevAddr(dev.Session.DevAddr).OrZero(), pingPeriod)
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Failed to compute ping offset")
			return time.Time{}, false
		}

		t := gpstime.Parse(beaconTime + tBeaconDelay + beaconReserved + time.Duration(pingOffset)*pingSlotLen).UTC()
		if !earliestAt.After(t) {
			return t, true
		}
		sub := earliestAt.Sub(t)
		if sub >= BeaconPeriod {
			panic(fmt.Errorf("difference between earliestAt and first ping slot must be below '%s', got '%s'", BeaconPeriod, sub))
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

func DeviceResetsFCnt(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings) bool {
	switch {
	case dev.GetMacSettings().GetResetsFCnt() != nil:
		return dev.MacSettings.ResetsFCnt.Value
	case defaults.GetResetsFCnt() != nil:
		return defaults.ResetsFCnt.Value
	default:
		return false
	}
}

func DeviceSupports32BitFCnt(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings) bool {
	switch {
	case dev.GetMacSettings().GetSupports_32BitFCnt() != nil:
		return dev.MacSettings.Supports_32BitFCnt.Value
	case defaults.GetSupports_32BitFCnt() != nil:
		return defaults.Supports_32BitFCnt.Value
	default:
		return true
	}
}

var errClassAMulticast = errors.DefineInvalidArgument("class_a_multicast", "multicast device in class A mode")

func DeviceDefaultClass(dev *ttnpb.EndDevice) (ttnpb.Class, error) {
	switch {
	case !macspec.UseDeviceModeInd(dev.LorawanVersion) && dev.SupportsClassC:
		return ttnpb.Class_CLASS_C, nil
	case !dev.Multicast:
		return ttnpb.Class_CLASS_A, nil
	case dev.SupportsClassC:
		return ttnpb.Class_CLASS_C, nil
	case dev.SupportsClassB:
		return ttnpb.Class_CLASS_B, nil
	default:
		return ttnpb.Class_CLASS_A, errClassAMulticast.New()
	}
}

func DeviceDefaultLoRaWANVersion(dev *ttnpb.EndDevice) ttnpb.MACVersion {
	switch {
	case dev.Multicast:
		return dev.LorawanVersion
	case macspec.UseRekeyInd(dev.LorawanVersion):
		return macspec.RekeyPeriodVersion(dev.LorawanVersion)
	default:
		return dev.LorawanVersion
	}
}

func DeviceDefaultPingSlotPeriodicity(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings) *ttnpb.PingSlotPeriodValue {
	switch {
	case dev.GetMacSettings().GetPingSlotPeriodicity() != nil:
		return dev.MacSettings.PingSlotPeriodicity
	case defaults.GetPingSlotPeriodicity() != nil:
		return defaults.PingSlotPeriodicity
	default:
		return nil
	}
}

func DeviceUplinkDwellTime(dev *ttnpb.EndDevice, phy *band.Band, defaults *ttnpb.MACSettings) *ttnpb.BoolValue {
	switch {
	case !phy.TxParamSetupReqSupport:
		return nil
	case dev.GetMacSettings().GetUplinkDwellTime() != nil:
		return &ttnpb.BoolValue{Value: dev.MacSettings.UplinkDwellTime.Value}
	case defaults.UplinkDwellTime != nil:
		return &ttnpb.BoolValue{Value: defaults.UplinkDwellTime.Value}
	default:
		return nil
	}
}

func DeviceDownlinkDwellTime(dev *ttnpb.EndDevice, phy *band.Band, defaults *ttnpb.MACSettings) *ttnpb.BoolValue {
	switch {
	case !phy.TxParamSetupReqSupport:
		return nil
	case dev.GetMacSettings().GetDownlinkDwellTime() != nil:
		return &ttnpb.BoolValue{Value: dev.MacSettings.DownlinkDwellTime.Value}
	case defaults.DownlinkDwellTime != nil:
		return &ttnpb.BoolValue{Value: defaults.DownlinkDwellTime.Value}
	default:
		return nil
	}
}

func DeviceDesiredMaxEIRP(dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults *ttnpb.MACSettings) float32 {
	switch {
	case dev.GetMacSettings().GetDesiredMaxEirp() != nil:
		return lorawan.DeviceEIRPToFloat32(dev.GetMacSettings().GetDesiredMaxEirp().GetValue())
	case fp.MaxEIRP != nil && *fp.MaxEIRP > 0 && *fp.MaxEIRP < phy.DefaultMaxEIRP:
		return *fp.MaxEIRP
	default:
		return phy.DefaultMaxEIRP
	}
}

func DeviceDesiredUplinkDwellTime(phy *band.Band, fp *frequencyplans.FrequencyPlan) *ttnpb.BoolValue {
	switch {
	case !phy.TxParamSetupReqSupport:
		return nil
	case fp.DwellTime.Uplinks != nil:
		return &ttnpb.BoolValue{Value: *fp.DwellTime.Uplinks}
	default:
		return &ttnpb.BoolValue{Value: true}
	}
}

func DeviceDesiredDownlinkDwellTime(phy *band.Band, fp *frequencyplans.FrequencyPlan) *ttnpb.BoolValue {
	switch {
	case !phy.TxParamSetupReqSupport:
		return nil
	case fp.DwellTime.Downlinks != nil:
		return &ttnpb.BoolValue{Value: *fp.DwellTime.Downlinks}
	default:
		return &ttnpb.BoolValue{Value: true}
	}
}

func DeviceDefaultRX1Delay(dev *ttnpb.EndDevice, phy *band.Band, defaults *ttnpb.MACSettings) ttnpb.RxDelay {
	switch {
	case dev.GetMacSettings().GetRx1Delay() != nil:
		return dev.MacSettings.Rx1Delay.Value
	case defaults.Rx1Delay != nil:
		return defaults.Rx1Delay.Value
	default:
		return ttnpb.RxDelay(phy.ReceiveDelay1.Seconds())
	}
}

func DeviceDesiredRX1Delay(dev *ttnpb.EndDevice, phy *band.Band, defaults *ttnpb.MACSettings) ttnpb.RxDelay {
	switch {
	case dev.GetMacSettings().GetDesiredRx1Delay() != nil:
		return dev.MacSettings.DesiredRx1Delay.Value
	case defaults.DesiredRx1Delay != nil:
		return defaults.DesiredRx1Delay.Value
	default:
		return DeviceDefaultRX1Delay(dev, phy, defaults)
	}
}

func DeviceDesiredADRAckLimitExponent(dev *ttnpb.EndDevice, phy *band.Band, defaults *ttnpb.MACSettings) *ttnpb.ADRAckLimitExponentValue {
	switch {
	case dev.GetMacSettings().GetDesiredAdrAckLimitExponent() != nil:
		return &ttnpb.ADRAckLimitExponentValue{Value: dev.MacSettings.DesiredAdrAckLimitExponent.Value}
	case defaults.DesiredAdrAckLimitExponent != nil:
		return &ttnpb.ADRAckLimitExponentValue{Value: defaults.DesiredAdrAckLimitExponent.Value}
	default:
		return &ttnpb.ADRAckLimitExponentValue{Value: phy.ADRAckLimit}
	}
}

func DeviceDesiredADRAckDelayExponent(dev *ttnpb.EndDevice, phy *band.Band, defaults *ttnpb.MACSettings) *ttnpb.ADRAckDelayExponentValue {
	switch {
	case dev.GetMacSettings().GetDesiredAdrAckDelayExponent() != nil:
		return &ttnpb.ADRAckDelayExponentValue{Value: dev.MacSettings.DesiredAdrAckDelayExponent.Value}
	case defaults.DesiredAdrAckDelayExponent != nil:
		return &ttnpb.ADRAckDelayExponentValue{Value: defaults.DesiredAdrAckDelayExponent.Value}
	default:
		return &ttnpb.ADRAckDelayExponentValue{Value: phy.ADRAckDelay}
	}
}

func DeviceDefaultRX1DataRateOffset(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings) ttnpb.DataRateOffset {
	switch {
	case dev.GetMacSettings().GetRx1DataRateOffset() != nil:
		return dev.MacSettings.Rx1DataRateOffset.Value
	case defaults.Rx1DataRateOffset != nil:
		return defaults.Rx1DataRateOffset.Value
	default:
		return ttnpb.DataRateOffset_DATA_RATE_OFFSET_0
	}
}

func DeviceDesiredRX1DataRateOffset(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings) ttnpb.DataRateOffset {
	switch {
	case dev.GetMacSettings().GetDesiredRx1DataRateOffset() != nil:
		return dev.MacSettings.DesiredRx1DataRateOffset.Value
	case defaults.DesiredRx1DataRateOffset != nil:
		return defaults.DesiredRx1DataRateOffset.Value
	default:
		return DeviceDefaultRX1DataRateOffset(dev, defaults)
	}
}

func DeviceDefaultRX2DataRateIndex(dev *ttnpb.EndDevice, phy *band.Band, defaults *ttnpb.MACSettings) ttnpb.DataRateIndex {
	switch {
	case dev.GetMacSettings().GetRx2DataRateIndex() != nil:
		return dev.MacSettings.Rx2DataRateIndex.Value
	case defaults.Rx2DataRateIndex != nil:
		return defaults.Rx2DataRateIndex.Value
	default:
		return phy.DefaultRx2Parameters.DataRateIndex
	}
}

func DeviceDesiredRX2DataRateIndex(dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults *ttnpb.MACSettings) ttnpb.DataRateIndex {
	switch {
	case dev.GetMacSettings().GetDesiredRx2DataRateIndex() != nil:
		return dev.MacSettings.DesiredRx2DataRateIndex.Value
	case fp.DefaultRx2DataRate != nil:
		return ttnpb.DataRateIndex(*fp.DefaultRx2DataRate)
	case defaults.DesiredRx2DataRateIndex != nil:
		return defaults.DesiredRx2DataRateIndex.Value
	default:
		return DeviceDefaultRX2DataRateIndex(dev, phy, defaults)
	}
}

func DeviceDefaultRX2Frequency(dev *ttnpb.EndDevice, phy *band.Band, defaults *ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMacSettings().GetRx2Frequency() != nil && dev.MacSettings.Rx2Frequency.Value != 0:
		return dev.MacSettings.Rx2Frequency.Value
	case defaults.Rx2Frequency != nil && dev.MacSettings.Rx2Frequency.Value != 0:
		return defaults.Rx2Frequency.Value
	default:
		return phy.DefaultRx2Parameters.Frequency
	}
}

func DeviceDesiredRX2Frequency(dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults *ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMacSettings().GetDesiredRx2Frequency() != nil && dev.MacSettings.DesiredRx2Frequency.Value != 0:
		return dev.MacSettings.DesiredRx2Frequency.Value
	case fp.Rx2Channel != nil:
		return fp.Rx2Channel.Frequency
	case defaults.DesiredRx2Frequency != nil && defaults.DesiredRx2Frequency.Value != 0:
		return defaults.DesiredRx2Frequency.Value
	default:
		return DeviceDefaultRX2Frequency(dev, phy, defaults)
	}
}

func DeviceDefaultMaxDutyCycle(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings) ttnpb.AggregatedDutyCycle {
	switch {
	case dev.GetMacSettings().GetMaxDutyCycle() != nil:
		return dev.MacSettings.MaxDutyCycle.Value
	case defaults.MaxDutyCycle != nil:
		return defaults.MaxDutyCycle.Value
	default:
		return ttnpb.AggregatedDutyCycle_DUTY_CYCLE_1
	}
}

func DeviceDesiredMaxDutyCycle(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings) ttnpb.AggregatedDutyCycle {
	switch {
	case dev.GetMacSettings().GetDesiredMaxDutyCycle() != nil:
		return dev.MacSettings.DesiredMaxDutyCycle.Value
	case defaults.DesiredMaxDutyCycle != nil:
		return defaults.DesiredMaxDutyCycle.Value
	default:
		return DeviceDefaultMaxDutyCycle(dev, defaults)
	}
}

func DeviceDefaultPingSlotFrequency(dev *ttnpb.EndDevice, phy *band.Band, defaults *ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMacSettings().GetPingSlotFrequency() != nil && dev.MacSettings.PingSlotFrequency.Value != 0:
		return dev.MacSettings.PingSlotFrequency.Value
	case defaults.PingSlotFrequency != nil && defaults.PingSlotFrequency.Value != 0:
		return defaults.PingSlotFrequency.Value
	case phy.PingSlotFrequency != nil:
		return *phy.PingSlotFrequency
	default:
		return 0
	}
}

func DeviceDesiredPingSlotFrequency(dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults *ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMacSettings().GetDesiredPingSlotFrequency() != nil && dev.MacSettings.DesiredPingSlotFrequency.Value != 0:
		return dev.MacSettings.DesiredPingSlotFrequency.Value
	case fp.PingSlot != nil && fp.PingSlot.Frequency != 0:
		return fp.PingSlot.Frequency
	case defaults.DesiredPingSlotFrequency != nil && defaults.DesiredPingSlotFrequency.Value != 0:
		return defaults.DesiredPingSlotFrequency.Value
	default:
		return DeviceDefaultPingSlotFrequency(dev, phy, defaults)
	}
}

func DeviceDefaultPingSlotDataRateIndexValue(dev *ttnpb.EndDevice, phy *band.Band, defaults *ttnpb.MACSettings) *ttnpb.DataRateIndexValue {
	switch {
	case dev.GetMacSettings().GetPingSlotDataRateIndex() != nil:
		return dev.MacSettings.PingSlotDataRateIndex
	case defaults.PingSlotDataRateIndex != nil:
		return defaults.PingSlotDataRateIndex
	default:
		// Default to mbed-os and LoRaMac-node behavior.
		// https://github.com/Lora-net/LoRaMac-node/blob/87f19e84ae2fc4af72af9567fe722386de6ce9f4/src/mac/region/RegionCN779.h#L235.
		return &ttnpb.DataRateIndexValue{Value: phy.Beacon.DataRateIndex}
	}
}

func DeviceDesiredPingSlotDataRateIndexValue(dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults *ttnpb.MACSettings) *ttnpb.DataRateIndexValue {
	switch {
	case dev.GetMacSettings().GetDesiredPingSlotDataRateIndex() != nil:
		return dev.MacSettings.DesiredPingSlotDataRateIndex
	case fp.DefaultPingSlotDataRate != nil:
		return &ttnpb.DataRateIndexValue{Value: ttnpb.DataRateIndex(*fp.DefaultPingSlotDataRate)}
	case defaults.DesiredPingSlotDataRateIndex != nil:
		return defaults.DesiredPingSlotDataRateIndex
	default:
		return DeviceDefaultPingSlotDataRateIndexValue(dev, phy, defaults)
	}
}

func DeviceDefaultBeaconFrequency(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMacSettings().GetBeaconFrequency() != nil && dev.MacSettings.BeaconFrequency.Value != 0:
		return dev.MacSettings.BeaconFrequency.Value
	case defaults.BeaconFrequency != nil:
		return defaults.BeaconFrequency.Value
	default:
		return 0
	}
}

func DeviceDesiredBeaconFrequency(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMacSettings().GetDesiredBeaconFrequency() != nil && dev.MacSettings.DesiredBeaconFrequency.Value != 0:
		return dev.MacSettings.DesiredBeaconFrequency.Value
	case defaults.DesiredBeaconFrequency != nil && defaults.DesiredBeaconFrequency.Value != 0:
		return defaults.DesiredBeaconFrequency.Value
	default:
		return DeviceDefaultBeaconFrequency(dev, defaults)
	}
}

func deviceFactoryPresetFrequencies(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings) []uint64 {
	if freqs := dev.GetMacSettings().GetFactoryPresetFrequencies(); len(freqs) > 0 {
		return freqs
	}
	return defaults.GetFactoryPresetFrequencies()
}

func DeviceDefaultChannels(
	dev *ttnpb.EndDevice, phy *band.Band, defaults *ttnpb.MACSettings,
) []*ttnpb.MACParameters_Channel {
	if len(phy.DownlinkChannels) > len(phy.UplinkChannels) ||
		len(phy.UplinkChannels) > int(phy.MaxUplinkChannels) ||
		len(phy.DownlinkChannels) > int(phy.MaxDownlinkChannels) {
		// NOTE: In case the spec changes and this assumption is not valid anymore,
		// the implementation of this function won't be valid and has to be changed.
		panic("uplink/downlink channel length is inconsistent")
	}

	factoryPresetFreqs := deviceFactoryPresetFrequencies(dev, defaults)

	chs := make([]*ttnpb.MACParameters_Channel, 0, len(phy.UplinkChannels)+len(factoryPresetFreqs))
	for i, phyUpCh := range phy.UplinkChannels {
		downFreq := phy.DownlinkChannels[i%len(phy.DownlinkChannels)].Frequency
		if dev.Multicast {
			chs = append(chs, &ttnpb.MACParameters_Channel{
				DownlinkFrequency: downFreq,
			})
			continue
		}
		chs = append(chs, &ttnpb.MACParameters_Channel{
			MinDataRateIndex:  phyUpCh.MinDataRate,
			MaxDataRateIndex:  phyUpCh.MaxDataRate,
			UplinkFrequency:   phyUpCh.Frequency,
			DownlinkFrequency: downFreq,
			EnableUplink:      len(factoryPresetFreqs) == 0,
		})
	}

outer:
	for _, freq := range factoryPresetFreqs {
		for _, ch := range chs {
			if ch.UplinkFrequency == freq {
				ch.EnableUplink = true
				// NOTE: duplicates should not be allowed.
				continue outer
			}
		}
		if dev.Multicast {
			chs = append(chs, &ttnpb.MACParameters_Channel{
				DownlinkFrequency: freq,
			})
			continue
		}
		chs = append(chs, &ttnpb.MACParameters_Channel{
			MaxDataRateIndex:  phy.MaxADRDataRateIndex,
			UplinkFrequency:   freq,
			DownlinkFrequency: freq,
			EnableUplink:      true,
		})
	}
	return chs
}

func DeviceDesiredChannels(
	dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults *ttnpb.MACSettings,
) ([]*ttnpb.MACParameters_Channel, error) {
	if len(phy.DownlinkChannels) > len(phy.UplinkChannels) ||
		len(phy.UplinkChannels) > int(phy.MaxUplinkChannels) ||
		len(phy.DownlinkChannels) > int(phy.MaxDownlinkChannels) ||
		len(fp.DownlinkChannels) != 0 && len(fp.DownlinkChannels) != len(fp.UplinkChannels) ||
		len(fp.UplinkChannels) > int(phy.MaxUplinkChannels) ||
		len(fp.DownlinkChannels) > int(phy.MaxDownlinkChannels) {
		// NOTE: In case the spec changes and this assumption is not valid anymore,
		// the implementation of this function won't be valid and has to be changed.
		panic("uplink/downlink channel length is inconsistent")
	}

	defaultChs := DeviceDefaultChannels(dev, phy, defaults)

	chs := make([]*ttnpb.MACParameters_Channel, 0, len(defaultChs)+len(fp.UplinkChannels))
	for _, ch := range defaultChs {
		chs = append(chs, &ttnpb.MACParameters_Channel{
			MinDataRateIndex:  ch.MinDataRateIndex,
			MaxDataRateIndex:  ch.MaxDataRateIndex,
			UplinkFrequency:   ch.UplinkFrequency,
			DownlinkFrequency: ch.DownlinkFrequency,
		})
	}

	latestPhy, err := band.GetLatest(phy.ID)
	if err != nil {
		return nil, err
	}

outer:
	for i, fpUpCh := range fp.UplinkChannels {
		minDataRateIndex, maxDataRateIndex, err := mapDataRateIndices(
			&latestPhy,
			ttnpb.DataRateIndex(fpUpCh.MinDataRate),
			ttnpb.DataRateIndex(fpUpCh.MaxDataRate),
			phy,
		)
		if err != nil {
			return nil, err
		}
		for _, ch := range chs {
			if ch.UplinkFrequency == fpUpCh.Frequency {
				ch.MinDataRateIndex = minDataRateIndex
				ch.MaxDataRateIndex = maxDataRateIndex
				ch.EnableUplink = true
				// NOTE: duplicates should not be allowed.
				continue outer
			}
		}
		downFreq := fpUpCh.Frequency
		if i < len(fp.DownlinkChannels) {
			downFreq = fp.DownlinkChannels[i].Frequency
		}
		chs = append(chs, &ttnpb.MACParameters_Channel{
			MinDataRateIndex:  minDataRateIndex,
			MaxDataRateIndex:  maxDataRateIndex,
			UplinkFrequency:   fpUpCh.Frequency,
			DownlinkFrequency: downFreq,
			EnableUplink:      true,
		})
	}
	return chs, nil
}

const defaultClassBCDownlinkInterval = time.Second

func DeviceClassBCDownlinkInterval(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings) time.Duration {
	if t := dev.GetMacSettings().GetClassBCDownlinkInterval(); t != nil {
		return ttnpb.StdDurationOrZero(t)
	}
	if t := defaults.GetClassBCDownlinkInterval(); t != nil {
		return ttnpb.StdDurationOrZero(t)
	}
	return defaultClassBCDownlinkInterval
}

func NewState(dev *ttnpb.EndDevice, fps *frequencyplans.Store, defaults *ttnpb.MACSettings) (*ttnpb.MACState, error) {
	fp, phy, err := internal.DeviceFrequencyPlanAndBand(dev, fps)
	if err != nil {
		return nil, err
	}
	class, err := DeviceDefaultClass(dev)
	if err != nil {
		return nil, err
	}

	current := &ttnpb.MACParameters{
		MaxEirp:                    phy.DefaultMaxEIRP,
		AdrDataRateIndex:           ttnpb.DataRateIndex_DATA_RATE_0,
		AdrNbTrans:                 1,
		Rx1Delay:                   DeviceDefaultRX1Delay(dev, phy, defaults),
		Rx1DataRateOffset:          DeviceDefaultRX1DataRateOffset(dev, defaults),
		Rx2DataRateIndex:           DeviceDefaultRX2DataRateIndex(dev, phy, defaults),
		Rx2Frequency:               DeviceDefaultRX2Frequency(dev, phy, defaults),
		MaxDutyCycle:               DeviceDefaultMaxDutyCycle(dev, defaults),
		RejoinTimePeriodicity:      ttnpb.RejoinTimeExponent_REJOIN_TIME_0,
		RejoinCountPeriodicity:     ttnpb.RejoinCountExponent_REJOIN_COUNT_16,
		PingSlotFrequency:          DeviceDefaultPingSlotFrequency(dev, phy, defaults),
		BeaconFrequency:            DeviceDefaultBeaconFrequency(dev, defaults),
		Channels:                   DeviceDefaultChannels(dev, phy, defaults),
		UplinkDwellTime:            DeviceUplinkDwellTime(dev, phy, defaults),
		DownlinkDwellTime:          DeviceDownlinkDwellTime(dev, phy, defaults),
		AdrAckLimitExponent:        &ttnpb.ADRAckLimitExponentValue{Value: phy.ADRAckLimit},
		AdrAckDelayExponent:        &ttnpb.ADRAckDelayExponentValue{Value: phy.ADRAckDelay},
		PingSlotDataRateIndexValue: DeviceDefaultPingSlotDataRateIndexValue(dev, phy, defaults),
	}
	desired := current
	if !dev.Multicast {
		deviceDesiredChannels, err := DeviceDesiredChannels(dev, phy, fp, defaults)
		if err != nil {
			return nil, err
		}
		desired = &ttnpb.MACParameters{
			MaxEirp:                    DeviceDesiredMaxEIRP(dev, phy, fp, defaults),
			AdrDataRateIndex:           ttnpb.DataRateIndex_DATA_RATE_0,
			AdrNbTrans:                 1,
			Rx1Delay:                   DeviceDesiredRX1Delay(dev, phy, defaults),
			Rx1DataRateOffset:          DeviceDesiredRX1DataRateOffset(dev, defaults),
			Rx2DataRateIndex:           DeviceDesiredRX2DataRateIndex(dev, phy, fp, defaults),
			Rx2Frequency:               DeviceDesiredRX2Frequency(dev, phy, fp, defaults),
			MaxDutyCycle:               DeviceDesiredMaxDutyCycle(dev, defaults),
			RejoinTimePeriodicity:      ttnpb.RejoinTimeExponent_REJOIN_TIME_0,
			RejoinCountPeriodicity:     ttnpb.RejoinCountExponent_REJOIN_COUNT_16,
			PingSlotFrequency:          DeviceDesiredPingSlotFrequency(dev, phy, fp, defaults),
			BeaconFrequency:            DeviceDesiredBeaconFrequency(dev, defaults),
			Channels:                   deviceDesiredChannels,
			UplinkDwellTime:            DeviceDesiredUplinkDwellTime(phy, fp),
			DownlinkDwellTime:          DeviceDesiredDownlinkDwellTime(phy, fp),
			AdrAckLimitExponent:        DeviceDesiredADRAckLimitExponent(dev, phy, defaults),
			AdrAckDelayExponent:        DeviceDesiredADRAckDelayExponent(dev, phy, defaults),
			PingSlotDataRateIndexValue: DeviceDesiredPingSlotDataRateIndexValue(dev, phy, fp, defaults),
		}
	}
	// TODO: Support rejoins. (https://github.com/TheThingsNetwork/lorawan-stack/issues/8)
	return &ttnpb.MACState{
		LorawanVersion:      DeviceDefaultLoRaWANVersion(dev),
		DeviceClass:         class,
		PingSlotPeriodicity: DeviceDefaultPingSlotPeriodicity(dev, defaults),
		CurrentParameters:   current,
		DesiredParameters:   desired,
	}, nil
}

func DeviceExpectedUplinkDwellTime(macState *ttnpb.MACState, fp *frequencyplans.FrequencyPlan, phy *band.Band) bool {
	switch {
	case macState.GetCurrentParameters().GetUplinkDwellTime() != nil:
		return macState.CurrentParameters.UplinkDwellTime.Value
	case phy.BootDwellTime.Uplinks != nil:
		return *phy.BootDwellTime.Uplinks
	case fp.DwellTime.Uplinks != nil:
		return *fp.DwellTime.Uplinks
	default:
		return false
	}
}

func DeviceExpectedDownlinkDwellTime(macState *ttnpb.MACState, fp *frequencyplans.FrequencyPlan, phy *band.Band) bool {
	switch {
	case macState.GetCurrentParameters().GetDownlinkDwellTime() != nil:
		return macState.CurrentParameters.DownlinkDwellTime.Value
	case phy.BootDwellTime.Downlinks != nil:
		return *phy.BootDwellTime.Downlinks
	case fp.DwellTime.Downlinks != nil:
		return *fp.DwellTime.Downlinks
	default:
		return true
	}
}

// DeviceScheduleDownlinks checks if the Network Server should schedule downlinks
// for the provided end device.
func DeviceScheduleDownlinks(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings) bool {
	switch {
	case dev.GetMacSettings().GetScheduleDownlinks() != nil:
		return dev.MacSettings.ScheduleDownlinks.Value
	case defaults.GetScheduleDownlinks() != nil:
		return defaults.ScheduleDownlinks.Value
	default:
		return true
	}
}

// mapDataRateIndices maps the provided data rate indices from the latest version
// of the band to the provided target band. If the minimum data rate index is not
// available, ttnpb.DataRateIndex_DATA_RATE_0 will be used.
func mapDataRateIndices(
	sourceBand *band.Band,
	sourceMinDataRateIndex,
	sourceMaxDataRateIndex ttnpb.DataRateIndex,
	targetBand *band.Band,
) (targetMinDataRateIndex, targetMaxDataRateIndex ttnpb.DataRateIndex, err error) {
	targetMinDataRateIndex, err = band.MapDataRateIndex(sourceBand, sourceMinDataRateIndex, targetBand)
	if err != nil {
		targetMinDataRateIndex = ttnpb.DataRateIndex_DATA_RATE_0
	}
	targetMaxDataRateIndex, err = band.MapDataRateIndex(sourceBand, sourceMaxDataRateIndex, targetBand)
	if err != nil {
		return 0, 0, err
	}
	return targetMinDataRateIndex, targetMaxDataRateIndex, nil
}

func containsMACCommandIdentifier(cid ttnpb.MACCommandIdentifier) func(...ttnpb.MACCommandIdentifier) bool {
	return func(cmds ...ttnpb.MACCommandIdentifier) bool {
		return slices.Contains(cmds, cid)
	}
}

func consumeMACCommandIdentifier(
	cid ttnpb.MACCommandIdentifier,
) func(...ttnpb.MACCommandIdentifier) ([]ttnpb.MACCommandIdentifier, bool) {
	return func(cmds ...ttnpb.MACCommandIdentifier) ([]ttnpb.MACCommandIdentifier, bool) {
		if i := slices.Index(cmds, cid); i >= 0 {
			return append(cmds[:i], cmds[i+1:]...), true
		}
		return cmds, false
	}
}
