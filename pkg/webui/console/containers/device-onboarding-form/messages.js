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

import { defineMessages } from 'react-intl'

export default defineMessages({
  endDeviceType: 'End device type',
  provisioningTitle: 'Provisioning information',
  inputMethod: 'Input Method',
  inputMethodDeviceRepo: 'Select the end device in the LoRaWAN Device Repository',
  inputMethodManual: 'Enter end device specifics manually',
  continueManual: 'To continue, please enter versions and frequency plan information',
  continueJoinEui:
    'To continue, please enter the JoinEUI of the end device so we can determine onboarding options',
  changeDeviceType:
    'Are you sure you want to change the input method? Your current form progress will be lost.',
  changeDeviceTypeButton: 'Change input method',
  confirmedRegistration: 'This end device can be registered on the network',
  confirmedClaiming: 'This end device can be claimed',
  // Shared messages.
  classCapabilities: 'Additional LoRaWAN class capabilities',
  submitTitle: 'Register end device',
  afterRegistration: 'After registration',
  singleRegistration: 'View registered end device',
  multipleRegistration: 'Register another end device of this type',
  createSuccess: 'End device registered',
  deviceIdDescription: 'This value is automatically prefilled using the DevEUI',
  onboardingDisabled:
    'Device onboarding can only be performed on deployments that have Network Server, Application Server and Join Server activated. Please use the CLI to register devices on individual components.',
  // Manual messages.
  beaconFrequency: 'Beacon frequency',
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
  // QR code section.
  hasEndDeviceQR: 'Does your end device have a QR code? Scan it to speed up onboarding.',
  learnMore: 'Learn more',
  scanEndDevice: 'Scan end device QR code',
  deviceInfo: 'Found QR code data',
  resetQRCodeData: 'Reset QR code data',
  resetConfirm:
    'Are you sure you want to discard QR code data? The scanned device will not be registered and the form will be reset.',
  scanSuccess: 'QR code scanned successfully',
})
