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

import { APPLICATION, END_DEVICE, GATEWAY } from '@console/constants/entities'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { IconApplication, IconDevice, IconGateway } from '@ttn-lw/components/icon'

import BlurryNocMetricsPanel from '@console/components/blurry-noc-metrics-panel'

import ShortcutPanel from '@console/containers/shortcut-panel'
import NotificationsDashboardPanel from '@console/containers/notifications-dashboard-panel'
import DocumentationDashboardPanel from '@console/containers/documentation-dashboard-panel'
import TopEntitiesDashboardPanel from '@console/containers/top-entities-dashboard-panel'

import sharedMessages from '@ttn-lw/lib/shared-messages'

const Overview = () => {
  useBreadcrumbs('overview.dashboard', <Breadcrumb path="/" content={sharedMessages.dashboard} />)

  return (
    <div className="container container--xl grid p-ls-s gap-ls-s md:p-cs-xs md:gap-cs-xs">
      <div className="item-12 md-lg:item-4">
        <BlurryNocMetricsPanel
          entity={APPLICATION}
          title={sharedMessages.activeApplications}
          icon={IconApplication}
          entityPath="/applications"
        />
      </div>
      <div className="item-12 md-lg:item-4">
        <BlurryNocMetricsPanel
          entity={GATEWAY}
          title={sharedMessages.connectedGateways}
          icon={IconGateway}
          entityPath="/gateways"
        />
      </div>
      <div className="item-12 md-lg:item-4">
        <BlurryNocMetricsPanel
          entity={END_DEVICE.split('_')[1]}
          title={sharedMessages.activeDevices}
          icon={IconDevice}
        />
      </div>
      <div className="item-12 xl:item-6 md-lg:item-6">
        <TopEntitiesDashboardPanel />
      </div>
      <div className="item-12 xl:item-6 md-lg:item-6">
        <NotificationsDashboardPanel />
      </div>
      <div className="item-12 xl:item-6 md-lg:item-6">
        <DocumentationDashboardPanel />
      </div>
      <div className="item-12 xl:item-6 md-lg:item-6">
        <ShortcutPanel />
      </div>
    </div>
  )
}

export default Overview
