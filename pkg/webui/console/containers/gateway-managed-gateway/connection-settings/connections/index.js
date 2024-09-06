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

import React, { useCallback } from 'react'
import { useSelector } from 'react-redux'
import classnames from 'classnames'

import managedGatewayImage from '@assets/misc/gateway.svg'

import Icon, {
  IconCircleCheck,
  IconExclamationCircle,
  IconThermometer,
} from '@ttn-lw/components/icon'
import DataSheet from '@ttn-lw/components/data-sheet'

import Message from '@ttn-lw/lib/components/message'

import {
  connectionIconMap,
  connectionNameMap,
  formatMACAddress,
  getCellularDetails,
  getConnectionType,
  getEthernetDetails,
  getWifiDetails,
  NETWORK_INTERFACE_TYPES,
} from '@console/containers/gateway-managed-gateway/connection-settings/connections/utils'
import { CONNECTION_TYPES } from '@console/containers/gateway-managed-gateway/shared/utils'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { selectSelectedManagedGateway } from '@console/store/selectors/gateways'

import m from './messages'

import style from './connections.styl'

const ConnectionByType = ({ type, isConnected, details, connectedVia, macAddress }) => (
  <div className="d-flex flex-column gap-cs-xxs">
    <div className="d-flex al-center gap-cs-xxs">
      <Icon className={style.icon} icon={connectionIconMap[type]} />
      <Message content={connectionNameMap[type]} component="p" className="m-0 fw-bold" />
      <div className={classnames(style.connection, 'd-flex al-center gap-cs-xxs ml-cs-xs')}>
        {isConnected ? (
          <Icon icon={IconCircleCheck} className="c-text-success-normal" small />
        ) : (
          <Icon icon={IconExclamationCircle} className="c-text-error-normal" small />
        )}
        <Message
          content={
            isConnected
              ? Boolean(connectedVia)
                ? m.connectedVia
                : sharedMessages.connected
              : sharedMessages.disconnected
          }
          values={{ ...(isConnected && { connectedVia }) }}
          component="p"
          className="m-0"
        />
      </div>
    </div>
    <div className={style.details}>
      {Boolean(macAddress) && (
        <Message
          content={m.macAddress}
          values={{ address: formatMACAddress(macAddress) }}
          className="m-vert-cs-xs d-block"
        />
      )}
      {Boolean(details?.[0]?.items?.length) && (
        <details>
          <summary>
            <Message content={sharedMessages.details} />
          </summary>
          <DataSheet data={details} className={style.detail} />
        </details>
      )}
    </div>
  </div>
)

ConnectionByType.propTypes = {
  connectedVia: PropTypes.string,
  details: PropTypes.arrayOf(
    PropTypes.shape({
      header: PropTypes.string,
      items: PropTypes.array,
    }),
  ),
  isConnected: PropTypes.bool,
  macAddress: PropTypes.string,
  type: PropTypes.oneOf([
    CONNECTION_TYPES.WIFI,
    CONNECTION_TYPES.ETHERNET,
    CONNECTION_TYPES.CELLULAR,
  ]).isRequired,
}

ConnectionByType.defaultProps = {
  isConnected: false,
  connectedVia: undefined,
  details: undefined,
  macAddress: undefined,
}

