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
import { Routes, Route, Navigate, useParams } from 'react-router-dom'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import GatewayConnectionSettings from '@console/containers/gateway-managed-gateway/connection-settings'
import GatewayWifiProfiles from '@console/containers/gateway-managed-gateway/wifi-profiles'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayViewManagedGateway } from '@console/lib/feature-checks'

const GatewayManagedGateway = () => {
  const { gtwId } = useParams()

  useBreadcrumbs(
    'gtws.single.managed-gateway',
    <Breadcrumb
      path={`/gateways/${gtwId}/managed-gateway`}
      content={sharedMessages.managedGateway}
    />,
  )

  return (
    <Require featureCheck={mayViewManagedGateway} otherwise={{ redirect: `/gateways/${gtwId}` }}>
      <div className="container container--xxl grid">
        <Routes>
          <Route index element={<Navigate to="connection-settings" replace />} />
          <Route path="connection-settings" Component={GatewayConnectionSettings} />
          <Route path="wifi-profiles/*" Component={GatewayWifiProfiles} />
          <Route path="*" element={<Navigate to="connection-settings" replace />} />
        </Routes>
      </div>
    </Require>
  )
}

export default GatewayManagedGateway
