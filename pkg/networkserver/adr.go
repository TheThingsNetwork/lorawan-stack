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
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
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

// safetyMargin is a margin in dB added for ADR parameter calculation if less than 20 uplinks are available.
const safetyMargin = 2.5

// drStep is the difference between 2 consequitive data rates in dB.
const drStep = 2.5

// maxNbTrans is the maximum NbTrans parameter used by the algorithm.
const maxNbTrans = 1

// optimalADRUplinkCount is the amount of uplinks required to ensure optimal results from the ADR algorithm.
const optimalADRUplinkCount = 20

// DefaultADRMargin is the default ADR margin used if not specified in MACSettings of the device or NS-wide defaults.
const DefaultADRMargin = 15

func deviceADRMargin(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) float32 {
	if dev.MACSettings != nil && dev.MACSettings.ADRMargin != nil {
		return dev.MACSettings.ADRMargin.Value
	}
	if defaults.ADRMargin != nil {
		return defaults.ADRMargin.Value
	}
	return DefaultADRMargin
}

func adaptDataRate(dev *ttnpb.EndDevice, fps *frequencyplans.Store, defaults ttnpb.MACSettings) error {
	ups := dev.RecentADRUplinks
	if len(ups) == 0 {
		return nil
	}

	maxSNR := ups[0].RxMetadata[0].SNR
	for _, up := range ups {
		for _, md := range up.RxMetadata {
			if md.SNR > maxSNR {
				maxSNR = md.SNR
			}
		}
	}

	_, phy, err := getDeviceBandVersion(dev, fps)
	if err != nil {
		return err
	}

	up := ups[len(ups)-1]

	dev.MACState.DesiredParameters.ADRDataRateIndex = dev.MACState.CurrentParameters.ADRDataRateIndex
	dev.MACState.DesiredParameters.ADRTxPowerIndex = dev.MACState.CurrentParameters.ADRTxPowerIndex

	// NOTE: We currently assume that the uplink's SF and BW correspond to CurrentParameters.ADRDataRateIndex.
	var df float32
	if dr := up.Settings.DataRate.GetLoRa(); dr != nil {
		var ok bool
		df, ok = demodulationFloor[dr.SpreadingFactor][dr.Bandwidth]
		if !ok {
			return errInvalidDataRate
		}
	}

	// The link margin indicates how much stronger the signal (SNR) is than the
	// minimum (floor) that we need to demodulate the signal. We subtract a
	// configurable margin, and an extra safety margin if we're afraid that we
	// don't have enough data for our decision.
	margin := maxSNR - df - deviceADRMargin(dev, defaults)
	if len(ups) < optimalADRUplinkCount {
		margin -= safetyMargin
	}

	// As long as we have enough margin to increase the data rate, we do that.
	// If we change the DR, we reset the Tx power.
	for int(dev.MACState.DesiredParameters.ADRDataRateIndex) < int(phy.MaxADRDataRateIndex) {
		newMargin := margin - drStep
		if newMargin < 0 {
			break
		}
		margin = newMargin
		dev.MACState.DesiredParameters.ADRDataRateIndex++
		dev.MACState.DesiredParameters.ADRTxPowerIndex = 0
	}

	// If we still have margin left, we decrease the Tx power (increase the index).
	for int(dev.MACState.DesiredParameters.ADRTxPowerIndex) < int(phy.MaxTxPowerIndex) {
		newMargin := margin - (phy.TxOffset[dev.MACState.DesiredParameters.ADRTxPowerIndex] - phy.TxOffset[dev.MACState.DesiredParameters.ADRTxPowerIndex+1])
		if newMargin < 0 {
			break
		}
		margin = newMargin
		dev.MACState.DesiredParameters.ADRTxPowerIndex++
	}

	dev.MACState.DesiredParameters.ADRNbTrans = dev.MACState.CurrentParameters.ADRNbTrans
	if dev.MACState.DesiredParameters.ADRNbTrans > maxNbTrans {
		dev.MACState.DesiredParameters.ADRNbTrans = maxNbTrans
	}

	if len(ups) >= 2 {
		lossRate := float32(up.Payload.GetMACPayload().FHDR.FCnt-ups[0].Payload.GetMACPayload().FHDR.FCnt-uint32(len(ups))) /
			float32(up.Payload.GetMACPayload().FHDR.FCnt-ups[0].Payload.GetMACPayload().FHDR.FCnt)
		switch {
		case lossRate < 0.05:
			dev.MACState.DesiredParameters.ADRNbTrans = 1 + dev.MACState.DesiredParameters.ADRNbTrans/3
		case lossRate < 0.10:
		case lossRate < 0.30:
			dev.MACState.DesiredParameters.ADRNbTrans = 2 + dev.MACState.DesiredParameters.ADRNbTrans/2
		default:
			dev.MACState.DesiredParameters.ADRNbTrans = maxNbTrans
		}
	}
	return nil
}
