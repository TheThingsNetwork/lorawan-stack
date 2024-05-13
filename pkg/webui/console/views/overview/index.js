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

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import ShortcutPanel from '@console/containers/shortcut-panel'
import NotificationsDashboardPanel from '@console/containers/notifications-dashboard-panel'
import DocumentationDashboardPanel from '@console/containers/documentation-dashboard-panel'
import TopEntitiesDashboardPanel from '@console/containers/top-entities-dashboard-panel'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getApplicationsList } from '@console/store/actions/applications'
import { getGatewaysList } from '@console/store/actions/gateways'

const Overview = () => {
  useBreadcrumbs('overview', <Breadcrumb path="/" content={sharedMessages.overview} />)

  return (
    <RequireRequest
      requestAction={[
        getApplicationsList(),
        getGatewaysList(undefined, ['name', 'gateway_server_address'], {
          withStatus: true,
        }),
      ]}
    >
      <div className="container container--xl grid p-ls-s gap-ls-s">
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <TopEntitiesDashboardPanel />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <NotificationsDashboardPanel />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <DocumentationDashboardPanel />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <ShortcutPanel />
        </div>
      </div>
    </RequireRequest>
  )
}

export default Overview
