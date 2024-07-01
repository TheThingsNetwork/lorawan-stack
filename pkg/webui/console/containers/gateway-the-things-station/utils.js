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

import m from './connection-profiles/messages'

export const CONNECTION_TYPES = Object.freeze({
  WIFI: 'wifi',
  ETHERNET: 'ethernet',
})

export const getFormTypeMessage = (type, profileId) => {
  if (type === CONNECTION_TYPES.WIFI) {
    return Boolean(profileId) ? m.updateWifiProfile : m.addWifiProfile
  }
  return Boolean(profileId) ? m.updateEthernetProfile : m.addEthernetProfile
}

export const getInitialProfile = (type, isShared) => ({
  _connection_type: type,
  profile_name: '',
  shared: isShared,
  ...(type === CONNECTION_TYPES.WIFI && {
    access_point: {
      _type: 'all',
      ssid: '',
      password: '',
      security: '',
      signal_strength: 0,
      is_active: true,
    },
  }),
  default_network_interface: true,
  network_interface_addresses: {
    ip_addresses: [''],
    subnet_mask: '',
    gateway: '',
    dns_servers: [''],
  },
})
