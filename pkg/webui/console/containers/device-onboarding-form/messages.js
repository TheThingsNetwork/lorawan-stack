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
  cannotConfirmEui:
    'There was an error and the JoinEUI could not be confirmed. Please try again later.',

  // Shared messages.
  classCapabilities: 'Additional LoRaWAN class capabilities',
  afterRegistration: 'After registration',
  singleRegistration: 'View registered end device',
  multipleRegistration: 'Register another end device of this type',
  createSuccess: 'End device registered',
  onboardingDisabled:
    'Device onboarding can only be performed on deployments that have Network Server, Application Server and Join Server activated. Please use the CLI to register devices on individual components.',
  pingSlotDataRateTitle: 'Ping slot data rate',
  rx2DataRateIndexTitle: 'Rx2 data rate',
  defaultNetworksSettings: "Use network's default MAC settings",
  clusterSettings: 'Cluster settings',
  networkDefaults: 'Network defaults',

  // QR code section.
  hasEndDeviceQR:
    'Does your end device have a LoRaWAN® Device Identification QR Code? Scan it to speed up onboarding.',
  deviceGuide: 'Device registration help',
  invalidData:
    'Invalid QR code data. Please note that only TR005 LoRaWAN® Device Identification QR Code can be scanned. Some devices have unrelated QR codes printed on them that cannot be used.',
})
