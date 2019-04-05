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

// LoRaWAN 1.0.3rA -> 1.0.2rB downgrades

func disableCFList1_0_2(b Band) Band {
	b.ImplementsCFList = false
	return b
}

// LoRaWAN 1.0.3rA -> 1.0.2rB downgrades

func disableChMaskCntl51_0_2(b Band) Band {
	b.GenerateChMasks = makeGenerateChMask72(false)
	return b
}

// LoRaWAN 1.0.2rB -> 1.0.2rA downgrades

func auDataRates1_0_2(b Band) Band {
	for i := 0; i < 4; i++ {
		b.DataRates[i] = b.DataRates[i+2]
	}
	b.DataRates[5] = DataRate{}
	b.DataRates[6] = DataRate{}
	return b
}

func usBeacon1_0_2(b Band) Band {
	b.Beacon.DataRateIndex = 3
	return b
}
