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
import classnames from 'classnames'

import gatewayIcon from '@assets/misc/gateway.svg'

import Icon from '@ttn-lw/components/icon'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import {
  connectionMessageMap,
  exampleConnectionsResponse,
  getCellularDetails,
  getConnectionType,
  getDetails,
  getEthernetDetails,
  getWifiDetails,
  isConnected,
  NETWORK_INTERFACE_TYPES,
} from '@console/containers/gateway-managed-gateway/connection-settings/connections/utils'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import m from './messages'

import style from './connections.styl'

const ManagedGatewayConnections = () => {
  const managedGateway = exampleConnectionsResponse

  const gatewayControllerConnection =
    managedGateway.result.controller_connection.network_interface_type
  const gatewayServerConnection =
    managedGateway.result.gateway_server_connection.network_interface_type

  const controllerConnectionIsSpecified =
    gatewayControllerConnection !== NETWORK_INTERFACE_TYPES.UNSPECIFIED
  const serverConnectionIsSpecified =
    gatewayServerConnection !== NETWORK_INTERFACE_TYPES.UNSPECIFIED

  const isCellularConnected = isConnected(
    managedGateway.result.cellular_backhaul.network_interface.status,
  )

  const isWifiConnected = isConnected(managedGateway.result.wifi_backhaul.network_interface.status)

  const isEthernetConnected = isConnected(
    managedGateway.result.ethernet_backhaul.network_interface.status,
  )

  const getIsConnectedDiv = isConnected => (
    <div className="d-flex al-center gap-cs-xxs">
      <Icon icon={isConnected ? 'valid' : 'cancel'} />
      <Message content={isConnected ? sharedMessages.connected : sharedMessages.disconnected} />
    </div>
  )

  return (
    <div className={style.root}>
      <Message className="fw-bold mt-0" component="h2" content={sharedMessages.managedGateway} />
      <div className={classnames(style.top, 'd-flex j-between al-center pb-cs-xs mb-cs-xs')}>
        <div>
          <p className="m-0">
            <Message content={sharedMessages.hardwareVersion} />:{' '}
            {managedGateway.result.entity.version_ids.hardware_version}
          </p>
          <p className="m-0">
            <Message content={sharedMessages.firmwareVersion} />:{' '}
            {managedGateway.result.entity.version_ids.firmware_version}
          </p>
          <div className="d-flex al-center gap-cs-xxs mb-cs-s mt-cs-xxs">
            <Icon icon="cloud" />
            {managedGateway.result.system_metrics.temperature} &deg;C
          </div>
          <Link.Anchor primary href="/gateways/adding-gateways">
            {m.officialDocumentation.defaultMessage}
          </Link.Anchor>
        </div>
        <img className={style.image} src={gatewayIcon} alt="managed-gateway" />
      </div>

      <Message className="fw-bold mt-0 mb-cs-xs" component="h3" content={m.connections} />
      {controllerConnectionIsSpecified && (
        <div className="d-flex al-center gap-cs-xxs">
          <Icon icon="cloud" />
          <Message
            content={m.connectedToGatewayController}
            values={{
              type: connectionMessageMap[
                getConnectionType(gatewayControllerConnection)
              ]?.defaultMessage.toLowerCase(),
            }}
          />
        </div>
      )}
      {serverConnectionIsSpecified && (
        <div className="d-flex al-center gap-cs-xxs">
          <Icon icon="cloud" />
          <Message
            content={m.connectedToGatewayServer}
            values={{
              type: connectionMessageMap[
                getConnectionType(gatewayServerConnection)
              ]?.defaultMessage.toLowerCase(),
            }}
          />
        </div>
      )}

      <Message className="fw-bold mb-cs-xs" component="h4" content={m.cellular} />
      <div className="d-flex al-center gap-cs-m">
        {getIsConnectedDiv(isCellularConnected)}
        {isCellularConnected && (
          <div className="d-flex al-center gap-cs-xxs">
            <Icon icon="signal_cellular_alt" />
            {managedGateway.result.cellular_backhaul.operator}
          </div>
        )}
      </div>
      {isCellularConnected &&
        getDetails(getCellularDetails(managedGateway.result.cellular_backhaul))}

      <Message className="fw-bold mb-cs-xs" component="h4" content={m.wifi} />
      <div className="d-flex al-center gap-cs-m">
        {getIsConnectedDiv(isWifiConnected)}
        {isWifiConnected && (
          <div className="d-flex al-center gap-cs-xxs">
            <Icon icon="wifi" />
            {managedGateway.result.wifi_backhaul.ssid}
          </div>
        )}
      </div>
      <div>
        <Message content={m.macAddress} />: {managedGateway.result.entity.wifi_mac_address}
      </div>
      {isWifiConnected && getDetails(getWifiDetails(managedGateway.result.wifi_backhaul))}

      <Message className="fw-bold mb-cs-xs" component="h4" content={m.ethernet} />
      <div className="d-flex al-center gap-cs-m">
        {getIsConnectedDiv(isEthernetConnected)}
        {isEthernetConnected && (
          <div className="d-flex al-center gap-cs-xxs">
            <Icon icon="router" />
            {/* TODO: Check which property is displayed here*/}
            {managedGateway.result.ethernet_backhaul.network_interface.addresses.gateway}
          </div>
        )}
      </div>
      <div>
        <Message content={m.macAddress} />: {managedGateway.result.entity.ethernet_mac_address}
      </div>
      {isEthernetConnected &&
        getDetails(getEthernetDetails(managedGateway.result.ethernet_backhaul))}
    </div>
  )
}

export default ManagedGatewayConnections
