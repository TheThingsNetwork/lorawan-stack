// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var demodulationFloor = map[uint32]map[uint32]float32{
	6: {
		125_000: -5,
		250_000: -2,
		500_000: 1,
	},
	7: {
		125_000: -7.5,
		250_000: -4.5,
		500_000: -1.5,
	},
	8: {
		125_000: -10,
		250_000: -7,
		500_000: -4,
	},
	9: {
		125_000: -12.5,
		250_000: -9.5,
		500_000: -6.5,
	},
	10: {
		125_000: -15,
		250_000: -12,
		500_000: -9,
	},
	11: {
		125_000: -17.5,
		250_000: -14.5,
		500_000: -11.5,
	},
	12: {
		125_000: -20,
		250_000: -17,
		500_000: -14,
	},
}

const (
	// safetyMargin is a margin in dB added for ADR parameter calculation if less than 20 uplinks are available.
	safetyMargin = 2.5

	// drStep is the difference between 2 consequitive data rates in dB.
	drStep = 2.5

	// maxNbTrans is the maximum NbTrans parameter used by the algorithm.
	maxNbTrans = 3

	// OptimalADRUplinkCount is the amount of uplinks required to ensure optimal results from the ADR algorithm.
	OptimalADRUplinkCount = 20

	// DefaultADRMargin is the default ADR margin used if not specified in MACSettings of the device or NS-wide defaults.
	DefaultADRMargin = 15
)

func deviceUseADR(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings, phy *band.Band) bool {
	switch {
	case dev.GetMulticast():
		return false

	case dev.GetMacSettings().GetAdr().GetDisabled() != nil:
		return false
	case dev.GetMacSettings().GetAdr().GetStatic() != nil:
		return true

	case defaults.GetAdr().GetDisabled() != nil:
		return false
	case defaults.GetAdr().GetStatic() != nil:
		return true

	case !phy.SupportsDynamicADR:
		return false

	case dev.GetMacSettings().GetAdr().GetDynamic() != nil:
		return true

	case defaults.GetAdr().GetDynamic() != nil:
		return true

	default:
		return true
	}
}

// DeviceUseADR returns if the Network Server uses the ADR algorithm for the end device.
func DeviceUseADR(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings, phy *band.Band) bool {
	if dev.GetMacSettings().GetAdr() != nil {
		return deviceUseADR(dev, defaults, phy)
	}
	return legacyDeviceUseADR(dev, defaults, phy)
}

func deviceShouldAdaptDataRate(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings, phy *band.Band) (adaptDataRate bool, resetDesiredParameters bool, staticSettings *ttnpb.ADRSettings_StaticMode) {
	switch {
	case dev.GetMulticast():
		return false, true, nil

	case dev.GetMacSettings().GetAdr().GetDisabled() != nil:
		return false, true, nil
	case dev.GetMacSettings().GetAdr().GetStatic() != nil:
		return false, false, dev.MacSettings.Adr.GetStatic()

	case defaults.GetAdr().GetDisabled() != nil:
		return false, true, nil
	case defaults.GetAdr().GetStatic() != nil:
		return false, false, defaults.GetAdr().GetStatic()

	case !phy.SupportsDynamicADR:
		return false, true, nil

	case dev.GetMacSettings().GetAdr().GetDynamic() != nil:
		return true, true, nil

	case defaults.GetAdr().GetDynamic() != nil:
		return true, true, nil

	default:
		return false, false, nil
	}
}

// DeviceShouldAdaptDataRate returns if the ADR algorithm should be run for the end device.
func DeviceShouldAdaptDataRate(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings, phy *band.Band) (adaptDataRate bool, resetDesiredParameters bool, staticSettings *ttnpb.ADRSettings_StaticMode) {
	if dev.GetMacSettings().GetAdr() != nil {
		return deviceShouldAdaptDataRate(dev, defaults, phy)
	}
	return legacyDeviceShouldAdaptDataRate(dev, defaults, phy)
}

func deviceADRMargin(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings) float32 {
	switch {
	case dev.GetMacSettings().GetAdr().GetDynamic().GetMargin() != nil:
		return dev.MacSettings.Adr.GetDynamic().Margin.Value

	case defaults.GetAdr().GetDynamic().GetMargin() != nil:
		return defaults.GetAdr().GetDynamic().Margin.Value

	default:
		return DefaultADRMargin
	}
}

// DeviceADRMargin returns the margin to be used by the ADR algorithm.
func DeviceADRMargin(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings) float32 {
	if dev.GetMacSettings().GetAdr() != nil {
		return deviceADRMargin(dev, defaults)
	}
	return legacyDeviceADRMargin(dev, defaults)
}

