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

import { CONNECTION_TYPES } from '@console/containers/gateway-the-things-station/connection-profiles/utils'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { ipAddress } from '@console/lib/regexp'

import m from './messages'

const hasSelectedAccessPoint = value =>
  (value.ssid !== '' && value._type === 'all') || value._type === 'other'

const wifiValidationSchema = Yup.object().shape({
  access_point: Yup.object()
    .shape({
      _type: Yup.string(),
      ssid: Yup.string()
        .default('')
        .when('_type', {
          is: 'other',
          then: schema => schema.required(sharedMessages.validateRequired),
          otherwise: schema => schema.strip(),
        }),
      password: Yup.string().when('_type', {
        is: 'other',
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

const ethernetValidationSchema = Yup.object().shape({})

const hasAtLeastOneValidEntry = dnsServers =>
  dnsServers &&
  dnsServers.length > 0 &&
  dnsServers.some(entry => entry !== '' && entry !== undefined && Boolean(ipAddress.test(entry)))

const hasNoEmptyEntry = dnsServers =>
  dnsServers && dnsServers.every(entry => entry !== '' && entry !== undefined)

export const validationSchema = Yup.object({
  _connection_type: Yup.string()
    .oneOf(Object.values(CONNECTION_TYPES))
    .default(CONNECTION_TYPES.WIFI),
  name: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(50, Yup.passValues(sharedMessages.validateTooLong))
    .required(sharedMessages.validateRequired),
  default_network_interface: Yup.boolean(),
  ip_address: Yup.string().when('default_network_interface', {
    is: false,
    then: schema =>
      schema
        .required(sharedMessages.validateRequired)
        .matches(ipAddress, Yup.passValues(m.validateIpAddress)),
    otherwise: schema => schema.strip(),
  }),
  subnet_mask: Yup.string().when('default_network_interface', {
    is: false,
    then: schema =>
      schema
        .required(sharedMessages.validateRequired)
        .matches(ipAddress, Yup.passValues(m.validateIpAddress)),
    otherwise: schema => schema.strip(),
  }),
  dns_servers: Yup.array().when('default_network_interface', {
    is: false,
    then: schema =>
      schema
        .default([])
        .test('has at least one entry', m.validateDnsServers, hasAtLeastOneValidEntry)
        .test('has no empty entry', m.validateEmptyDnsServer, hasNoEmptyEntry),
    otherwise: schema => schema.strip(),
  }),
}).when('._connection_type', {
  is: CONNECTION_TYPES.WIFI,
  then: schema => schema.concat(wifiValidationSchema),
  otherwise: schema => schema.concat(ethernetValidationSchema),
})

export default validationSchema
