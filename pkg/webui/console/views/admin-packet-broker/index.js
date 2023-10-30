// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import { Routes, Route } from 'react-router-dom'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import RequireRequest from '@ttn-lw/lib/components/require-request'
import ValidateRouteParam from '@ttn-lw/lib/components/validate-route-param'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { pathId as pathIdRegexp } from '@ttn-lw/lib/regexp'

import { mayConfigurePacketBroker } from '@console/lib/feature-checks'

import { getPacketBrokerInfo } from '@console/store/actions/packet-broker'

import PacketBroker from './admin-packet-broker'
import NetworkRoutingPolicy from './network-routing-policy'

const PacketBrokerRouter = () => {
  useBreadcrumbs(
    'admin-panel.packet-broker',
    <Breadcrumb
      path={'/admin-panel/packet-broker/routing-configuration'}
      content={sharedMessages.packetBroker}
    />,
  )

  return (
    <Require featureCheck={mayConfigurePacketBroker} otherwise={{ redirect: '/' }}>
      <RequireRequest requestAction={getPacketBrokerInfo()}>
        <Routes>
          <Route
            path="routing-configuration/networks/:netId/:tenantId?"
            element={
              <ValidateRouteParam
                check={{ tenantId: pathIdRegexp, netId: /^\d+$/ }}
                Component={NetworkRoutingPolicy}
              />
            }
          />
          <Route path="*" Component={PacketBroker} />
        </Routes>
      </RequireRequest>
    </Require>
  )
}

export default PacketBrokerRouter
