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
  abp: 'Activation by personalization (ABP)',
  activationMode: 'Activation mode',
  activationSettings: 'Activation settings',
  appEUIDescription:
    'The AppEUI is a global application ID that uniquely identifies the owner of the end device',
  appKeyDescription:
    'The root key to derive session keys to secure communication between the end device and the application',
  appKeyNewDescription:
    'The root key to derive the application session key to secure communication between the end device and the application',
  appSKeyDescription: 'Application session key',
  asServerID: 'Application Server ID',
  asServerIDDescription: 'The AS-ID of the Application Server to use',
  asServerKekLabel: 'Application Server KEK label',
  asServerKekLabelDescription:
    'The KEK label of the Application Server to use for wrapping the application session key',
  createDevice: 'Create end device',
  deviceAddrDescription:
    'Device address, issued by the Network Server or chosen by device manufacturer in case of testing range',
  deleteDevice: 'Delete end device',
  deleteSuccess: 'End device deleted',
  deleteWarning:
    'Are you sure you want to delete "{deviceId}"? Deleting an end device cannot be undone!',
  deviceDescDescription:
    'Optional device description; can also be used to save notes about the end device',
  deviceDescPlaceholder: 'Description for my new end device',
  deviceEUIDescription: 'The DevEUI is the unique identifier for this end device',
  deviceIdPlaceholder: 'my-new-device',
  deviceNamePlaceholder: 'My new end device',
  external: 'External',
  externalJoinServer: 'External Join Server',
  fNwkSIntKeyDescription: 'Forwarding network session integrity key',
  homeNetID: 'Home NetID',
  homeNetIDDescription: 'ID to identify the LoRaWAN network',
  joinEUIDescription: 'JoinEUI identifies the Join Server',
  leaveBlankPlaceholder: 'Leave blank to generate automatically',
  lorawanOptions: 'LoRaWAN options',
  lorawanVersionDescription: 'The LoRaWAN MAC version of the end device',
  lorawanPhyVersionDescription: 'The LoRaWAN PHY version of the end device',
  multicast: 'Multicast',
  nsServerKekLabel: 'Network Server KEK label',
  nsServerKekLabelDescription:
    'The KEK label of the Network Server to use for wrapping the network session key',
  nwkSEncKeyDescription: 'Network session encryption key',
  nwkKeyDescription:
    'The root key to derive network session keys to secure communication between the end device and the network',
  nwkSKeyDescription: 'Network session key',
  otaa: 'Over the air activation (OTAA)',
  resetsFCnt: 'Resets frame counters',
  resetsJoinNonces: 'Resets join nonces',
  resetWarning: 'Reseting is insecure and makes your end device susceptible for replay attacks',
  sNwkSIKeyDescription: 'Serving network session integrity key',
  supportsClassC: 'Supports class C',
  unexposed: 'Unexposed',
  updateSuccess: 'End device updated',
})
