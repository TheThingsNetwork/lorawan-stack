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

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import BlurryNetworkActivityPanel from '@console/components/blurry-network-activity-panel'
import ApplicationMapPanel from '@console/components/application-map-panel'

import LatestDecodedPayloadPanel from '@console/containers/latest-decoded-payload-panel'
import ApplicationOverviewHeader from '@console/containers/application-overview-header'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { isOtherClusterApp } from '@console/lib/application-utils'
import { mayViewApplicationInfo } from '@console/lib/feature-checks'
import { checkFromState } from '@account/lib/feature-checks'

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

  const otherwise = {
    redirect: '/applications',
    message: m.failedAccessOtherHostApplication,
  }

  return (
    <Require condition={condition} otherwise={otherwise}>
      <IntlHelmet title={sharedMessages.overview} />
      <ApplicationOverviewHeader />
      <div className="container container--xl grid p-ls-s gap-ls-s">
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <div style={{ height: '30rem', backgroundColor: 'lightgray' }} />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <BlurryNetworkActivityPanel />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <LatestDecodedPayloadPanel
            events={events}
            shortCutLinkPath={`/applications/${appId}/data`}
          />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
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
