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

import { useMemo } from 'react'

export const CONNECTION_TYPES = Object.freeze({
  CELLULAR: 'cellular',
  WIFI: 'wifi',
  ETHERNET: 'ethernet',
})

export const initialWifiProfile = {
  profile_name: '',
  profile_id: '',
  shared: true,
  _profile_of: '',
  ssid: '',
  password: '',
  _access_point: {
    type: 'all',
    ssid: '',
    bssid: '',
    channel: 0,
    authentication_mode: '',
    rssi: 0,
    is_password_set: false,
  },
  _default_network_interface: true,
  network_interface_addresses: {
    ip_addresses: [''],
    subnet_mask: '',
    gateway: '',
    dns_servers: [],
  },
}

export const getValuesNormalized = (namePrefix, values) => {
  if (!namePrefix) return values
  let name = namePrefix
  if (namePrefix.endsWith('.')) {
    name = name.slice(0, -1)
  }
  const nameSplitted = name.split('.')
  let result = values
  nameSplitted.forEach(part => {
    result = result[isNaN(part) ? part : Number(part)]
  })
  return result
}

export const normalizeWifiProfile = (profile, shared = true) => {
  const { _profile_of, _access_point, _default_network_interface, ...rest } = profile

  if (_default_network_interface) {
    rest.network_interface_addresses = undefined
  }

  if (_access_point.is_password_set) {
    delete rest.ssid
    delete rest.password
  }

  if (!Boolean(rest.password)) {
    delete rest.password
  }

  rest.shared = shared

  return rest
}

export const initialEthernetProfile = {
  profile_name: Date.now().toString(),
  shared: false,
  _enable_ethernet_connection: false,
  _use_static_ip: false,
  network_interface_addresses: {
    ip_addresses: [''],
    subnet_mask: '',
    gateway: '',
    dns_servers: [],
  },
}

export const normalizeEthernetProfile = profile => {
  const { _enable_ethernet_connection, _use_static_ip, ...rest } = profile

  if (!_enable_ethernet_connection) {
    rest.network_interface_addresses = undefined
  }

  if (!_use_static_ip) {
    delete rest.network_interface_addresses?.ip_addresses
    delete rest.network_interface_addresses?.subnet_mask
    delete rest.network_interface_addresses?.gateway
  }

  rest.profile_name = Date.now().toString()

  return rest
}
