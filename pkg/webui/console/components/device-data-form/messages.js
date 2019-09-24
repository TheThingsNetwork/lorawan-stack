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
  abp: 'Activation By Personalization (ABP)',
  activationMode: 'Activation Mode',
  activationSettings: 'Activation Settings',
  appKeyDescription:
    'The root key to derive the application session key to secure communication between the device and the application',
  appSKeyDescription: 'Application Session Key',
  createDevice: 'Create Device',
  deleteDevice: 'Delete End Device',
  deleteSuccess: 'The end device has been deleted successfully',
  deleteWarning:
    'Are you sure you want to delete "{deviceId}"? Deleting an end device cannot be undone!',
  deviceDescDescription:
    'Optional device description; can also be used to save notes about the device',
  deviceEUIDescription: 'The DevEUI is the unique identifier for this device on the network.',
  deviceIdDescription: 'ID of your device; this cannot be changed afterwards',
  deviceIdPlaceholder: 'my-new-device',
  deviceNameDescription: 'Human friendly name of your device for display purposes',
  deviceNamePlaceholder: 'My New Device',
  external: 'External',
  externalJoinServer: 'External Join Server',
  fwdNtwkKeyDescription: 'Forwarding Network Session Integrity Key (or LoRaWAN 1.0.x NwkSKey)',
  joinEUIDescription: 'JoinEUI identifies the Join Server (in LoRaWAN 1.0.x known as AppEUI)',
  leaveBlankPlaceholder: 'Leave blank to generate automatically',
  ntwkSEncKeyDescription: 'Network Session Encryption Key (only for LoRaWAN 1.1+)',
  nwkKeyDescription:
    'The root key to derive network session keys to secure communication between the device and the network',
  otaa: 'Over The Air Activation (OTAA)',
  resetsFCnt: 'Resets Frame Counters',
  resetsJoinNonces: 'Resets Join Nonces',
  resetWarning: 'Reseting is insecure and makes your device susceptible for replay attacks',
  sNtwkSIKeyDescription: 'Serving Network Session Integrity Key (only for LoRaWAN 1.1+)',
  supportsClassC: 'Supports Class C',
  updateSuccess: 'Successfully updated end device',
  validate16: 'This value needs to be exactly 16 characters long',
  validate32: 'This value needs to be exactly 32 characters long',
  validate8: 'This value needs to be exactly 8 characters long',
})
