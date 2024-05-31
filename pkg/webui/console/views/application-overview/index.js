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
import { defineMessages } from 'react-intl'
import { useParams } from 'react-router-dom'
import { useSelector } from 'react-redux'

import { PAGE_SIZES } from '@ttn-lw/constants/page-sizes'

import DataSheet from '@ttn-lw/components/data-sheet'

import DateTime from '@ttn-lw/lib/components/date-time'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import BlurryNetworkActivityPanel from '@console/components/blurry-network-activity-panel'
import GatewayMapPanel from '@console/components/gateway-map-panel'

import DevicesTable from '@console/containers/devices-table'
import ApplicationEvents from '@console/containers/application-events'
import ApplicationTitleSection from '@console/containers/application-title-section'
import GatewayOverviewHeader from '@console/containers/gateway-overview-header'
import ApplicationOverviewHeader from '@console/containers/application-overview-header'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { isOtherClusterApp } from '@console/lib/application-utils'
import { mayViewApplicationInfo } from '@console/lib/feature-checks'
import { checkFromState } from '@account/lib/feature-checks'

import { selectSelectedApplication } from '@console/store/selectors/applications'

import style from './application-overview.styl'

const m = defineMessages({
  failedAccessOtherHostApplication:
    'The application you attempted to visit is registered on a different cluster and needs to be accessed using its host Console.',
})

const ApplicationOverview = () => {
  const { appId } = useParams()
  const application = useSelector(selectSelectedApplication)
  const may = useSelector(state => checkFromState(mayViewApplicationInfo, state))
  const { created_at, updated_at } = application
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
          <div style={{ height: '30rem', backgroundColor: 'lightgray' }} />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <div style={{ height: '30rem', backgroundColor: 'lightgray' }} />
        </div>
      </div>
    </Require>
  )
}

export default ApplicationOverview
