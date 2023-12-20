// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
import { useSelector } from 'react-redux'
import { useLocation, useParams, Routes, Route } from 'react-router-dom'
import { Col, Row, Container } from 'react-grid-system'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Tabs from '@ttn-lw/components/tabs'

import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import DeviceTitleSection from '@console/containers/device-title-section'

import DeviceData from '@console/views/device-data'
import DeviceGeneralSettings from '@console/views/device-general-settings'
import DeviceMessaging from '@console/views/device-messaging'
import DeviceLocation from '@console/views/device-location'
import DevicePayloadFormatters from '@console/views/device-payload-formatters'
import DeviceOverview from '@console/views/device-overview'

import getHostnameFromUrl from '@ttn-lw/lib/host-from-url'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectApplicationSiteName, selectAsConfig } from '@ttn-lw/lib/selectors/env'

import {
  mayScheduleDownlinks as mayScheduleDownlinksCheck,
  maySendUplink as maySendUplinkCheck,
  checkFromState,
} from '@console/lib/feature-checks'

import { selectSelectedDevice } from '@console/store/selectors/devices'
import { selectSelectedApplicationId } from '@console/store/selectors/applications'

import style from './device.styl'

const Device = () => {
  const { devId } = useParams()
  const appId = useSelector(selectSelectedApplicationId)
  const device = useSelector(state => selectSelectedDevice(state))

  const { name, application_server_address } = device

  const mayScheduleDownlinks = useSelector(state =>
    checkFromState(mayScheduleDownlinksCheck, state),
  )
  const maySendUplink = useSelector(state => checkFromState(maySendUplinkCheck, state))

  const location = useLocation()

  const siteName = selectApplicationSiteName()
  const asConfig = selectAsConfig()
  const hasAs =
    asConfig.enabled && application_server_address === getHostnameFromUrl(asConfig.base_url)
  const hideMessaging = !hasAs || !(mayScheduleDownlinks || maySendUplink)
  const hidePayloadFormatters = !hasAs

  const basePath = `/applications/${appId}/devices/${devId}`

  const payloadFormattersLink = location.pathname.startsWith(`${basePath}/payload-formatters`)
    ? location.pathname
    : 'payload-formatters'
  const messagingLink = location.pathname.startsWith(`${basePath}/messaging`)
    ? location.pathname
    : 'messaging'

  useBreadcrumbs(
    'apps.single.devices.single',
    <Breadcrumb path={`/applications/${appId}/devices/${devId}`} content={name || devId} />,
  )

  const tabs = [
    { title: sharedMessages.overview, name: 'overview', link: basePath },
    { title: sharedMessages.liveData, name: 'data', link: `${basePath}/data` },
    {
      title: sharedMessages.messaging,
      name: 'messaging',
      exact: false,
      link: messagingLink,
      hidden: hideMessaging,
    },
    { title: sharedMessages.location, name: 'location', link: `${basePath}/location` },
    {
      title: sharedMessages.payloadFormatters,
      name: 'develop',
      link: payloadFormattersLink,
      exact: false,
      hidden: hidePayloadFormatters,
    },
    {
      title: sharedMessages.generalSettings,
      name: 'general-settings',
      link: `${basePath}/general-settings`,
    },
  ]

  return (
    <>
      <IntlHelmet titleTemplate={`%s - ${name || devId} - ${siteName}`} />
      <div className={style.titleSection}>
        <Container>
          <Row>
            <Col sm={12}>
              <DeviceTitleSection appId={appId} devId={devId}>
                <Tabs className={style.tabs} narrow tabs={tabs} />
              </DeviceTitleSection>
            </Col>
          </Row>
        </Container>
      </div>
      <Routes>
        <Route index Component={DeviceOverview} />
        <Route path="data" Component={DeviceData} />
        {!hideMessaging && <Route path="messaging/*" Component={DeviceMessaging} />}
        <Route path="location" Component={DeviceLocation} />
        <Route path="general-settings" Component={DeviceGeneralSettings} />
        {!hidePayloadFormatters && (
          <Route path="payload-formatters/*" Component={DevicePayloadFormatters} />
        )}
        <Route path="*" element={<GenericNotFound />} />
      </Routes>
    </>
  )
}

export default Device
