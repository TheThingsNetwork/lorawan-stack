// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import { defineMessages } from 'react-intl'

const messages = defineMessages({
  // Shared messages.
  repositoryTabTitle: 'From The LoRaWAN Device Repository',
  manualTabTitle: 'Manually',
  classCapabilities: 'Additional LoRaWAN class capabilities',
  submitTitle: 'Register end device',
  activationModeNone: 'Do not configure activation',
  afterRegistration: 'After registration',
  singleRegistration: 'View registered end device',
  multipleRegistration: 'Register another end device of this type',
  createSuccess: 'End device registered',
  deviceIdDescription: 'This value is automatically prefilled using the DevEUI',
  // Manual messages.
  basicTitle: 'Basic settings',
  basicDescription: "End device ID's, Name and Description",
  basicDetails: 'Defines general settings of an end device',
  beaconFrequency: 'Beacon frequency',
  networkTitle: 'Network layer settings',
  networkDescription: 'Frequency plan, regional parameters, end device class and session keys.',
  appTitle: 'Application layer settings',
  appDescription: 'Application session key to encrypt/decrypt LoRaWAN payload.',
  joinTitle: 'Join settings',
  joinDescription: 'Root keys, NetID and kek labels.',
  rx1DataRateOffsetTitle: 'Rx1 data rate offset',
  rx1DelayTitle: 'Rx1 delay',
  factoryPresetFreqTitle: 'Factory preset frequencies',
  freqAdd: 'Add Frequency',
  frequencyPlaceholder: 'e.g. 869525000 for 869,525 MHz',
  pingSlotPeriodicityTitle: 'Ping slot periodicity',
  pingSlotPeriodicityValue: '{count, plural, one {every second} other {every {count} seconds}}',
  pingSlotDataRateTitle: 'Ping slot data rate',
  pingSlotFrequencyTitle: 'Ping slot frequency',
  rx2DataRateIndexTitle: 'Rx2 data rate',
  rx2FrequencyTitle: 'Rx2 frequency',
  classBTimeout: 'Class B timeout',
  classCTimeout: 'Class C timeout',
  defaultNetworksSettings: "Use network's default MAC settings",
  clusterSettings: 'Cluster settings',
  networkDefaults: 'Network defaults',
  // Device repository messages.
  otherOption: 'Other…',
  typeToSearch: 'Type to search…',
  unknownHwOption: 'Unknown ver.',
})

export default messages
