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
  basicTitle: 'Basic settings',
  basicDescription: "General settings of the end device. End device ID's, Name and Description",
  basicDetails: 'Defines general settings of an end device',
  networkTitle: 'Network layer settings',
  networkDescription: 'Frequency plan, regional parameters, end device class and session keys.',
  appTitle: 'Application layer settings',
  appDescription: 'Application session key to encrypt/decrypt LoRaWAN payload.',
  joinTitle: 'Join settings',
  joinDescription: 'Root keys, net ID and kek labels.',
})

export default messages
