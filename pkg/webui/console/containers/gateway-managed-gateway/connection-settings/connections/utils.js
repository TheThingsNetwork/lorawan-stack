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

import DataSheet from '@ttn-lw/components/data-sheet'

import Message from '@ttn-lw/lib/components/message'

import { CONNECTION_TYPES } from '@console/containers/gateway-managed-gateway/shared/utils'
import style from '@console/containers/gateway-managed-gateway/connection-settings/connections/connections.styl'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import m from './messages'

export const NETWORK_INTERFACE_TYPES = Object.freeze({
  UNSPECIFIED: 'MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_UNSPECIFIED',
  CELLULAR: 'MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_CELLULAR',
  WIFI: 'MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_WIFI',
  ETHERNET: 'MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_ETHERNET',
})

export const getConnectionType = type => {
  if (type === NETWORK_INTERFACE_TYPES.CELLULAR) return CONNECTION_TYPES.CELLULAR
  if (type === NETWORK_INTERFACE_TYPES.WIFI) return CONNECTION_TYPES.WIFI
  if (type === NETWORK_INTERFACE_TYPES.ETHERNET) return CONNECTION_TYPES.ETHERNET
  return null
}

export const isConnected = type => type !== 'MANAGED_GATEWAY_NETWORK_INTERFACE_STATUS_UNSPECIFIED'

export const connectionMessageMap = {
  [CONNECTION_TYPES.CELLULAR]: m.cellular,
  [CONNECTION_TYPES.WIFI]: m.wifi,
  [CONNECTION_TYPES.ETHERNET]: m.ethernet,
}

export const getDetails = details => (
  <details>
    <summary>
      <Message content={sharedMessages.details} />
    </summary>
    <DataSheet data={details} className={style.details} />
  </details>
)

export const getCellularDetails = cellularBackhaul => [
  {
    header: '',
    items: [
      {
        key: sharedMessages.rssi,
        value: <Message content={m.rssiValue} values={{ value: cellularBackhaul.rssi }} />,
      },
    ],
  },
]

export const getWifiDetails = wifiBackhaul => [
  {
    header: '',
    items: [
      {
        key: m.ipAddress,
        value: wifiBackhaul.network_interface.addresses.ip_addresses[0], // TODO: Check logic to display correct IP address
      },
      // TODO: Check router part from wifi details
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
        value: wifiBackhaul.bssid,
      },
      {
        key: sharedMessages.rssi,
        value: <Message content={m.rssiValue} values={{ value: wifiBackhaul.rssi }} />,
      },
    ],
  },
]

export const getEthernetDetails = ethernetBackhaul => [
  {
    header: '',
    items: [
      {
        key: m.ipAddress,
        value: ethernetBackhaul.network_interface.addresses.ip_addresses[0], // TODO: Check logic to display correct IP address
      },
      // TODO: Check router part from ethernet details
      {
        key: sharedMessages.subnetMask,
        value: ethernetBackhaul.network_interface.addresses.subnet_mask,
      },
      // TODO: Check security and bssid part from ethernet details
    ],
  },
]

export const exampleConnectionsResponse = {
  result: {
    entity: {
      ids: {
        gateway_id: 'string',
        eui: '70B3D57ED000ABCD',
      },
      version_ids: {
        brand_id: 'string',
        model_id: 'string',
        hardware_version: 'v1.1',
        firmware_version: 'v1.2b',
        runtime_version: 'string',
      },
      cellular_imei: 'string',
      cellular_imsi: 'string',
      wifi_mac_address: '00:11:22:33:44:55',
      ethernet_mac_address: '00:11:22:33:44:55',
      wifi_profile_id: 'string',
      ethernet_profile_id: 'string',
    },
    location: {
      latitude: 0,
      longitude: 0,
      altitude: 0,
      accuracy: 0,
      source: 'SOURCE_UNKNOWN',
    },
    system_metrics: {
      temperature: 43.2,
    },
    controller_connection: {
      network_interface_type: 'MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_ETHERNET',
    },
    gateway_server_connection: {
      network_interface_type: 'MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_WIFI',
      address: 'string',
    },
    cellular_backhaul: {
      network_interface: {
        status: 'MANAGED_GATEWAY_NETWORK_INTERFACE_STATUS_UNSPECIFIED',
        dhcp_enabled: true,
        addresses: {
          ip_addresses: ['192.0.2.0', '2001:db8::1'],
          subnet_mask: '192.0.2.0',
          gateway: '192.0.2.0',
          dns_servers: ['192.0.2.0', '192.0.2.1'],
        },
      },
      operator: 'KPN',
      rssi: -60,
    },
    wifi_backhaul: {
      network_interface: {
        status: 'MANAGED_GATEWAY_NETWORK_INTERFACE_STATUS_UNSPECIFIED',
        dhcp_enabled: true,
        addresses: {
          ip_addresses: ['192.0.2.0', '2001:db8::1'],
          subnet_mask: '192.0.2.0',
          gateway: '192.0.2.0',
          dns_servers: ['192.0.2.0', '192.0.2.1'],
        },
      },
      ssid: 'IoT Backhaul AP',
      bssid: '00:11:22:33:44:55',
      channel: 0,
      authentication_mode: 'WPA2-PSK',
      rssi: -60,
    },
    ethernet_backhaul: {
      network_interface: {
        status: 'MANAGED_GATEWAY_NETWORK_INTERFACE_STATUS_UNSPECIFIED',
        dhcp_enabled: true,
        addresses: {
          ip_addresses: ['192.0.2.0', '2001:db8::1'],
          subnet_mask: '192.0.2.0',
          gateway: '192.0.2.0',
          dns_servers: ['192.0.2.0', '192.0.2.1'],
        },
      },
    },
  },
  error: {
    code: 0,
    message: 'string',
    details: [
      {
        '@type': 'string',
        additionalProp1: 'string',
        additionalProp2: 'string',
        additionalProp3: 'string',
      },
    ],
  },
}
