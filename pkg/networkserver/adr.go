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

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// TODO: The values for BW250 and BW500 need to be verified
// (https://github.com/TheThingsNetwork/lorawan-stack/issues/21)

var demodulationFloor = map[uint32]map[uint32]float32{
	6: {
		125000: -5,
		250000: -2,
		500000: 1,
	},
	7: {
		125000: -7.5,
		250000: -4.5,
		500000: -1.5,
	},
	8: {
		125000: -10,
		250000: -7,
		500000: -4,
	},
	9: {
		125000: -12.5,
		250000: -9.5,
		500000: -6.5,
	},
	10: {
		125000: -15,
		250000: -12,
		500000: -9,
	},
	11: {
		125000: -17.5,
		250000: -14.5,
		500000: -11.5,
	},
	12: {
		125000: -20,
		250000: -17,
		500000: -24,
	},
}

const (
	// safetyMargin is a margin in dB added for ADR parameter calculation if less than 20 uplinks are available.
	safetyMargin = 2.5

	// drStep is the difference between 2 consequitive data rates in dB.
	drStep = 2.5

	// maxNbTrans is the maximum NbTrans parameter used by the algorithm.
	maxNbTrans = 3

	// optimalADRUplinkCount is the amount of uplinks required to ensure optimal results from the ADR algorithm.
	optimalADRUplinkCount = 20

	// DefaultADRMargin is the default ADR margin used if not specified in MACSettings of the device or NS-wide defaults.
	DefaultADRMargin = 15
)

func deviceADRMargin(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) float32 {
	if dev.MACSettings != nil && dev.MACSettings.ADRMargin != nil {
		return dev.MACSettings.ADRMargin.Value
	}
	if defaults.ADRMargin != nil {
		return defaults.ADRMargin.Value
	}
	return DefaultADRMargin
}

func lossRate(nbTrans uint32, ups ...*ttnpb.UplinkMessage) float32 {
	if len(ups) < 2 {
		return 0
	}

	min := ups[0].Payload.GetMACPayload().FullFCnt
	max := lastUplink(ups...).Payload.GetMACPayload().FullFCnt

	fCnt := min + 1
	var fCntTrans uint32
	var lost uint32
	n := uint32(len(ups))

loop:
	for i := 0; i < len(ups); i++ {
		switch v := ups[i].Payload.GetMACPayload().FullFCnt; v {
		case min:
			continue

		case fCnt:
			fCntTrans++

		default:
			if v < fCnt {
				// FCnt reset encountered
				return lossRate(nbTrans, ups[i:]...)
			}

			if fCntTrans < nbTrans {
				d := nbTrans - fCntTrans
				lost += d
				n += d
			}
			d := (v - fCnt - 1) * nbTrans
			lost += d
			n += d

			if v == max {
				break loop
			}
			fCnt = v
			fCntTrans = 1
		}
	}
	return float32(lost) / float32(n)
}

func maxSNRFromMetadata(mds ...*ttnpb.RxMetadata) (float32, bool) {
	if len(mds) == 0 {
		return 0, false
	}
	maxSNR := mds[0].SNR
	for _, md := range mds[1:] {
		if md.SNR > maxSNR {
			maxSNR = md.SNR
		}
	}
	return maxSNR, true
}

