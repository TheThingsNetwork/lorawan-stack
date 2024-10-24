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
import { Routes, Route, useParams, useLocation } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'

import { GATEWAY } from '@console/constants/entities'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import GatewayOverviewHeader from '@console/containers/gateway-overview-header'
import EventSplitFrame from '@console/containers/event-split-frame'
import GatewayEvents from '@console/containers/gateway-events'

import GatewayCollaborators from '@console/views/gateway-collaborators'
import GatewayLocation from '@console/views/gateway-location'
import GatewayData from '@console/views/gateway-data'
import GatewayGeneralSettings from '@console/views/gateway-general-settings'
import GatewayApiKeys from '@console/views/gateway-api-keys'
import GatewayOverview from '@console/views/gateway-overview'
import GatewayManagedGateway from '@console/views/gateway-managed-gateway'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'

import {
  getGateway,
  stopGatewayEventsStream,
  getGatewaysRightsList,
  getGatewayClaimInfoByEui,
} from '@console/store/actions/gateways'
import { getGsFrequencyPlans } from '@console/store/actions/configuration'
import { trackRecencyFrequencyItem } from '@console/store/actions/recency-frequency-items'

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

      await dispatch(getGsFrequencyPlans())

      const { ids } = await dispatch(attachPromise(getGateway(gtwId, selector)))

      await dispatch(attachPromise(getGatewayClaimInfoByEui(ids.eui, true)))
    },
    [gtwId],
  )
  useEffect(() => () => dispatch(stopGatewayEventsStream(gtwId)), [gtwId, dispatch])

  // Track gateway access.
  useEffect(() => {
    dispatch(trackRecencyFrequencyItem(GATEWAY, gtwId))
  }, [dispatch, gtwId])

  const gateway = useSelector(selectSelectedGateway)
  const hasGateway = Boolean(gateway)

  return (
    <RequireRequest requestAction={initialFetch} requestOnChange>
      {hasGateway && <GatewayInner />}
    </RequireRequest>
  )
}

const GatewayInner = () => {
  const { gtwId } = useParams()
  const { pathname } = useLocation()
  const gateway = useSelector(selectSelectedGateway)
  const isEventsPath = pathname.endsWith('/data')

  const gatewayName = gateway?.name || gtwId

  useBreadcrumbs('gtws.single', <Breadcrumb path={`/gateways/${gtwId}`} content={gatewayName} />)

  return (
    <>
      <IntlHelmet titleTemplate={`%s - ${gatewayName} - ${selectApplicationSiteName()}`} />
      <GatewayOverviewHeader gateway={gateway} />
      <Routes>
        <Route index Component={GatewayOverview} />
        <Route path="managed-gateway/*" Component={GatewayManagedGateway} />
        <Route path="api-keys/*" Component={GatewayApiKeys} />
        <Route path="collaborators/*" Component={GatewayCollaborators} />
        <Route path="location" Component={GatewayLocation} />
        <Route path="data" Component={GatewayData} />
        <Route path="general-settings" Component={GatewayGeneralSettings} />
        <Route path="*" element={GenericNotFound} />
      </Routes>
      {!isEventsPath && (
        <EventSplitFrame>
          <GatewayEvents gtwId={gtwId} darkTheme framed />
        </EventSplitFrame>
      )}
    </>
  )
}

export default Gateway