const ManagedGatewayConnections = ({ connectionsData }) => {
  const selectedManagedGateway = useSelector(selectSelectedManagedGateway)
  const {
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
  } = connectionsData

  const gatewayControllerConnection =
    controllerConnection?.network_interface_type ?? NETWORK_INTERFACE_TYPES.UNSPECIFIED
  const gatewayServerConnection =
    serverConnection?.network_interface_type ?? NETWORK_INTERFACE_TYPES.UNSPECIFIED

  const controllerConnectionIsSpecified =
    gatewayControllerConnection !== NETWORK_INTERFACE_TYPES.UNSPECIFIED
  const serverConnectionIsSpecified =
    gatewayServerConnection !== NETWORK_INTERFACE_TYPES.UNSPECIFIED

  const managedGateway = updatedManagedGateway ?? selectedManagedGateway

  const getConnectionData = useCallback(
    ({ isConnected, connectedMessage, disconnectedMessage, type }) => (
      <div className="d-flex al-center gap-cs-xxs">
        {isConnected ? (
          <Icon icon={IconCircleCheck} className="c-text-success-normal" />
        ) : (
          <Icon icon={IconExclamationCircle} className="c-text-error-normal" />
        )}
        <Message
          content={isConnected ? connectedMessage : disconnectedMessage}
          component="div"
          className="d-flex gap-cs-xxs flex-wrap"
          values={{
            span: txt => <span className="fw-bold">{txt}</span>,
            ...(isConnected && {
              type: (
                <span className="fw-bold">
                  {
                    <div className="d-flex al-center gap-cs-xxs fw-bold">
                      <Icon icon={connectionIconMap[getConnectionType(type)]} />
                      <Message
                        content={connectionNameMap[getConnectionType(type)]}
                        component="p"
                        className="m-0"
                      />
                    </div>
                  }
                </span>
              ),
            }),
          }}
        />
      </div>
    ),
    [],
  )

  return (
    <div className={style.root}>
      <Message
        className="fw-bold m-0 lh-1"
        component="h3"
        content={managedGateway.version_ids?.model_id ?? sharedMessages.managedGateway}
      />
      <div className="d-flex direction-column gap-cs-s">
        <div className={style.top}>
          <div className={classnames(style.imgDiv, 'd-flex al-center')}>
            <img className={style.image} src={managedGatewayImage} alt="managed-gateway" />
          </div>

          <div className="d-flex direction-column j-center p-cs-l gap-cs-xs">
            <Message
              component="p"
              className="m-0 c-text-neutral-light"
              content={m.hardwareVersion}
              values={{
                span: text => <span className="c-text-neutral-heavy">{text}</span>,
                version: managedGateway.version_ids.hardware_version,
              }}
            />
            <Message
              component="p"
              className="m-0 c-text-neutral-light"
              content={m.firmwareVersion}
              values={{
                span: text => <span className="c-text-neutral-heavy">{text}</span>,
                version: managedGateway.version_ids.firmware_version,
              }}
            />

            {systemStatus?.cpu_temperature && (
              <div className="d-flex al-center gap-cs-xxs">
                <IconThermometer />
                <Message
                  content={m.cpuTemperature}
                  values={{ temperature: `${systemStatus.cpu_temperature}°C` }}
                />
              </div>
            )}
          </div>
        </div>

        {/* TODO: Add link to official once available.
        <Link.DocLink primary path="#">
          <Message content={m.officialDocumentation} />
        </Link.DocLink>
        */}
      </div>

      <hr className={classnames(style.horizontalLine, 'w-full m-0')} />

      <Message className="fw-bold m-0 lh-1" component="h3" content={m.connections} />

      <div className={style.connections}>
        {getConnectionData({
          isConnected: controllerConnectionIsSpecified,
          connectedMessage: m.connectedToGatewayController,
          disconnectedMessage: m.disconnectedFromGatewayController,
          type: gatewayControllerConnection,
        })}
        {getConnectionData({
          isConnected: serverConnectionIsSpecified,
          connectedMessage: m.connectedToGatewayServer,
          disconnectedMessage: m.disconnectedFromGatewayServer,
          type: gatewayServerConnection,
        })}
      </div>

      <div className="d-flex flex-column gap-ls-xs">
        <ConnectionByType
          isConnected={isCellularConnected}
          type={CONNECTION_TYPES.CELLULAR}
          details={isCellularConnected && getCellularDetails(cellularBackhaul)}
          connectedVia={cellularBackhaul?.operator}
        />
        <ConnectionByType
          isConnected={isWifiConnected}
          type={CONNECTION_TYPES.WIFI}
          details={isWifiConnected && getWifiDetails(wifiBackhaul)}
          connectedVia={wifiBackhaul?.ssid}
          macAddress={managedGateway.wifi_mac_address}
        />
        <ConnectionByType
          isConnected={isEthernetConnected}
          type={CONNECTION_TYPES.ETHERNET}
          details={isEthernetConnected && getEthernetDetails(ethernetBackhaul)}
          macAddress={managedGateway.ethernet_mac_address}
        />
      </div>
    </div>
  )
}

ManagedGatewayConnections.propTypes = {
  connectionsData: PropTypes.shape({
    systemStatus: PropTypes.object,
    controllerConnection: PropTypes.object,
    serverConnection: PropTypes.object,
    cellularBackhaul: PropTypes.object,
    wifiBackhaul: PropTypes.object,
    ethernetBackhaul: PropTypes.object,
    updatedManagedGateway: PropTypes.object,
    isCellularConnected: PropTypes.bool,
    isWifiConnected: PropTypes.bool,
    isEthernetConnected: PropTypes.bool,
  }).isRequired,
}

export default ManagedGatewayConnections
