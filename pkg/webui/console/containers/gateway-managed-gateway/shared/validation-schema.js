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

import { ipAddress } from '@console/lib/regexp'

const m = defineMessages({
  validateDnsServers: 'There are some not valid dns servers.',
  validateIpAddresses: 'There are some not valid IP addresses.',
  validateIpAddress: '{field} must contain a valid address.',
  validateNotSelectedAccessPoint: 'There must be at least one access point / SSID selected',
})

const hasSelectedAccessPoint = value =>
  (value.ssid !== '' && value._type === 'all') || value._type === 'other'

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
    .test('has valid entries', m.validateIpAddresses, hasValidIpAddresses),
  subnet_mask: Yup.string()
    .required(sharedMessages.validateRequired)
    .matches(ipAddress, Yup.passValues(m.validateIpAddress)),
  gateway: Yup.string()
    .required(sharedMessages.validateRequired)
    .matches(ipAddress, Yup.passValues(m.validateIpAddress)),
  dns_servers: Yup.array()
    .default([])
    .test('has valid entries', m.validateDnsServers, hasValidDnsServers),
}

export const wifiValidationSchema = Yup.object().shape({
  profile_name: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(50, Yup.passValues(sharedMessages.validateTooLong))
    .required(sharedMessages.validateRequired),
  profileOf: Yup.string(),
  default_network_interface: Yup.boolean(),
  network_interface_addresses: Yup.object().when('default_network_interface', {
    is: false,
    then: schema => schema.shape(networkInterfaceSettings),
    otherwise: schema => schema.strip(),
  }),
  access_point: Yup.object()
    .shape({
      _type: Yup.string(),
      ssid: Yup.string()
        .default('')
        .when('_type', {
          is: 'other',
          then: schema => schema.required(sharedMessages.validateRequired),
        }),
      password: Yup.string().when('security', {
        is: 'WPA2',
        then: schema =>
          schema
            .min(8, Yup.passValues(sharedMessages.validateTooShort))
            .required(sharedMessages.validateRequired),
        otherwise: schema => schema.strip(),
      }),
      security: Yup.string(),
      signal_strength: Yup.number(),
      is_active: Yup.bool(),
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
