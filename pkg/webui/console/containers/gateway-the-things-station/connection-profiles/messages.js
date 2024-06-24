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
  theThingsStationConnectionProfiles: 'The Things Station connection profiles',
  wifiProfiles: 'WiFi profiles',
  ethernetProfiles: 'Ethernet profiles',
  information:
    'Connection profiles are setup to allow for multiple gateways to connect via the same settings. You can use this view to manage all your profiles or create new ones, after which you can assign them to your gateway.<br></br> <link>Learn more about gateway network connection profiles.</link>',
  addWifiProfile: 'Add WiFi profile',
  addEthernetProfile: 'Add Ethernet profile',
  updateWifiProfile: 'Update WiFi profile',
  updateEthernetProfile: 'Update Ethernet profile',
  profileId: 'Profile ID',
  accessPoint: 'Access point',
  deleteSuccess: 'Connection profile deleted',
  deleteFail: 'There was an error and the connection profile could not be deleted',
  profileName: 'Profile name',
  useDefaultNetworkInterfaceSettings: 'Use default network interface settings',
  uncheckToSetCustomSettings:
    'Uncheck if you need to set custom IP addresses, subnet mask and DNS server',
  ipAddress: 'IP address',
  subnetMask: 'Subnet mask',
  dnsServers: 'DNS servers',
  addServerAddress: 'Add server address',
  validateDnsServers: 'There must be at least one valid dns server.',
  validateEmptyDnsServer:
    'There must be no empty dns server entries. Please remove such entries before submitting.',
  dnsServerPlaceholder: '0.0.0.0',
  validateIpAddress: '{field} must contain a valid address.',
  accessPointAndSsid: 'Access point / SSID',
  validateNotSelectedAccessPoint: 'There must be at least one access point / SSID selected',
  wifiPassword: 'WiFi password',
  ssid: 'SSID',
  isSet: '(is set)',
})

export default messages
