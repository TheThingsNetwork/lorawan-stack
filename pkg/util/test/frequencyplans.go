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

package test

import (
	"go.thethings.network/lorawan-stack/pkg/fetch"
)

const (
	frequencyPlansDescription = `- id: EU_863_870
  name: Europe 863-870 MHz
  base-frequency: 868
  file: EU_863_870.yml
- id: KR_920_923
  name: Korea 920-923 MHz
  base-frequency: 915
  file: KR_920_923.yml
- id: EXAMPLE
  name: Example 866.1 MHz
  base-frequency: 868
  file: EXAMPLE.yml`

	// EUFrequencyPlanID is a European frequency plan for testing.
	EUFrequencyPlanID = "EU_863_870"
	euFrequencyPlan   = `band-id: EU_863_870
uplink-channels:
- frequency: 867100000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 867300000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 867500000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 867700000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 867900000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 868100000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 868300000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 868500000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 869525000
  min-data-rate: 0
  max-data-rate: 5
downlink-channels:
- frequency: 867100000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 867300000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 867500000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 867700000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 867900000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 868100000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 868300000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 868500000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 869525000
  min-data-rate: 0
  max-data-rate: 5
lora-standard-channel:
  frequency: 863000000
  data-rate: 6
fsk-channel:
  frequency: 868800000
  data-rate: 7
`

	// KRFrequencyPlanID is a Korean frequency plan for testing.
	KRFrequencyPlanID = "KR_920_923"
	krFrequencyPlan   = `band-id: KR_920_923
uplink-channels:
- frequency: 922100000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 922300000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 922500000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 922700000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 922900000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 923100000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 923300000
  min-data-rate: 0
  max-data-rate: 5
downlink-channels:
- frequency: 922100000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 922300000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 922500000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 922700000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 922900000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 923100000
  min-data-rate: 0
  max-data-rate: 5
- frequency: 923300000
  min-data-rate: 0
  max-data-rate: 5
lbt:
  rssi-target: -80
  scan-time: 128`

	// ExampleFrequencyPlanID is an example frequency plan.
	ExampleFrequencyPlanID = "EXAMPLE"
	exampleFrequencyPlan   = `band-id: EU_863_870
uplink-channels:
- frequency: 868100000
  min-data-rate: 0
  max-data-rate: 5
  radio: 0
downlink-channels:
- frequency: 868100000
  min-data-rate: 0
  max-data-rate: 5
  radio: 0
lora-standard-channel:
  frequency: 863000000
  data-rate: 6
fsk-channel:
  frequency: 868800000
  data-rate: 7
time-off-air:
  fraction: 0.1
  duration: 1s
dwell-time:
  uplinks: true
  downlinks: true
  duration: 1s
lbt:
  rssi-offset: 0
  rssi-target: -80
  scan-time: 128000
radios:
- enable: true
  chip-type: SX1257
  frequency: 867500000
  rssi-offset: -166
  tx:
    min-frequency: 863000000
    max-frequency: 867000000
    notch-frequency: 129000
clock-source: 0
ping-slot:
  frequency: 869525000
  min-data-rate: 0
  max-data-rate: 5
  radio: 0
ping-slot-default-data-rate: 3
rx2-channel:
  frequency: 869525000
  min-data-rate: 0
  max-data-rate: 5
  radio: 0
rx2-default-data-rate: 0
max-eirp: 27`
)

// FrequencyPlansFetcher fetches frequency plans from memory.
var FrequencyPlansFetcher = fetch.NewMemFetcher(map[string][]byte{
	"frequency-plans.yml": []byte(frequencyPlansDescription),
	"EU_863_870.yml":      []byte(euFrequencyPlan),
	"KR_920_923.yml":      []byte(krFrequencyPlan),
	"EXAMPLE.yml":         []byte(exampleFrequencyPlan),
})
