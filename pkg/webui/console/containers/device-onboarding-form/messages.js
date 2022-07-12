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
  inputMethod: 'Input Method',
  inputMethodDeviceRepo: 'Select the end device in the LoRaWAN Device Repository',
  inputMethodManual: 'Enter end device specifics manually',
  endDeviceType: 'End device type information',
  provisioningTitle: 'Provisioning information',
  deviceIdDescription: 'This value is automatically prefilled using the DevEUI',
  claimEndDevice: 'Claim end device',
  registerEndDevice: 'Register end device',
  otherOption: 'Other…',
  typeToSearch: 'Type to search…',
  unknownHwOption: 'Unknown ver.',
  afterRegistration: 'After registration',
  singleRegistration: 'View registered end device',
  multipleRegistration: 'Register another end device of this type',
  createSuccess: 'End device registered',
  changeDeviceTypeButton: 'Change device type',
  changeDeviceType:
    'Are you sure you want to switch device type information? Already entered provisioning information might be lost',
})
