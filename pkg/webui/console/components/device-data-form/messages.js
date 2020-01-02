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
  appEUIDescription:
    'The AppEUI is a global application ID that uniquely identifies the owner of the ­device',
  appKeyDescription:
    'The root key to derive session keys to secure communication between the device and the application',
  appKeyNewDescription:
    'The root key to derive the application session key to secure communication between the device and the application',
  appSKeyDescription: 'Application Session Key',
  asServerID: 'Application Server ID',
  asServerIDDescription: 'The AS-ID of the Application Server to use',
  asServerKekLabel: 'Application Server KEK Label',
  asServerKekLabelDescription:
    'The KEK label of the Application Server to use for wrapping the application session key',
  createDevice: 'Create Device',
  deviceAddrDescription:
    'Device Address, issued by the Network Server or chosen by device manufacturer in case of testing range',
  deleteDevice: 'Delete End Device',
  deleteSuccess: 'The end device has been deleted successfully',
  deleteWarning:
    'Are you sure you want to delete "{deviceId}"? Deleting an end device cannot be undone!',
  deviceDescDescription:
    'Optional device description; can also be used to save notes about the device',
  deviceDescPlaceholder: 'Description for my new device',
  deviceEUIDescription: 'The DevEUI is the unique identifier for this device',
  deviceIdPlaceholder: 'my-new-device',
  deviceNamePlaceholder: 'My New Device',
  external: 'External',
  externalJoinServer: 'External Join Server',
  fNwkSIntKeyDescription: 'Forwarding Network Session Integrity Key',
  homeNetID: 'Home NetID',
  homeNetIDDescription: 'ID to identify the LoRaWAN network',
  joinEUIDescription: 'JoinEUI identifies the Join Server',
  leaveBlankPlaceholder: 'Leave blank to generate automatically',
  lorawanVersionDescription: 'The LoRaWAN MAC Version of the end device',
  multicast: 'Multicast',
  nsServerKekLabel: 'Network Server KEK Label',
  nsServerKekLabelDescription:
    'The KEK label of the Network Server to use for wrapping the network session key',
  nwkSEncKeyDescription: 'Network Session Encryption Key',
  nwkKeyDescription:
    'The root key to derive network session keys to secure communication between the device and the network',
  nwkSKeyDescription: 'Network Session Key',
  otaa: 'Over The Air Activation (OTAA)',
  resetsFCnt: 'Resets Frame Counters',
  resetsJoinNonces: 'Resets Join Nonces',
  resetWarning: 'Reseting is insecure and makes your device susceptible for replay attacks',
  sNwkSIKeyDescription: 'Serving Network Session Integrity Key',
  supportsClassC: 'Supports Class C',
  unexposed: 'Unexposed',
  updateSuccess: 'Successfully updated end device',
  validate16: 'This value needs to be exactly 16 characters long',
  validate32: 'This value needs to be exactly 32 characters long',
  validate6: 'This value needs to be exactly 6 characters long',
  validate8: 'This value needs to be exactly 8 characters long',
})