func uplinkMetadata(ups ...*ttnpb.UplinkMessage) []*ttnpb.RxMetadata {
	mds := make([]*ttnpb.RxMetadata, 0, len(ups))
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

func adaptDataRate(ctx context.Context, dev *ttnpb.EndDevice, phy *band.Band, defaults ttnpb.MACSettings) error {
	if len(dev.RecentADRUplinks) == 0 {
		return nil
	}

	minDataRateIndex, maxDataRateIndex, ok := channelDataRateRange(dev.MACState.CurrentParameters.Channels...)
	if !ok {
		return errCorruptedMACState
	}
	if maxDataRateIndex > phy.MaxADRDataRateIndex {
		maxDataRateIndex = phy.MaxADRDataRateIndex
	}
	rejectedDataRateIndexes := make(map[ttnpb.DataRateIndex]struct{}, len(dev.MACState.RejectedADRDataRateIndexes))
	for _, idx := range dev.MACState.RejectedADRDataRateIndexes {
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
		log.FromContext(ctx).Debug("Device has rejected all possible data rate values given the channels enabled, avoid ADR.")
		return nil
	}
	if dev.MACState.CurrentParameters.ADRDataRateIndex > minDataRateIndex {
		minDataRateIndex = dev.MACState.CurrentParameters.ADRDataRateIndex
	}

	minTxPowerIndex := uint8(0)
	maxTxPowerIndex := phy.MaxTxPowerIndex()
	rejectedTxPowerIndexes := make(map[uint8]struct{}, len(dev.MACState.RejectedADRTxPowerIndexes))
	for _, idx := range dev.MACState.RejectedADRTxPowerIndexes {
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
		log.FromContext(ctx).Debug("Device has rejected all possible TX output power index values, avoid ADR.")
		return nil
	}

	maxSNR, ok := maxSNRFromMetadata(uplinkMetadata(dev.RecentADRUplinks...)...)
	if !ok {
		log.FromContext(ctx).Debug("Failed to determine max SNR, avoid ADR.")
		return nil
	}
	up := lastUplink(dev.RecentADRUplinks...)

	// The link margin indicates how much stronger the signal (SNR) is than the
	// minimum (floor) that we need to demodulate the signal. We subtract a
	// configurable margin, and an extra safety margin if we're afraid that we
	// don't have enough data for our decision.
	var margin float32
	// NOTE: We currently assume that the uplink's SF and BW correspond to CurrentParameters.ADRDataRateIndex.
	if dr := up.Settings.DataRate.GetLoRa(); dr != nil {
		var ok bool
		df, ok := demodulationFloor[dr.SpreadingFactor][dr.Bandwidth]
		if !ok {
			return errInvalidDataRate.New()
		}
		margin = maxSNR - df - deviceADRMargin(dev, defaults)
	}
	if len(dev.RecentADRUplinks) < optimalADRUplinkCount {
		margin -= safetyMargin
	}

	// NOTE: Network Server may only increase the data rate index of the device.
	// NOTE(2): TX output power is reset whenever data rate is increased.
	dev.MACState.DesiredParameters.ADRDataRateIndex = dev.MACState.CurrentParameters.ADRDataRateIndex
	dev.MACState.DesiredParameters.ADRTxPowerIndex = dev.MACState.CurrentParameters.ADRTxPowerIndex
	if dev.MACState.CurrentParameters.ADRDataRateIndex < minDataRateIndex {
		margin -= float32(minDataRateIndex-dev.MACState.CurrentParameters.ADRDataRateIndex) * drStep
		dev.MACState.DesiredParameters.ADRDataRateIndex = minDataRateIndex
		dev.MACState.DesiredParameters.ADRTxPowerIndex = 0
	}
	if marginSteps := (margin - txPowerStep(phy, 0, minTxPowerIndex)) / drStep; marginSteps >= 0 && marginSteps < float32(maxDataRateIndex-dev.MACState.DesiredParameters.ADRDataRateIndex) {
		maxDataRateIndex = dev.MACState.DesiredParameters.ADRDataRateIndex + ttnpb.DataRateIndex(marginSteps)
	}
	for drIdx := maxDataRateIndex; drIdx > minDataRateIndex; drIdx-- {
		if _, ok := rejectedDataRateIndexes[drIdx]; ok {
			continue
		}
		margin -= float32(drIdx-dev.MACState.DesiredParameters.ADRDataRateIndex) * drStep
		dev.MACState.DesiredParameters.ADRDataRateIndex = drIdx
		dev.MACState.DesiredParameters.ADRTxPowerIndex = 0
		break
	}

	if dev.MACState.DesiredParameters.ADRTxPowerIndex < uint32(minTxPowerIndex) {
		margin -= txPowerStep(phy, uint8(dev.MACState.DesiredParameters.ADRTxPowerIndex), minTxPowerIndex)
		dev.MACState.DesiredParameters.ADRTxPowerIndex = uint32(minTxPowerIndex)
	}
	if dev.MACState.DesiredParameters.ADRTxPowerIndex > uint32(maxTxPowerIndex) {
		margin += txPowerStep(phy, maxTxPowerIndex, uint8(dev.MACState.DesiredParameters.ADRTxPowerIndex))
		dev.MACState.DesiredParameters.ADRTxPowerIndex = uint32(maxTxPowerIndex)
	}
	// If we still have margin left, we decrease the TX output power (increase the index).
	for txPowerIdx := maxTxPowerIndex; txPowerIdx > minTxPowerIndex; txPowerIdx-- {
		diff := txPowerStep(phy, uint8(dev.MACState.DesiredParameters.ADRTxPowerIndex), txPowerIdx)
		if _, ok := rejectedTxPowerIndexes[txPowerIdx]; ok || diff > margin {
			continue
		}
		margin -= diff
		dev.MACState.DesiredParameters.ADRTxPowerIndex = uint32(txPowerIdx)
		break
	}

	dev.MACState.DesiredParameters.ADRNbTrans = dev.MACState.CurrentParameters.ADRNbTrans
	if dev.MACState.DesiredParameters.ADRNbTrans > maxNbTrans {
		dev.MACState.DesiredParameters.ADRNbTrans = maxNbTrans
	}
	if len(dev.RecentADRUplinks) >= 2 {
		switch r := lossRate(dev.MACState.CurrentParameters.ADRNbTrans, dev.RecentADRUplinks...); {
		case r < 0.05:
			dev.MACState.DesiredParameters.ADRNbTrans = 1 + dev.MACState.DesiredParameters.ADRNbTrans/3
		case r < 0.10:
		case r < 0.30:
			dev.MACState.DesiredParameters.ADRNbTrans = 2 + dev.MACState.DesiredParameters.ADRNbTrans/2
		default:
			dev.MACState.DesiredParameters.ADRNbTrans = maxNbTrans
		}
	}
	return nil
}
