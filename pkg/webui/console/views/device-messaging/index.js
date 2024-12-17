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
import { Routes, Route, Navigate, useParams } from 'react-router-dom'
import { useSelector } from 'react-redux'

import Tabs from '@ttn-lw/components/tabs'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import DownlinkForm from '@console/components/downlink-form'
import UplinkForm from '@console/components/uplink-form'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  mayWriteTraffic,
  mayScheduleDownlinks,
  maySendUplink,
  checkFromState,
} from '@console/lib/feature-checks'

const DeviceMessaging = () => {
  const { appId, devId } = useParams()
  const mayScheduleDown = useSelector(state => checkFromState(mayScheduleDownlinks, state))
  const maySendUp = useSelector(state => checkFromState(maySendUplink, state))
  const tabs =
    mayScheduleDown && maySendUp
      ? [
          { title: sharedMessages.scheduleDownlink, name: 'downlink', link: 'downlink' },
          { title: sharedMessages.simulateUplink, name: 'uplink', link: 'uplink' },
        ]
      : []

  useBreadcrumbs(
    'apps.single.devices.single.messaging',
    <Breadcrumb
      path={`/applications/${appId}/devices/${devId}`}
      content={sharedMessages.messaging}
    />,
  )

  return (
    <Require
      featureCheck={mayWriteTraffic}
      otherwise={{ redirect: `/applications/${appId}/devices/${devId}` }}
    >
      <div className="container container--xl grid">
        <IntlHelmet title={sharedMessages.messaging} />
        {tabs.length > 0 && (
          <div className="item-12">
            <Tabs className="mt-0 mb-ls-s md-lg:bg-none md-lg:mr-0" tabs={tabs} divider />
          </div>
        )}
        <div className="item-12">
          <Routes>
            {mayScheduleDown && <Route path="downlink" Component={DownlinkForm} />}
            {maySendUp && <Route path="uplink" Component={UplinkForm} />}
            <Route index element={<Navigate to="downlink" replace />} />
          </Routes>
        </div>
      </div>
    </Require>
  )
}

export default DeviceMessaging
