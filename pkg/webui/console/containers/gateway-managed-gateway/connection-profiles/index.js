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
import { Col, Container, Row } from 'react-grid-system'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import GatewayConnectionProfilesForm from '@console/containers/gateway-managed-gateway/connection-profiles/form'
import GatewayConnectionProfilesOverview from '@console/containers/gateway-managed-gateway/connection-profiles/overview'
import { CONNECTION_TYPES } from '@console/containers/gateway-managed-gateway/utils'

import sharedMessages from '@ttn-lw/lib/shared-messages'

const GatewayConnectionProfiles = () => {
  const { gtwId, type } = useParams()
  useBreadcrumbs(
    'gtws.single.managed-gateway.connection-profiles',
    <Breadcrumb
      path={`/gateways/${gtwId}/managed-gateway/connection-profiles/${type}`}
      content={sharedMessages.connectionProfiles}
    />,
  )

  if (!Object.values(CONNECTION_TYPES).includes(type)) {
    return (
      <Navigate
        to={`/gateways/${gtwId}/managed-gateway/connection-profiles/${CONNECTION_TYPES.WIFI}`}
        replace
      />
    )
  }

  return (
    <Row>
      <Col sm={12} lg={8}>
        <Routes>
          <Route index Component={GatewayConnectionProfilesOverview} />
          <Route path="add" Component={GatewayConnectionProfilesForm} />
          <Route path="edit/:profileId" Component={GatewayConnectionProfilesForm} />
          <Route path="*" element={<Navigate to="" replace />} />
        </Routes>
      </Col>
    </Row>
  )
}

export default GatewayConnectionProfiles