func adrLossRate(ups ...*ttnpb.MACState_UplinkMessage) float32 {
	if len(ups) < 2 {
		return 0
	}

	min := ups[0].GetPayload().GetMacPayload().GetFullFCnt()
	lastFCnt := min
	var lost uint32
	for i, up := range ups[1:] {
		fCnt := up.GetPayload().GetMacPayload().GetFullFCnt()
		switch {
		case fCnt < lastFCnt:
			return adrLossRate(ups[1+i:]...)
		case fCnt >= lastFCnt+1:
			lost += fCnt - lastFCnt - 1
		}
		lastFCnt = fCnt
	}
	return float32(lost) / float32(1+lastFCnt-min)
}

func maxSNRFromMetadata(mds ...*ttnpb.MACState_UplinkMessage_RxMetadata) (float32, bool) {
	if len(mds) == 0 {
		return 0, false
	}
	maxSNR := mds[0].Snr
	for _, md := range mds[1:] {
		if md.Snr > maxSNR {
			maxSNR = md.Snr
		}
	}
	return maxSNR, true
}

func uplinkMetadata(ups ...*ttnpb.MACState_UplinkMessage) []*ttnpb.MACState_UplinkMessage_RxMetadata {
	mds := make([]*ttnpb.MACState_UplinkMessage_RxMetadata, 0, len(ups))
	for _, up := range ups {
		mds = append(mds, up.RxMetadata...)
	}
	return mds
}

func txPowerStep(phy *band.Band, from, to uint8) float32 {
	max := phy.MaxTxPowerIndex()
	if from > max {
		from = max
	}
	if to > max {
		to = max
	}
	return phy.TxOffset[from] - phy.TxOffset[to]
}

func clampDataRateRange(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings, minDataRateIndex, maxDataRateIndex ttnpb.DataRateIndex) (min, max ttnpb.DataRateIndex) {
	clamp := func(dynamicSettings *ttnpb.ADRSettings_DynamicMode) (min, max ttnpb.DataRateIndex) {
		min, max = minDataRateIndex, maxDataRateIndex
		minSetting, maxSetting := dynamicSettings.MinDataRateIndex, dynamicSettings.MaxDataRateIndex
		if minSetting != nil && minSetting.Value > minDataRateIndex {
			min = minSetting.Value
		}
		if maxSetting != nil && maxSetting.Value < maxDataRateIndex {
			max = maxSetting.Value
		}
		return min, max
	}
	switch {
	case dev.GetMacSettings().GetAdr().GetDynamic() != nil:
		return clamp(dev.MacSettings.Adr.GetDynamic())

	case defaults.GetAdr().GetDynamic() != nil:
		return clamp(defaults.GetAdr().GetDynamic())

	default:
		return minDataRateIndex, maxDataRateIndex
	}
}

func clampTxPowerRange(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings, minTxPowerIndex, maxTxPowerIndex uint8) (min, max uint8) {
	clamp := func(dynamicSettings *ttnpb.ADRSettings_DynamicMode) (min, max uint8) {
		min, max = minTxPowerIndex, maxTxPowerIndex
		minSetting, maxSetting := dynamicSettings.MinTxPowerIndex, dynamicSettings.MaxTxPowerIndex
		if minSetting != nil && uint8(minSetting.Value) > minTxPowerIndex {
			min = uint8(minSetting.Value)
		}
		if maxSetting != nil && uint8(maxSetting.Value) < maxTxPowerIndex {
			max = uint8(maxSetting.Value)
		}
		return min, max
	}
	switch {
	case dev.GetMacSettings().GetAdr().GetDynamic() != nil:
		return clamp(dev.MacSettings.Adr.GetDynamic())

	case defaults.GetAdr().GetDynamic() != nil:
		return clamp(defaults.GetAdr().GetDynamic())

	default:
		return minTxPowerIndex, maxTxPowerIndex
	}
}

func clampNbTrans(dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings, nbTrans uint32) uint32 {
	clamp := func(dynamicSettings *ttnpb.ADRSettings_DynamicMode) uint32 {
		nbTrans := nbTrans
		minSetting, maxSetting := dynamicSettings.MinNbTrans, dynamicSettings.MaxNbTrans
		if minSetting != nil && minSetting.Value > nbTrans {
			nbTrans = minSetting.Value
		}
		if maxSetting != nil && maxSetting.Value < nbTrans {
			nbTrans = maxSetting.Value
		}
		return nbTrans
	}
	switch {
	case dev.GetMacSettings().GetAdr().GetDynamic() != nil:
		return clamp(dev.MacSettings.Adr.GetDynamic())

	case defaults.GetAdr().GetDynamic() != nil:
		return clamp(defaults.GetAdr().GetDynamic())

	default:
		return nbTrans
	}
}

