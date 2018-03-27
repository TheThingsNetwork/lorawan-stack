// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
