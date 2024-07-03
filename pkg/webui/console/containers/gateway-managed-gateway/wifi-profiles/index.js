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
import { Navigate, Route, Routes, useParams } from 'react-router-dom'
import { Col, Row } from 'react-grid-system'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import GatewayWifiProfilesForm from '@console/containers/gateway-managed-gateway/wifi-profiles/form'
import GatewayConnectionProfilesOverview from '@console/containers/gateway-managed-gateway/wifi-profiles/overview'

import sharedMessages from '@ttn-lw/lib/shared-messages'

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
    <Row>
      <Col sm={12} lg={8}>
        <Routes>
          <Route index Component={GatewayConnectionProfilesOverview} />
          <Route path="add" Component={GatewayWifiProfilesForm} />
          <Route path="edit/:profileId" Component={GatewayWifiProfilesForm} />
          <Route path="*" element={<Navigate to="" replace />} />
        </Routes>
      </Col>
    </Row>
  )
}

export default GatewayWifiProfiles
