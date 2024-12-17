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
import { useParams } from 'react-router-dom'
import { defineMessages } from 'react-intl'

import Panel from '@ttn-lw/components/panel'
import DataSheet from '@ttn-lw/components/data-sheet'
import Icon, { IconCircleCheck, IconExclamationCircle } from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { selectSelectedGateway } from '@console/store/selectors/gateways'
import { selectGsFrequencyPlans } from '@console/store/selectors/configuration'

import style from './gateway-general-information-panel.styl'

const m = defineMessages({
  networkSettings: 'Network settings',
  autoUpdateDescription: 'When enabled, the gateway can be updated automatically.',
  requireAuthenticatedConnectionDescription:
    'When enabled, this gateway may only connect if it uses an authenticated Basic Station or MQTT connection.',
  publicStatus: 'Public status',
  publicStatusDescription:
    'When enabled, the status of this gateway may be visible to other users of the network.',
  publicLocation: 'Public location',
  publicLocationDescription:
    'When enabled, the gateway location may be visible to other users of the network.',
  packetBroker: 'Packet Broker forwarding',
  packetBrokerDescription:
    'When disabled, uplink messages received from this gateway will not be forwarded to Packet Broker.',
  locationUpdates: 'Status location updates',
  locationUpdatesDescription:
    'When enabled, the location of this gateway is updated from status messages. This only works for gateways that send their locations within status messages while using an authenticated connection; gateways connected over UDP are not supported. Please refer to the manual of your gateway model to see whether sending location data is supported.',
  enforceDutyCycleDescription:
    'It is recommended that this is enabled for all gateways in order to respect spectrum regulations.',
})

const GatewayGeneralInformationPanel = () => {
  const { gtwId } = useParams()
  const gateway = useSelector(selectSelectedGateway)
  const frequencyPlanIds = useSelector(selectGsFrequencyPlans)
  const {
    created_at,
    ids,
    enforce_duty_cycle,
    frequency_plan_ids,
    location_public,
    status_public,
    update_location_from_status,
    require_authenticated_connection,
    disable_packet_broker_forwarding,
  } = gateway
  const frequencyPlanIdsFormatted = frequency_plan_ids?.map(
    id => frequencyPlanIds.find(plan => plan.id === id)?.name,
  )

  const getNetworkSettingsInfo = useCallback(
    field =>
      field ? (
        <div className="d-flex al-center gap-cs-xxs">
          <Icon icon={IconCircleCheck} small />
          <Message content={sharedMessages.enabled} />
        </div>
      ) : (
        <div className="d-flex al-center gap-cs-xxs">
          <Icon icon={IconExclamationCircle} small className="c-text-error-normal" />
          <Message content={sharedMessages.disabled} />
        </div>
      ),
    [],
  )

  const sheetData = [
    {
      header: sharedMessages.generalInformation,
      items: [
        {
          key: sharedMessages.gatewayID,
          value: gtwId,
          type: 'code',
          sensitive: false,
        },
        {
          key: sharedMessages.gatewayEUI,
          value: ids.eui,
          type: 'byte',
          sensitive: false,
        },
        {
          key: sharedMessages.frequencyPlan,
          value:
            frequencyPlanIdsFormatted?.length !== 0
              ? frequencyPlanIdsFormatted?.join(' , ')
              : undefined,
        },
        {
          key: sharedMessages.createdAt,
          value: <DateTime value={created_at} />,
        },
      ],
    },
    {
      header: m.networkSettings,
      items: [
        {
          key: sharedMessages.requireAuthenticatedConnection,
          tooltipMessage: m.requireAuthenticatedConnectionDescription,
          value: getNetworkSettingsInfo(require_authenticated_connection),
        },
        {
          key: m.publicStatus,
          tooltipMessage: m.publicStatusDescription,
          value: getNetworkSettingsInfo(status_public),
        },
        {
          key: m.publicLocation,
          tooltipMessage: m.publicLocationDescription,
          value: getNetworkSettingsInfo(location_public),
        },
        {
          key: m.packetBroker,
          tooltipMessage: m.packetBrokerDescription,
          value: getNetworkSettingsInfo(!disable_packet_broker_forwarding),
        },
        {
          key: m.locationUpdates,
          tooltipMessage: m.locationUpdatesDescription,
          value: getNetworkSettingsInfo(update_location_from_status),
        },
        {
          key: sharedMessages.enforceDutyCycle,
          tooltipMessage: m.enforceDutyCycleDescription,
          value: getNetworkSettingsInfo(enforce_duty_cycle),
        },
      ],
    },
  ]

  return (
    <Panel className={style.infoPanel}>
      <DataSheet data={sheetData} />
    </Panel>
  )
}

export default GatewayGeneralInformationPanel
