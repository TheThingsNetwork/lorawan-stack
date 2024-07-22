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

  const systemStatus = useMemo(
    () => events.find(e => /\.system_status\./.test(e.name))?.data,
    [events],
  )

  const controllerConnection = useMemo(
    () => events.find(e => /\.controller\./.test(e.name))?.data,
    [events],
  )

  const serverConnection = useMemo(() => events.find(e => /\.gs\./.test(e.name))?.data, [events])

  const cellularBackhaul = useMemo(
    () => events.find(e => /\.cellular\./.test(e.name))?.data,
    [events],
  )

  const wifiBackhaul = useMemo(() => events.find(e => /\.wifi\./.test(e.name))?.data, [events])

  const ethernetBackhaul = useMemo(
    () => events.find(e => /\.ethernet\./.test(e.name))?.data,
    [events],
  )

  const updatedManagedGateway = useMemo(
    () => events.find(e => /\.managed\.update$/.test(e.name))?.data,
    [events],
  )

  const isCellularConnected = isConnected(cellularBackhaul?.network_interface?.status)

  const isWifiConnected = isConnected(wifiBackhaul?.network_interface?.status)

  const isEthernetConnected = isConnected(ethernetBackhaul?.network_interface?.status)

  return {
    systemStatus,
    controllerConnection,
    serverConnection,
    cellularBackhaul,
    wifiBackhaul,
    ethernetBackhaul,
    updatedManagedGateway,
    isCellularConnected,
    isWifiConnected,
    isEthernetConnected,
  }
}

export default useConnectionsData
