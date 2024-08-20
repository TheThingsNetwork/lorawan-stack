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

import { useSelector } from 'react-redux'
import { useParams } from 'react-router-dom'
import { useMemo } from 'react'

import { NETWORK_INTERFACE_STATUS } from '@console/containers/gateway-managed-gateway/connection-settings/connections/utils'

import { selectGatewayEvents } from '@console/store/selectors/gateways'

const isConnected = type => type && type !== NETWORK_INTERFACE_STATUS.UNSPECIFIED

const useConnectionsData = () => {
  const { gtwId } = useParams()
  const events = useSelector(state => selectGatewayEvents(state, gtwId))

  const connectionsData = useMemo(() => {
    const matchEvent = (e, regex) => (regex.test(e.name) ? e.data : null)

    const eventTypes = [
      { key: 'systemStatus', regex: /\.system_status\./ },
      { key: 'controllerConnection', regex: /\.controller\./ },
      { key: 'serverConnection', regex: /\.gs\./ },
      { key: 'cellularBackhaul', regex: /\.cellular\./ },
      { key: 'wifiBackhaul', regex: /\.wifi\./ },
      { key: 'ethernetBackhaul', regex: /\.ethernet\./ },
      { key: 'updatedManagedGateway', regex: /\.managed\.update$/ },
    ]

    const result = {}
    for (const e of events) {
      const remainingEventTypes = eventTypes.filter(({ key }) => !(key in result))
      for (const { key, regex } of remainingEventTypes) {
        const matchedData = matchEvent(e, regex)
        if (matchedData !== null) {
          result[key] = matchedData
        }
      }
      if (Object.keys(result).length === eventTypes.length) {
        break
      }
    }

    return result
  }, [events])

  const isCellularConnected = isConnected(
    connectionsData.cellularBackhaul?.network_interface?.status,
  )
  const isWifiConnected = isConnected(connectionsData.wifiBackhaul?.network_interface?.status)
  const isEthernetConnected = isConnected(
    connectionsData.ethernetBackhaul?.network_interface?.status,
  )

  return {
    ...connectionsData,
    isCellularConnected,
    isWifiConnected,
    isEthernetConnected,
  }
}

export default useConnectionsData
