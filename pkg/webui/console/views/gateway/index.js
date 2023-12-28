// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useEffect } from 'react'
import { Routes, Route, useParams } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import GatewayCollaborators from '@console/views/gateway-collaborators'
import GatewayLocation from '@console/views/gateway-location'
import GatewayData from '@console/views/gateway-data'
import GatewayGeneralSettings from '@console/views/gateway-general-settings'
import GatewayApiKeys from '@console/views/gateway-api-keys'
import GatewayOverview from '@console/views/gateway-overview'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'

import {
  getGateway,
  stopGatewayEventsStream,
  getGatewaysRightsList,
} from '@console/store/actions/gateways'

import { selectSelectedGateway } from '@console/store/selectors/gateways'

const Gateway = () => {
  const { gtwId } = useParams()
  const dispatch = useDispatch()
  const initialFetch = useCallback(
    async dispatch => {
      const rights = await dispatch(attachPromise(getGatewaysRightsList(gtwId)))

      const selector = [
        'name',
        'description',
        'enforce_duty_cycle',
        'frequency_plan_ids',
        'gateway_server_address',
        'antennas',
        'location_public',
        'status_public',
        'auto_update',
        'schedule_downlink_late',
        'update_location_from_status',
        'update_channel',
        'schedule_anytime_delay',
        'attributes',
        'require_authenticated_connection',
        'disable_packet_broker_forwarding',
        'administrative_contact',
        'technical_contact',
      ]

      if (rights.includes('RIGHT_GATEWAY_READ_SECRETS')) {
        selector.push('lbs_lns_secret')
      }

      return dispatch(attachPromise(getGateway(gtwId, selector)))
    },
    [gtwId],
  )
  useEffect(() => () => dispatch(stopGatewayEventsStream(gtwId)), [gtwId, dispatch])

  const gateway = useSelector(selectSelectedGateway)
  const hasGateway = Boolean(gateway)

  return (
    <RequireRequest requestAction={initialFetch}>{hasGateway && <GatewayInner />}</RequireRequest>
  )
}

const GatewayInner = () => {
  const { gtwId } = useParams()
  const gateway = useSelector(selectSelectedGateway)

  const gatewayName = gateway?.name || gtwId

  useBreadcrumbs('gtws.single', <Breadcrumb path={`/gateways/${gtwId}`} content={gatewayName} />)

  return (
    <>
      <IntlHelmet titleTemplate={`%s - ${gatewayName} - ${selectApplicationSiteName()}`} />
      <Routes>
        <Route index Component={GatewayOverview} />
        <Route path="api-keys/*" Component={GatewayApiKeys} />
        <Route path="collaborators/*" Component={GatewayCollaborators} />
        <Route path="location" Component={GatewayLocation} />
        <Route path="data" Component={GatewayData} />
        <Route path="general-settings" Component={GatewayGeneralSettings} />
        <Route path="*" element={GenericNotFound} />
      </Routes>
    </>
  )
}

export default Gateway
