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

package band

import "go.thethings.network/lorawan-stack/pkg/ttnpb"

func disableCFList(b Band) Band {
	b.ImplementsCFList = false
	return b
}

func disableChMaskCntl5(b Band) Band {
	b.GenerateChMasks = makeGenerateChMask72(false)
	return b
}

func disableTxParamSetupReq(b Band) Band {
	b.TxParamSetupReqSupport = false
	return b
}

func enableTxParamSetupReq(b Band) Band {
	b.TxParamSetupReqSupport = true
	return b
}

func makeSetMaxTxPowerIndexFunc(idx uint8) func(Band) Band {
	return func(b Band) Band {
		n := idx + 1
		b.TxOffset = append(make([]float32, 0, n), b.TxOffset[:n]...)
		return b
	}
}

func makeSetBeaconDataRateIndex(idx ttnpb.DataRateIndex) func(Band) Band {
	return func(b Band) Band {
		b.Beacon.DataRateIndex = idx
		return b
	}
}

func makeAddTxPowerFunc(offset float32) func(Band) Band {
	return func(b Band) Band {
		b.TxOffset = append(append(make([]float32, 0, len(b.TxOffset)+1), b.TxOffset...), offset)
		return b
	}
}