// AdaptDataRate adapts the end device desired ADR parameters based on previous transmissions and device settings.
func AdaptDataRate(ctx context.Context, dev *ttnpb.EndDevice, phy *band.Band, defaults *ttnpb.MACSettings) error {
	macState := dev.MacState
	if macState == nil {
		return nil
	}

	currentParameters, desiredParameters := macState.CurrentParameters, macState.DesiredParameters
	currentDataRateIndex := currentParameters.AdrDataRateIndex

	adrUplinks := func() []*ttnpb.MACState_UplinkMessage {
		for i := len(macState.RecentUplinks) - 1; i >= 0; i-- {
			up := macState.RecentUplinks[i]
			drIdx, _, ok := phy.FindUplinkDataRate(up.Settings.DataRate)
			if !ok {
				continue
			}
			switch {
			case up.Payload.MHdr.MType != ttnpb.MType_UNCONFIRMED_UP && up.Payload.MHdr.MType != ttnpb.MType_CONFIRMED_UP,
				macState.LastAdrChangeFCntUp != 0 && up.Payload.GetMacPayload().FullFCnt <= macState.LastAdrChangeFCntUp,
				drIdx != currentDataRateIndex:
				return macState.RecentUplinks[i+1:]
			}
		}
		return macState.RecentUplinks
	}()
	if len(adrUplinks) == 0 {
		return nil
	}

	minDataRateIndex, maxDataRateIndex, ok := channelDataRateRange(currentParameters.Channels...)
	if !ok {
		return internal.ErrCorruptedMACState.
			WithCause(internal.ErrChannelDataRateRange)
	}
	minDataRateIndex, maxDataRateIndex = clampDataRateRange(dev, defaults, minDataRateIndex, maxDataRateIndex)
	if minDataRateIndex > maxDataRateIndex {
		log.FromContext(ctx).Debug("No common data rate index range available, avoid ADR")
		return nil
	}
	if maxDataRateIndex > phy.MaxADRDataRateIndex {
		maxDataRateIndex = phy.MaxADRDataRateIndex
	}
	rejectedDataRateIndexes := make(map[ttnpb.DataRateIndex]struct{}, len(macState.RejectedAdrDataRateIndexes))
	for _, idx := range macState.RejectedAdrDataRateIndexes {
		rejectedDataRateIndexes[idx] = struct{}{}
	}
	_, ok = rejectedDataRateIndexes[minDataRateIndex]
	for ok && minDataRateIndex <= maxDataRateIndex {
		minDataRateIndex++
		_, ok = rejectedDataRateIndexes[minDataRateIndex]
	}
	_, ok = rejectedDataRateIndexes[maxDataRateIndex]
	for ok && maxDataRateIndex >= minDataRateIndex {
		maxDataRateIndex--
		_, ok = rejectedDataRateIndexes[maxDataRateIndex]
	}
	if minDataRateIndex > maxDataRateIndex {
		log.FromContext(ctx).Debug(
			"Device has rejected all possible data rate values given the channels enabled, avoid ADR",
		)
		return nil
	}
	if currentDataRateIndex > minDataRateIndex {
		minDataRateIndex = currentDataRateIndex
	}

	minTxPowerIndex := uint8(0)
	maxTxPowerIndex := phy.MaxTxPowerIndex()
	minTxPowerIndex, maxTxPowerIndex = clampTxPowerRange(dev, defaults, minTxPowerIndex, maxTxPowerIndex)
	if minTxPowerIndex > maxTxPowerIndex {
		log.FromContext(ctx).Debug("No common TX power index range available, avoid ADR")
		return nil
	}
	rejectedTxPowerIndexes := make(map[uint8]struct{}, len(macState.RejectedAdrTxPowerIndexes))
	for _, idx := range macState.RejectedAdrTxPowerIndexes {
		rejectedTxPowerIndexes[uint8(idx)] = struct{}{}
	}
	_, ok = rejectedTxPowerIndexes[minTxPowerIndex]
	for ok && minTxPowerIndex <= maxTxPowerIndex {
		minTxPowerIndex++
		_, ok = rejectedTxPowerIndexes[minTxPowerIndex]
	}
	_, ok = rejectedTxPowerIndexes[maxTxPowerIndex]
	for ok && maxTxPowerIndex >= minTxPowerIndex {
		maxTxPowerIndex--
		_, ok = rejectedTxPowerIndexes[maxTxPowerIndex]
	}
	if minTxPowerIndex > maxTxPowerIndex {
		log.FromContext(ctx).Debug("Device has rejected all possible TX output power index values, avoid ADR")
		return nil
	}

	maxSNR, ok := maxSNRFromMetadata(uplinkMetadata(adrUplinks...)...)
	if !ok {
		log.FromContext(ctx).Debug("Failed to determine max SNR, avoid ADR")
		return nil
	}

	// The link margin indicates how much stronger the signal (SNR) is than the
	// minimum (floor) that we need to demodulate the signal. We subtract a
	// configurable margin, and an extra safety margin if we're afraid that we
	// don't have enough data for our decision.
	var margin float32
	// NOTE: We currently assume that the uplink's SF and BW correspond to currentDataRateIndex.
	if dr := internal.LastUplink(adrUplinks...).Settings.DataRate.GetLora(); dr != nil {
		var ok bool
		df, ok := demodulationFloor[dr.SpreadingFactor][dr.Bandwidth]
		if !ok {
			return internal.ErrInvalidDataRate.New()
		}
		margin = maxSNR - df - DeviceADRMargin(dev, defaults)
	}
	if len(adrUplinks) < OptimalADRUplinkCount {
		margin -= safetyMargin
	}

	// NOTE: Network Server may only increase the data rate index of the device.
	// NOTE(2): TX output power is reset whenever data rate is increased.
	desiredParameters.AdrDataRateIndex = currentDataRateIndex
	desiredParameters.AdrTxPowerIndex = currentParameters.AdrTxPowerIndex
	if currentDataRateIndex < minDataRateIndex {
		margin -= float32(minDataRateIndex-currentDataRateIndex) * drStep
		desiredParameters.AdrDataRateIndex = minDataRateIndex
		desiredParameters.AdrTxPowerIndex = 0
	}
	maxMarginSteps := float32(maxDataRateIndex - desiredParameters.AdrDataRateIndex)
	marginSteps := (margin - txPowerStep(phy, 0, minTxPowerIndex)) / drStep
	if marginSteps >= 0 && marginSteps < maxMarginSteps {
		maxDataRateIndex = desiredParameters.AdrDataRateIndex + ttnpb.DataRateIndex(marginSteps)
	} else if marginSteps < 0 {
		maxDataRateIndex = minDataRateIndex
	}
	for drIdx := maxDataRateIndex; drIdx > minDataRateIndex; drIdx-- {
		if _, ok := rejectedDataRateIndexes[drIdx]; ok {
			continue
		}
		margin -= float32(drIdx-desiredParameters.AdrDataRateIndex) * drStep
		desiredParameters.AdrDataRateIndex = drIdx
		desiredParameters.AdrTxPowerIndex = 0
		break
	}

	if desiredParameters.AdrTxPowerIndex < uint32(minTxPowerIndex) {
		margin -= txPowerStep(phy, uint8(desiredParameters.AdrTxPowerIndex), minTxPowerIndex)
		desiredParameters.AdrTxPowerIndex = uint32(minTxPowerIndex)
	}
	if desiredParameters.AdrTxPowerIndex > uint32(maxTxPowerIndex) {
		margin += txPowerStep(phy, maxTxPowerIndex, uint8(desiredParameters.AdrTxPowerIndex))
		desiredParameters.AdrTxPowerIndex = uint32(maxTxPowerIndex)
	}
	// If we still have margin left, we decrease the TX output power (increase the index).
	for txPowerIdx := maxTxPowerIndex; txPowerIdx > minTxPowerIndex; txPowerIdx-- {
		diff := txPowerStep(phy, uint8(desiredParameters.AdrTxPowerIndex), txPowerIdx)
		if _, ok := rejectedTxPowerIndexes[txPowerIdx]; ok || diff > margin {
			continue
		}
		margin -= diff // nolint:ineffassign
		desiredParameters.AdrTxPowerIndex = uint32(txPowerIdx)
		break
	}

	nbTrans := clampNbTrans(dev, defaults, currentParameters.AdrNbTrans)
	if len(adrUplinks) >= OptimalADRUplinkCount/2 {
		switch r := adrLossRate(adrUplinks...); {
		case r < 0.05:
			nbTrans = 1 + nbTrans/3
		case r < 0.10:
		case r < 0.30:
			nbTrans = 2 + nbTrans/2
		default:
			nbTrans = maxNbTrans
		}
	}
	desiredParameters.AdrNbTrans = clampNbTrans(dev, defaults, nbTrans)

	return nil
}
