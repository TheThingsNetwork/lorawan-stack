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
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func channelDataRateRange(chs ...*ttnpb.MACParameters_Channel) (min, max ttnpb.DataRateIndex, ok bool) {
	i := 0
	for {
		if i >= len(chs) {
			return 0, 0, false
		}
		if chs[i].EnableUplink {
			break
		}
		i++
	}

	min = chs[i].MinDataRateIndex
	max = chs[i].MaxDataRateIndex
	for _, ch := range chs[i+1:] {
		if !ch.EnableUplink {
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

func DeviceClassBTimeout(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) time.Duration {
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

func DeviceClassCTimeout(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) time.Duration {
	if dev.MACSettings != nil && dev.MACSettings.ClassCTimeout != nil {
		return *dev.MACSettings.ClassCTimeout
	}
	if defaults.ClassCTimeout != nil {
		return *defaults.ClassCTimeout
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
	if dev.GetSession() == nil || dev.Session.DevAddr.IsZero() || dev.GetMACState() == nil || dev.MACState.PingSlotPeriodicity == nil {
		log.FromContext(ctx).Warn("Insufficient data to compute next ping slot")
		return time.Time{}, false
	}

	pingNb := uint16(1 << (7 - dev.MACState.PingSlotPeriodicity.Value))
	pingPeriod := uint16(1 << (5 + dev.MACState.PingSlotPeriodicity.Value))
	for beaconTime := beaconTimeBefore(earliestAt); beaconTime < math.MaxInt64; beaconTime += BeaconPeriod {
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

func DeviceUseADR(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings, phy *band.Band) bool {
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

func DeviceResetsFCnt(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) bool {
	switch {
	case dev.GetMACSettings().GetResetsFCnt() != nil:
		return dev.MACSettings.ResetsFCnt.Value
	case defaults.GetResetsFCnt() != nil:
		return defaults.ResetsFCnt.Value
	default:
		return false
	}
}

func DeviceSupports32BitFCnt(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) bool {
	switch {
	case dev.GetMACSettings().GetSupports32BitFCnt() != nil:
		return dev.MACSettings.Supports32BitFCnt.Value
	case defaults.GetSupports32BitFCnt() != nil:
		return defaults.Supports32BitFCnt.Value
	default:
		return true
	}
}

var errClassAMulticast = errors.DefineInvalidArgument("class_a_multicast", "multicast device in class A mode")

func DeviceDefaultClass(dev *ttnpb.EndDevice) (ttnpb.Class, error) {
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

func DeviceDefaultLoRaWANVersion(dev *ttnpb.EndDevice) ttnpb.MACVersion {
	switch {
	case dev.Multicast:
		return dev.LoRaWANVersion
	case dev.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0:
		return ttnpb.MAC_V1_1
	default:
		return dev.LoRaWANVersion
	}
}

func DeviceDefaultPingSlotPeriodicity(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) *ttnpb.PingSlotPeriodValue {
	switch {
	case dev.GetMACSettings().GetPingSlotPeriodicity() != nil:
		return dev.MACSettings.PingSlotPeriodicity
	case defaults.GetPingSlotPeriodicity() != nil:
		return defaults.PingSlotPeriodicity
	default:
		return nil
	}
}

func DeviceDesiredMaxEIRP(dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults ttnpb.MACSettings) float32 {
	switch {
	case dev.GetMACSettings().GetDesiredMaxEirp() != nil:
		return lorawan.DeviceEIRPToFloat32(dev.GetMACSettings().GetDesiredMaxEirp().GetValue())
	case fp.MaxEIRP != nil && *fp.MaxEIRP > 0 && *fp.MaxEIRP < phy.DefaultMaxEIRP:
		return *fp.MaxEIRP
	default:
		return phy.DefaultMaxEIRP
	}
}

func DeviceDesiredUplinkDwellTime(fp *frequencyplans.FrequencyPlan) *ttnpb.BoolValue {
	if fp.DwellTime.Uplinks == nil {
		return nil
	}
	return &ttnpb.BoolValue{Value: *fp.DwellTime.Uplinks}
}

func DeviceDesiredDownlinkDwellTime(fp *frequencyplans.FrequencyPlan) *ttnpb.BoolValue {
	if fp.DwellTime.Downlinks == nil {
		return nil
	}
	return &ttnpb.BoolValue{Value: *fp.DwellTime.Downlinks}
}

func DeviceDefaultRX1Delay(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) ttnpb.RxDelay {
	switch {
	case dev.GetMACSettings().GetRx1Delay() != nil:
		return dev.MACSettings.Rx1Delay.Value
	case defaults.Rx1Delay != nil:
		return defaults.Rx1Delay.Value
	default:
		return ttnpb.RxDelay(phy.ReceiveDelay1.Seconds())
	}
}

func DeviceDesiredRX1Delay(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) ttnpb.RxDelay {
	switch {
	case dev.GetMACSettings().GetDesiredRx1Delay() != nil:
		return dev.MACSettings.DesiredRx1Delay.Value
	case defaults.DesiredRx1Delay != nil:
		return defaults.DesiredRx1Delay.Value
	default:
		return DeviceDefaultRX1Delay(dev, phy, defaults)
	}
}

func DeviceDesiredADRAckLimitExponent(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) *ttnpb.ADRAckLimitExponentValue {
	switch {
	case dev.GetMACSettings().GetDesiredADRAckLimitExponent() != nil:
		return &ttnpb.ADRAckLimitExponentValue{Value: dev.MACSettings.DesiredADRAckLimitExponent.Value}
	case defaults.DesiredADRAckLimitExponent != nil:
		return &ttnpb.ADRAckLimitExponentValue{Value: defaults.DesiredADRAckLimitExponent.Value}
	default:
		return &ttnpb.ADRAckLimitExponentValue{Value: phy.ADRAckLimit}
	}
}

func DeviceDesiredADRAckDelayExponent(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) *ttnpb.ADRAckDelayExponentValue {
	switch {
	case dev.GetMACSettings().GetDesiredADRAckDelayExponent() != nil:
		return &ttnpb.ADRAckDelayExponentValue{Value: dev.MACSettings.DesiredADRAckDelayExponent.Value}
	case defaults.DesiredADRAckDelayExponent != nil:
		return &ttnpb.ADRAckDelayExponentValue{Value: defaults.DesiredADRAckDelayExponent.Value}
	default:
		return &ttnpb.ADRAckDelayExponentValue{Value: phy.ADRAckDelay}
	}
}

func DeviceDefaultRX1DataRateOffset(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) ttnpb.DataRateOffset {
	switch {
	case dev.GetMACSettings().GetRx1DataRateOffset() != nil:
		return dev.MACSettings.Rx1DataRateOffset.Value
	case defaults.Rx1DataRateOffset != nil:
		return defaults.Rx1DataRateOffset.Value
	default:
		return ttnpb.DataRateOffset_DATA_RATE_OFFSET_0
	}
}

func DeviceDesiredRX1DataRateOffset(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) ttnpb.DataRateOffset {
	switch {
	case dev.GetMACSettings().GetDesiredRx1DataRateOffset() != nil:
		return dev.MACSettings.DesiredRx1DataRateOffset.Value
	case defaults.DesiredRx1DataRateOffset != nil:
		return defaults.DesiredRx1DataRateOffset.Value
	default:
		return DeviceDefaultRX1DataRateOffset(dev, defaults)
	}
}

func DeviceDefaultRX2DataRateIndex(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) ttnpb.DataRateIndex {
	switch {
	case dev.GetMACSettings().GetRx2DataRateIndex() != nil:
		return dev.MACSettings.Rx2DataRateIndex.Value
	case defaults.Rx2DataRateIndex != nil:
		return defaults.Rx2DataRateIndex.Value
	default:
		return phy.DefaultRx2Parameters.DataRateIndex
	}
}

func DeviceDesiredRX2DataRateIndex(dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults ttnpb.MACSettings) ttnpb.DataRateIndex {
	switch {
	case dev.GetMACSettings().GetDesiredRx2DataRateIndex() != nil:
		return dev.MACSettings.DesiredRx2DataRateIndex.Value
	case fp.DefaultRx2DataRate != nil:
		return ttnpb.DataRateIndex(*fp.DefaultRx2DataRate)
	case defaults.DesiredRx2DataRateIndex != nil:
		return defaults.DesiredRx2DataRateIndex.Value
	default:
		return DeviceDefaultRX2DataRateIndex(dev, phy, defaults)
	}
}

func DeviceDefaultRX2Frequency(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMACSettings().GetRx2Frequency() != nil && dev.MACSettings.Rx2Frequency.Value != 0:
		return dev.MACSettings.Rx2Frequency.Value
	case defaults.Rx2Frequency != nil && dev.MACSettings.Rx2Frequency.Value != 0:
		return defaults.Rx2Frequency.Value
	default:
		return phy.DefaultRx2Parameters.Frequency
	}
}

func DeviceDesiredRX2Frequency(dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMACSettings().GetDesiredRx2Frequency() != nil && dev.MACSettings.DesiredRx2Frequency.Value != 0:
		return dev.MACSettings.DesiredRx2Frequency.Value
	case fp.Rx2Channel != nil:
		return fp.Rx2Channel.Frequency
	case defaults.DesiredRx2Frequency != nil && defaults.DesiredRx2Frequency.Value != 0:
		return defaults.DesiredRx2Frequency.Value
	default:
		return DeviceDefaultRX2Frequency(dev, phy, defaults)
	}
}

func DeviceDefaultMaxDutyCycle(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) ttnpb.AggregatedDutyCycle {
	switch {
	case dev.GetMACSettings().GetMaxDutyCycle() != nil:
		return dev.MACSettings.MaxDutyCycle.Value
	case defaults.MaxDutyCycle != nil:
		return defaults.MaxDutyCycle.Value
	default:
		return ttnpb.DUTY_CYCLE_1
	}
}

func DeviceDesiredMaxDutyCycle(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) ttnpb.AggregatedDutyCycle {
	switch {
	case dev.GetMACSettings().GetDesiredMaxDutyCycle() != nil:
		return dev.MACSettings.DesiredMaxDutyCycle.Value
	case defaults.DesiredMaxDutyCycle != nil:
		return defaults.DesiredMaxDutyCycle.Value
	default:
		return DeviceDefaultMaxDutyCycle(dev, defaults)
	}
}

func DeviceDefaultPingSlotFrequency(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) uint64 {
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

func DeviceDesiredPingSlotFrequency(dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMACSettings().GetDesiredPingSlotFrequency() != nil && dev.MACSettings.DesiredPingSlotFrequency.Value != 0:
		return dev.MACSettings.DesiredPingSlotFrequency.Value
	case fp.PingSlot != nil && fp.PingSlot.Frequency != 0:
		return fp.PingSlot.Frequency
	case defaults.DesiredPingSlotFrequency != nil && defaults.DesiredPingSlotFrequency.Value != 0:
		return defaults.DesiredPingSlotFrequency.Value
	default:
		return DeviceDefaultPingSlotFrequency(dev, phy, defaults)
	}
}

func DeviceDefaultPingSlotDataRateIndexValue(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) *ttnpb.DataRateIndexValue {
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

func DeviceDesiredPingSlotDataRateIndexValue(dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults ttnpb.MACSettings) *ttnpb.DataRateIndexValue {
	switch {
	case dev.GetMACSettings().GetDesiredPingSlotDataRateIndex() != nil:
		return dev.MACSettings.DesiredPingSlotDataRateIndex
	case fp.DefaultPingSlotDataRate != nil:
		return &ttnpb.DataRateIndexValue{Value: ttnpb.DataRateIndex(*fp.DefaultPingSlotDataRate)}
	case defaults.DesiredPingSlotDataRateIndex != nil:
		return defaults.DesiredPingSlotDataRateIndex
	default:
		return DeviceDefaultPingSlotDataRateIndexValue(dev, phy, defaults)
	}
}

func DeviceDefaultBeaconFrequency(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMACSettings().GetBeaconFrequency() != nil && dev.MACSettings.BeaconFrequency.Value != 0:
		return dev.MACSettings.BeaconFrequency.Value
	case defaults.BeaconFrequency != nil:
		return defaults.BeaconFrequency.Value
	default:
		return 0
	}
}

func DeviceDesiredBeaconFrequency(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) uint64 {
	switch {
	case dev.GetMACSettings().GetDesiredBeaconFrequency() != nil && dev.MACSettings.DesiredBeaconFrequency.Value != 0:
		return dev.MACSettings.DesiredBeaconFrequency.Value
	case defaults.DesiredBeaconFrequency != nil && defaults.DesiredBeaconFrequency.Value != 0:
		return defaults.DesiredBeaconFrequency.Value
	default:
		return DeviceDefaultBeaconFrequency(dev, defaults)
	}
}

func deviceFactoryPresetFrequencies(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) []uint64 {
	if freqs := dev.GetMACSettings().GetFactoryPresetFrequencies(); len(freqs) > 0 {
		return freqs
	}
	return defaults.GetFactoryPresetFrequencies()
}

func DeviceDefaultChannels(dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) []*ttnpb.MACParameters_Channel {
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
		// NOTE: FactoryPresetFrequencies does not indicate the data rate ranges allowed for channels.
		// In the latest regional parameters spec(1.1b) the data rate ranges are DR0-DR5 for mandatory channels in all non-fixed channel plans,
		// hence we assume the same range for predefined channels.
		chs = append(chs, &ttnpb.MACParameters_Channel{
			MaxDataRateIndex:  ttnpb.DATA_RATE_5,
			UplinkFrequency:   freq,
			DownlinkFrequency: freq,
			EnableUplink:      true,
		})
	}
	return chs
}

func DeviceDesiredChannels(dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, defaults ttnpb.MACSettings) []*ttnpb.MACParameters_Channel {
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

outer:
	for i, fpUpCh := range fp.UplinkChannels {
		for _, ch := range chs {
			if ch.UplinkFrequency == fpUpCh.Frequency {
				ch.MinDataRateIndex = ttnpb.DataRateIndex(fpUpCh.MinDataRate)
				ch.MaxDataRateIndex = ttnpb.DataRateIndex(fpUpCh.MaxDataRate)
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
			MinDataRateIndex:  ttnpb.DataRateIndex(fpUpCh.MinDataRate),
			MaxDataRateIndex:  ttnpb.DataRateIndex(fpUpCh.MaxDataRate),
			UplinkFrequency:   fpUpCh.Frequency,
			DownlinkFrequency: downFreq,
			EnableUplink:      true,
		})
	}
	return chs
}

func NewState(dev *ttnpb.EndDevice, fps *frequencyplans.Store, defaults ttnpb.MACSettings) (*ttnpb.MACState, error) {
	fp, phy, err := DeviceFrequencyPlanAndBand(dev, fps)
	if err != nil {
		return nil, err
	}
	class, err := DeviceDefaultClass(dev)
	if err != nil {
		return nil, err
	}

	current := ttnpb.MACParameters{
		MaxEIRP:                    phy.DefaultMaxEIRP,
		ADRDataRateIndex:           ttnpb.DATA_RATE_0,
		ADRNbTrans:                 1,
		Rx1Delay:                   DeviceDefaultRX1Delay(dev, phy, defaults),
		Rx1DataRateOffset:          DeviceDefaultRX1DataRateOffset(dev, defaults),
		Rx2DataRateIndex:           DeviceDefaultRX2DataRateIndex(dev, phy, defaults),
		Rx2Frequency:               DeviceDefaultRX2Frequency(dev, phy, defaults),
		MaxDutyCycle:               DeviceDefaultMaxDutyCycle(dev, defaults),
		RejoinTimePeriodicity:      ttnpb.REJOIN_TIME_0,
		RejoinCountPeriodicity:     ttnpb.REJOIN_COUNT_16,
		PingSlotFrequency:          DeviceDefaultPingSlotFrequency(dev, phy, defaults),
		BeaconFrequency:            DeviceDefaultBeaconFrequency(dev, defaults),
		Channels:                   DeviceDefaultChannels(dev, phy, defaults),
		ADRAckLimitExponent:        &ttnpb.ADRAckLimitExponentValue{Value: phy.ADRAckLimit},
		ADRAckDelayExponent:        &ttnpb.ADRAckDelayExponentValue{Value: phy.ADRAckDelay},
		PingSlotDataRateIndexValue: DeviceDefaultPingSlotDataRateIndexValue(dev, phy, defaults),
	}
	desired := current
	if !dev.Multicast {
		desired = ttnpb.MACParameters{
			MaxEIRP:                    DeviceDesiredMaxEIRP(dev, phy, fp, defaults),
			ADRDataRateIndex:           ttnpb.DATA_RATE_0,
			ADRNbTrans:                 1,
			Rx1Delay:                   DeviceDesiredRX1Delay(dev, phy, defaults),
			Rx1DataRateOffset:          DeviceDesiredRX1DataRateOffset(dev, defaults),
			Rx2DataRateIndex:           DeviceDesiredRX2DataRateIndex(dev, phy, fp, defaults),
			Rx2Frequency:               DeviceDesiredRX2Frequency(dev, phy, fp, defaults),
			MaxDutyCycle:               DeviceDesiredMaxDutyCycle(dev, defaults),
			RejoinTimePeriodicity:      ttnpb.REJOIN_TIME_0,
			RejoinCountPeriodicity:     ttnpb.REJOIN_COUNT_16,
			PingSlotFrequency:          DeviceDesiredPingSlotFrequency(dev, phy, fp, defaults),
			BeaconFrequency:            DeviceDesiredBeaconFrequency(dev, defaults),
			Channels:                   DeviceDesiredChannels(dev, phy, fp, defaults),
			UplinkDwellTime:            DeviceDesiredUplinkDwellTime(fp),
			DownlinkDwellTime:          DeviceDesiredDownlinkDwellTime(fp),
			ADRAckLimitExponent:        DeviceDesiredADRAckLimitExponent(dev, phy, defaults),
			ADRAckDelayExponent:        DeviceDesiredADRAckDelayExponent(dev, phy, defaults),
			PingSlotDataRateIndexValue: DeviceDesiredPingSlotDataRateIndexValue(dev, phy, fp, defaults),
		}
	}
	// TODO: Support rejoins. (https://github.com/TheThingsNetwork/lorawan-stack/issues/8)
	return &ttnpb.MACState{
		LoRaWANVersion:      DeviceDefaultLoRaWANVersion(dev),
		DeviceClass:         class,
		PingSlotPeriodicity: DeviceDefaultPingSlotPeriodicity(dev, defaults),
		CurrentParameters:   current,
		DesiredParameters:   desired,
	}, nil
}
