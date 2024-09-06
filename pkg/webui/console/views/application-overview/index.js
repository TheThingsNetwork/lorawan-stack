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
import { defineMessages } from 'react-intl'
import { useParams } from 'react-router-dom'
import { useSelector } from 'react-redux'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import BlurryNetworkActivityPanel from '@console/components/blurry-network-activity-panel'
import ApplicationMapPanel from '@console/components/application-map-panel'

import LatestDecodedPayloadPanel from '@console/containers/latest-decoded-payload-panel'
import DevicesPanel from '@console/containers/devices-panel'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { isOtherClusterApp } from '@console/lib/application-utils'
import { mayViewApplicationInfo, checkFromState } from '@console/lib/feature-checks'

import { getDevicesList } from '@console/store/actions/devices'

import {
  selectApplicationEvents,
  selectSelectedApplication,
} from '@console/store/selectors/applications'

const m = defineMessages({
  failedAccessOtherHostApplication:
    'The application you attempted to visit is registered on a different cluster and needs to be accessed using its host Console.',
})

const ApplicationOverview = () => {
  const { appId } = useParams()
  const events = useSelector(state => selectApplicationEvents(state, appId))
  const application = useSelector(selectSelectedApplication)
  const may = useSelector(state => checkFromState(mayViewApplicationInfo, state))
  const shouldRedirect = isOtherClusterApp(application)
  const condition = !shouldRedirect && may
  useBreadcrumbs(
    `apps.single#${appId}.overview`,
    <Breadcrumb path={`/applications/${appId}`} content={sharedMessages.appOverview} />,
  )

  const otherwise = {
    redirect: '/applications',
    message: m.failedAccessOtherHostApplication,
  }

  return (
    <Require condition={condition} otherwise={otherwise}>
      <IntlHelmet title={sharedMessages.overview} />
      <div className="container container--xl grid p-ls-s gap-ls-s md:p-cs-xs md:gap-cs-xs">
        <div className="item-12 xl:item-6 lg:item-6">
          <DevicesPanel />
        </div>
        <div className="item-12 xl:item-6 lg:item-6 d-flex">
          <BlurryNetworkActivityPanel />
        </div>
        <div className="item-12 xl:item-6 lg:item-6">
          <LatestDecodedPayloadPanel
            appId={appId}
            events={events}
            shortCutLinkPath={`/applications/${appId}/data`}
          />
        </div>
        <div className="item-12 xl:item-6 lg:item-6 d-flex">
          <RequireRequest
            requestAction={getDevicesList(
              application.ids.application_id,
              { page: 1, limit: 1000 },
              'locations',
            )}
          >
            <ApplicationMapPanel />
          </RequireRequest>
        </div>
      </div>
    </Require>
  )
}

export default ApplicationOverview
