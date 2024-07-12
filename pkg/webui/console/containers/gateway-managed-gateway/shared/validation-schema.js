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

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { ipAddress, subnetMask } from '@console/lib/regexp'

const m = defineMessages({
  validateDnsServers: 'There are some not valid dns servers.',
  validateIpAddresses: 'There are some not valid IP addresses.',
  validateIpAddress: '{field} must contain a valid address.',
  validateSubnetMask: '{field} must contain a valid subnet mask.',
  validateNotSelectedAccessPoint: 'There must be at least one access point / SSID selected',
  addressesValidateTooMany: '{field} must be 2 items or fewer',
})

const hasSelectedAccessPoint = value =>
  (value.ssid !== '' && value.type === 'all') || value.type === 'other'

const hasValidIpAddresses = ipAddresses =>
  ipAddresses &&
  ipAddresses.length > 0 &&
  ipAddresses.every(entry => entry !== '' && entry !== undefined && Boolean(ipAddress.test(entry)))

const hasValidDnsServers = dnsServers =>
  dnsServers &&
  dnsServers.every(entry => entry !== '' && entry !== undefined && Boolean(ipAddress.test(entry)))

const networkInterfaceSettings = {
  ip_addresses: Yup.array()
    .default([])
    .test('has valid entries', m.validateIpAddresses, hasValidIpAddresses)
    .max(2, Yup.passValues(m.addressesValidateTooMany)),
  subnet_mask: Yup.string()
    .required(sharedMessages.validateRequired)
    .matches(subnetMask, Yup.passValues(m.validateSubnetMask)),
  gateway: Yup.string()
    .required(sharedMessages.validateRequired)
    .matches(ipAddress, Yup.passValues(m.validateIpAddress)),
  dns_servers: Yup.array()
    .default([])
    .test('has valid entries', m.validateDnsServers, hasValidDnsServers)
    .max(2, Yup.passValues(m.addressesValidateTooMany)),
}

export const wifiValidationSchema = Yup.object().shape({
  profile_name: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(50, Yup.passValues(sharedMessages.validateTooLong))
    .required(sharedMessages.validateRequired),
  _profileOf: Yup.string(),
  _default_network_interface: Yup.boolean(),
  network_interface_addresses: Yup.object().when('_default_network_interface', {
    is: false,
    then: schema => schema.shape(networkInterfaceSettings),
    otherwise: schema => schema.strip(),
  }),
  ssid: Yup.string().when('_access_point', {
    is: accessPoint => accessPoint.type === 'other',
    then: schema => schema.required(sharedMessages.validateRequired),
    otherwise: schema => schema.strip(),
  }),
  password: Yup.string().when('_access_point', {
    is: accessPoint =>
      !Boolean(accessPoint.authentication_mode) ||
      accessPoint.authentication_mode === 'open' ||
      accessPoint.is_password_set,
    then: schema => schema.strip(),
    otherwise: schema =>
      schema
        .min(8, Yup.passValues(sharedMessages.validateTooShort))
        .required(sharedMessages.validateRequired),
  }),
  _access_point: Yup.object()
    .shape({
      type: Yup.string(),
      ssid: Yup.string().default(''),
      bssid: Yup.string(),
      channel: Yup.number(),
      authentication_mode: Yup.string(),
      rssi: Yup.number(),
      is_password_set: Yup.boolean(),
    })
    .test('has access point selected', m.validateNotSelectedAccessPoint, hasSelectedAccessPoint),
})

export const ethernetValidationSchema = Yup.object().shape({
  enable_ethernet_connection: Yup.boolean(),
  use_static_ip: Yup.boolean().when('enable_ethernet_connection', {
    is: false,
    then: schema => schema.required(sharedMessages.validateRequired),
  }),
  network_interface_addresses: Yup.object().when('use_static_ip', {
    is: true,
    then: schema => schema.shape(networkInterfaceSettings),
    otherwise: schema =>
      schema.shape({
        dns_servers: Yup.array()
          .default([])
          .test('has valid entries', m.validateDnsServers, hasValidDnsServers),
      }),
  }),
})
