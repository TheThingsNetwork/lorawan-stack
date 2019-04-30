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

import { defineMessages } from 'react-intl'

export default defineMessages({
  lorawanOptions: 'LoRaWAN Options',
  activationSettings: 'Activation Settings',
  createDevice: 'Create Device',
  deviceIdPlaceholder: 'my-new-device',
  deviceNamePlaceholder: 'My New Device',
  deviceIdDescription: 'Unique Identifier of your device; this cannot be changed afterwards',
  deviceNameDescription: 'Human friendly name of your device for display purposes',
  deviceDescDescription: 'Optional device description; can also be used to save notes about the device',
  joinEUIPlaceholder: 'The connected Join EUI',
  leaveBlankPlaceholder: 'Leave blank to generate automatically',
  resetsJoinNonces: 'Resets Join Nonces',
  resetsFCnt: 'Resets Frame Counters',
  deviceEUIDescription: 'The device EUI is the unique identifier for this device on the network. Can be changed later.',
  nwkKeyDescription: 'The encrypted Network Key',
  appKeyDescription: 'The App Key is used to secure the communication between your device and the network.',
  appSKeyDescription: 'App Session Key',
  fwdNtwkKeyDescription: 'Forwarding Network Session Integrity Key (or LoRaWAN 1.0.x NwkSKey)',
  sNtwkSIKeyDescription: 'Serving Network Session Integrity Key (only for LoRaWAN 1.1+)',
  ntwkSEncKeyDescription: 'Network Session Encryption Key (only for LoRaWAN 1.1+)',
  validate8: 'This value needs to be exactly 8 characters long',
  validate16: 'This value needs to be exactly 16 characters long',
  validate32: 'This value needs to be exactly 32 characters long',
  supportsClassC: 'Supports Class C',
  activationMode: 'Activation Mode',
  otaa: 'Over The Air Activation (OTAA)',
  abp: 'Activation By Personalization (ABP)',
  resetWarning: 'Reseting is insecure and makes your device susceptible for replay attacks',
  couldNotRetrieveFrequencyPlans: 'Could not retrieve the list of available frequency plans',
})
