// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

// LoRaWAN 1.1 -> 1.0.2 downgrades

func disableCFList_1_0_2(b Band) Band {
	b.ImplementsCFList = false
	return b
}

// LoRaWAN 1.0.2 -> 1.0.1 downgrades

func usBeacon_1_0_1(b Band) Band {
	b.Beacon.DataRateIndex = 3
	return b
}

func auDataRates_1_0_1(b Band) Band {
	b.DataRates[5] = DataRate{}
	b.DataRates[6] = DataRate{}
	return b
}
