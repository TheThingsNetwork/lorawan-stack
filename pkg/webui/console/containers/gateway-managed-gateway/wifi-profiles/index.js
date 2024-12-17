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
import { Route, Routes, useParams } from 'react-router-dom'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'
import ValidateRouteParam from '@ttn-lw/lib/components/validate-route-param'

import GatewayWifiProfilesForm from '@console/containers/gateway-managed-gateway/wifi-profiles/form'
import GatewayWifiProfilesOverview from '@console/containers/gateway-managed-gateway/wifi-profiles/overview'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { profileIdPath } from '@console/lib/regexp'

const GatewayWifiProfiles = () => {
  const { gtwId } = useParams()
  useBreadcrumbs(
    'gtws.single.managed-gateway.wifi-profiles',
    <Breadcrumb
      path={`/gateways/${gtwId}/managed-gateway/wifi-profiles`}
      content={sharedMessages.wifiProfiles}
    />,
  )

  return (
    <div className="item-12">
      <Routes>
        <Route index Component={GatewayWifiProfilesOverview} />
        <Route path="add" Component={GatewayWifiProfilesForm} />
        <Route
          path="edit/:profileId"
          element={
            <ValidateRouteParam
              check={{ profileId: profileIdPath }}
              Component={GatewayWifiProfilesForm}
              otherwise={{ redirect: `/gateways/${gtwId}/managed-gateway/wifi-profiles` }}
            />
          }
        />
        <Route path="*" element={<GenericNotFound />} />
      </Routes>
    </div>
  )
}

export default GatewayWifiProfiles
