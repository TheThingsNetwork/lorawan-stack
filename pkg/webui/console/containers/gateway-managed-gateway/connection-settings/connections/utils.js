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

import React from 'react'
import { IconCellSignal5, IconPlugConnected, IconWifi } from '@tabler/icons-react'

import Message from '@ttn-lw/lib/components/message'

import { CONNECTION_TYPES } from '@console/containers/gateway-managed-gateway/shared/utils'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import m from './messages'

export const NETWORK_INTERFACE_TYPES = Object.freeze({
  UNSPECIFIED: 'MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_UNSPECIFIED',
  CELLULAR: 'MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_CELLULAR',
  WIFI: 'MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_WIFI',
  ETHERNET: 'MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_ETHERNET',
})

export const NETWORK_INTERFACE_STATUS = Object.freeze({
  UNSPECIFIED: 'MANAGED_GATEWAY_NETWORK_INTERFACE_STATUS_UNSPECIFIED',
})

export const getConnectionType = type => {
  if (type === NETWORK_INTERFACE_TYPES.CELLULAR) return CONNECTION_TYPES.CELLULAR
  if (type === NETWORK_INTERFACE_TYPES.WIFI) return CONNECTION_TYPES.WIFI
  if (type === NETWORK_INTERFACE_TYPES.ETHERNET) return CONNECTION_TYPES.ETHERNET
  return null
}

export const connectionIconMap = {
  [CONNECTION_TYPES.CELLULAR]: IconCellSignal5,
  [CONNECTION_TYPES.WIFI]: IconWifi,
  [CONNECTION_TYPES.ETHERNET]: IconPlugConnected,
}

export const connectionNameMap = {
  [CONNECTION_TYPES.CELLULAR]: m.cellular,
  [CONNECTION_TYPES.WIFI]: m.wifi,
  [CONNECTION_TYPES.ETHERNET]: m.ethernet,
}

export const getCellularDetails = cellularBackhaul => [
  {
    header: '',
    items: [
      ...(cellularBackhaul.rssi
        ? [
            {
              key: sharedMessages.rssi,
              value: <Message content={m.rssiValue} values={{ value: cellularBackhaul.rssi }} />,
            },
          ]
        : []),
    ],
  },
]

export const getWifiDetails = wifiBackhaul => [
  {
    header: '',
    items: [
      {
        key: m.ipAddress,
        value: wifiBackhaul.network_interface.addresses.ip_addresses[0],
      },
      {
        key: sharedMessages.router,
        value: wifiBackhaul.network_interface.addresses.gateway,
      },
      {
        key: sharedMessages.subnetMask,
        value: wifiBackhaul.network_interface.addresses.subnet_mask,
      },
      {
        key: sharedMessages.security,
        value: wifiBackhaul.authentication_mode,
      },
      {
        key: m.bssid,
        value: formatMACAddress(wifiBackhaul.bssid),
      },
      ...(wifiBackhaul.rssi
        ? [
            {
              key: sharedMessages.rssi,
              value: <Message content={m.rssiValue} values={{ value: wifiBackhaul.rssi }} />,
            },
          ]
        : []),
    ],
  },
]

export const getEthernetDetails = ethernetBackhaul => [
  {
    header: '',
    items: [
      {
        key: m.ipAddress,
        value: ethernetBackhaul.network_interface.addresses.ip_addresses[0],
      },
      {
        key: sharedMessages.router,
        value: ethernetBackhaul.network_interface.addresses.gateway,
      },
      {
        key: sharedMessages.subnetMask,
        value: ethernetBackhaul.network_interface.addresses.subnet_mask,
      },
    ],
  },
]

export const formatMACAddress = value => {
  if (!Boolean(value)) {
    return value
  }

  return value.match(/.{1,2}/g).join(':')
}
