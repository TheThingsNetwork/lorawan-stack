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

func txPowerStep(phy *band.Band, from, to uint32) float32 {
	max := uint32(phy.MaxTxPowerIndex())
	if from > max {
		from = max
	}
	if to > max {
		to = max
	}
	return phy.TxOffset[from] - phy.TxOffset[to]
}

func clampDataRateRange(
	dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings, minDataRateIndex, maxDataRateIndex ttnpb.DataRateIndex,
) (min, max ttnpb.DataRateIndex) {
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

func clampTxPowerRange(
	dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings, minTxPowerIndex, maxTxPowerIndex uint32,
) (min, max uint32) {
	clamp := func(dynamicSettings *ttnpb.ADRSettings_DynamicMode) (min, max uint32) {
		min, max = minTxPowerIndex, maxTxPowerIndex
		minSetting, maxSetting := dynamicSettings.MinTxPowerIndex, dynamicSettings.MaxTxPowerIndex
		if minSetting != nil && minSetting.Value > minTxPowerIndex {
			min = minSetting.Value
		}
		if maxSetting != nil && maxSetting.Value < maxTxPowerIndex {
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

func adrUplinks(macState *ttnpb.MACState, phy *band.Band) []*ttnpb.MACState_UplinkMessage {
	currentDataRateIndex := macState.CurrentParameters.AdrDataRateIndex
	for i := len(macState.RecentUplinks) - 1; i >= 0; i-- {
		up := macState.RecentUplinks[i]
		drIdx, _, ok := phy.FindUplinkDataRate(up.Settings.DataRate)
		if !ok {
			continue
		}
		switch {
		case up.Payload.MHdr.MType != ttnpb.MType_UNCONFIRMED_UP && up.Payload.MHdr.MType != ttnpb.MType_CONFIRMED_UP,
			macState.LastAdrChangeFCntUp != 0 && up.Payload.GetMacPayload().FullFCnt < macState.LastAdrChangeFCntUp,
			drIdx != currentDataRateIndex:
			return macState.RecentUplinks[i+1:]
		}
	}
	return macState.RecentUplinks
}

func indicesMap[T comparable](indices ...T) map[T]struct{} {
	if len(indices) == 0 {
		return nil
	}
	m := make(map[T]struct{}, len(indices))
	for _, idx := range indices {
		m[idx] = struct{}{}
	}
	return m
}

func adrDataRateRange(
	ctx context.Context, dev *ttnpb.EndDevice, phy *band.Band, defaults *ttnpb.MACSettings,
) (min, max ttnpb.DataRateIndex, rejected map[ttnpb.DataRateIndex]struct{}, ok bool, err error) {
	macState := dev.MacState
	min, max, _, ok = channelDataRateRange(
		macState.DesiredParameters.Channels...,
	)
	if !ok {
		return 0, 0, nil, false, internal.ErrCorruptedMACState.
			WithCause(internal.ErrChannelDataRateRange)
	}
	min, max = clampDataRateRange(dev, defaults, min, max)
	if min > max {
		log.FromContext(ctx).Debug("No common data rate index range available, avoid ADR")
		return 0, 0, nil, false, nil
	}
	if max > phy.MaxADRDataRateIndex {
		max = phy.MaxADRDataRateIndex
	}
	rejected = indicesMap(macState.RejectedAdrDataRateIndexes...)
	_, ok = rejected[min]
	for ok && min <= max {
		min++
		_, ok = rejected[min]
	}
	_, ok = rejected[max]
	for ok && max >= min {
		max--
		_, ok = rejected[max]
	}
	if min > max {
		log.FromContext(ctx).Debug(
			"Device has rejected all possible data rate values given the channels enabled, avoid ADR",
		)
		return 0, 0, nil, false, nil
	}
	if currentDataRateIndex := macState.CurrentParameters.AdrDataRateIndex; currentDataRateIndex > min {
		min = currentDataRateIndex
	}
	return min, max, rejected, true, nil
}

func adrTxPowerRange(
	ctx context.Context, dev *ttnpb.EndDevice, phy *band.Band, defaults *ttnpb.MACSettings,
) (min, max uint32, rejected map[uint32]struct{}, ok bool) {
	min, max = uint32(0), uint32(phy.MaxTxPowerIndex())
	min, max = clampTxPowerRange(dev, defaults, min, max)
	if min > max {
		log.FromContext(ctx).Debug("No common TX power index range available, avoid ADR")
		return 0, 0, nil, false
	}
	rejected = indicesMap(dev.MacState.RejectedAdrTxPowerIndexes...)
	_, ok = rejected[min]
	for ok && min <= max {
		min++
		_, ok = rejected[min]
	}
	_, ok = rejected[max]
	for ok && max >= min {
		max--
		_, ok = rejected[max]
	}
	if min > max {
		log.FromContext(ctx).Debug("Device has rejected all possible TX output power index values, avoid ADR")
		return 0, 0, nil, false
	}
	return min, max, rejected, true
}

func adrMargin(
	ctx context.Context, dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings, adrUplinks ...*ttnpb.MACState_UplinkMessage,
) (margin float32, optimal bool, ok bool, err error) {
	maxSNR, ok := maxSNRFromMetadata(uplinkMetadata(adrUplinks...)...)
	if !ok {
		log.FromContext(ctx).Debug("Failed to determine max SNR, avoid ADR")
		return 0, false, false, nil
	}
	// The link margin indicates how much stronger the signal (SNR) is than the
	// minimum (floor) that we need to demodulate the signal. We subtract a
	// configurable margin.
	// NOTE: We currently assume that the uplink's SF and BW correspond to currentDataRateIndex.
	if dr := internal.LastUplink(adrUplinks...).Settings.DataRate.GetLora(); dr != nil {
		var ok bool
		df, ok := demodulationFloor[dr.SpreadingFactor][dr.Bandwidth]
		if !ok {
			return 0, false, false, internal.ErrInvalidDataRate.New()
		}
		margin = maxSNR - df - DeviceADRMargin(dev, defaults)
	}
	// We subtract an extra safety margin if we're afraid that we  don't have enough data
	// for our decision.
	optimal = len(adrUplinks) >= OptimalADRUplinkCount
	if !optimal {
		margin -= safetyMargin
	}
	return margin, optimal, true, nil
}

func adrAdaptDataRate(
	macState *ttnpb.MACState,
	phy *band.Band,
	minDataRateIndex, maxDataRateIndex ttnpb.DataRateIndex,
	rejectedDataRateIndices map[ttnpb.DataRateIndex]struct{},
	minTxPowerIndex uint32,
	margin float32,
) float32 {
	currentParameters, desiredParameters := macState.CurrentParameters, macState.DesiredParameters
	currentDataRateIndex := currentParameters.AdrDataRateIndex
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
		if _, ok := rejectedDataRateIndices[drIdx]; ok {
			continue
		}
		margin -= float32(drIdx-desiredParameters.AdrDataRateIndex) * drStep
		desiredParameters.AdrDataRateIndex = drIdx
		desiredParameters.AdrTxPowerIndex = 0
		break
	}
	return margin
}

func adrAdaptTxPowerIndex(
	macState *ttnpb.MACState,
	phy *band.Band,
	min, max uint32,
	rejected map[uint32]struct{},
	margin float32,
	optimal bool,
) float32 {
	desiredParameters := macState.DesiredParameters
	if desiredParameters.AdrTxPowerIndex < min {
		margin -= txPowerStep(phy, desiredParameters.AdrTxPowerIndex, min)
		desiredParameters.AdrTxPowerIndex = min
	}
	if desiredParameters.AdrTxPowerIndex > max {
		margin += txPowerStep(phy, max, desiredParameters.AdrTxPowerIndex)
		desiredParameters.AdrTxPowerIndex = max
	}
	// If we still have margin left, we decrease the TX output power (increase the index).
	// We can also compensate the missing margin by increasing the TX output power (decreasing the index).
	for txPowerIdx := max; txPowerIdx >= min; txPowerIdx-- {
		diff := txPowerStep(phy, desiredParameters.AdrTxPowerIndex, txPowerIdx)
		// As long as we are not at the minimal transmission power index, we skip
		// rejected indices or indices which do not fit in the margin.
		if _, ok := rejected[txPowerIdx]; (ok || diff > margin) && txPowerIdx != min {
			continue
		}
		if !optimal && diff < 0 && -diff <= safetyMargin {
			// If the transmission power is increased by a value lower than the safety margin
			// while the number of observed uplinks is not optimal, we avoid changing the
			// transmission power in order to avoid flip-flopping.
			break
		}
		margin -= diff
		desiredParameters.AdrTxPowerIndex = txPowerIdx
		break
	}
	return margin
}

func adrAdaptNbTrans(
	dev *ttnpb.EndDevice, defaults *ttnpb.MACSettings, adrUplinks []*ttnpb.MACState_UplinkMessage,
) {
	macState := dev.MacState
	currentParameters, desiredParameters := macState.CurrentParameters, macState.DesiredParameters
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
}

func adaptDataRate(ctx context.Context, dev *ttnpb.EndDevice, phy *band.Band, defaults *ttnpb.MACSettings) error {
	macState := dev.MacState
	adrUplinks := adrUplinks(macState, phy)
	if len(adrUplinks) == 0 {
		return nil
	}
	minDataRateIndex, maxDataRateIndex, rejectedDataRateIndices, ok, err := adrDataRateRange(ctx, dev, phy, defaults)
	if err != nil || !ok {
		return err
	}
	minTxPowerIndex, maxTxPowerIndex, rejectedTxPowerIndices, ok := adrTxPowerRange(ctx, dev, phy, defaults)
	if !ok {
		return nil
	}
	margin, optimal, ok, err := adrMargin(ctx, dev, defaults, adrUplinks...)
	if err != nil || !ok {
		return err
	}
	margin = adrAdaptDataRate(
		macState, phy, minDataRateIndex, maxDataRateIndex, rejectedDataRateIndices, minTxPowerIndex, margin,
	)
	margin = adrAdaptTxPowerIndex(
		macState, phy, minTxPowerIndex, maxTxPowerIndex, rejectedTxPowerIndices, margin, optimal,
	)
	_ = margin
	adrAdaptNbTrans(dev, defaults, adrUplinks)
	return nil
}

// AdaptDataRate adapts the end device desired ADR parameters based on previous transmissions and device settings.
func AdaptDataRate(ctx context.Context, dev *ttnpb.EndDevice, phy *band.Band, defaults *ttnpb.MACSettings) error {
	if dev.MacState == nil {
		return nil
	}
	return adaptDataRate(ctx, dev, phy, defaults)
}
