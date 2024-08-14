// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
  officialDocumentation: 'Official documentation',
  connections: 'Connections',
  connectedToGatewayController: 'Connected to the <span>Gateway Controller</span> via {type}',
  disconnectedFromGatewayController: 'Disconnected from the <span>Gateway Controller</span>',
  connectedToGatewayServer: 'Connected to the <span>Gateway Server</span> via {type}',
  disconnectedFromGatewayServer: 'Disconnected from the <span>Gateway Server</span>',
  cellular: 'Cellular',
  wifi: 'WiFi',
  ethernet: 'Ethernet',
  macAddress: 'MAC address: {address}',
  rssiValue: '{value} dBm',
  ipAddress: 'IP address',
  bssid: 'BSSID',
  hardwareVersion: 'Hardware version: <span>{version}</span>',
  firmwareVersion: 'Firmware version: <span>{version}</span>',
  connectedVia: 'Connected via {connectedVia}',
  cpuTemperature: 'CPU temperature: {temperature}',
})

export default messages
