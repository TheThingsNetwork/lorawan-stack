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
import { Container, Col, Row } from 'react-grid-system'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import GatewayConnectionSettings from '@console/containers/gateway-the-things-station/connection-settings'
import GatewayConnectionProfiles from '@console/containers/gateway-the-things-station/connection-profiles'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayViewTheThingsStation } from '@console/lib/feature-checks'

const GatewayTheThingsStation = () => {
  const { gtwId } = useParams()

  useBreadcrumbs(
    'gtws.single.the-things-station',
    <Breadcrumb
      path={`/gateways/${gtwId}/the-things-station`}
      content={sharedMessages.theThingsStation}
    />,
  )

  return (
    <Require featureCheck={mayViewTheThingsStation} otherwise={{ redirect: `/gateways/${gtwId}` }}>
      <Container>
        <Row>
          <Col>
            <Routes>
              <Route index element={<Navigate to="connection-settings" replace />} />
              <Route path="connection-settings" Component={GatewayConnectionSettings} />
              <Route path="connection-profiles/:type/*" Component={GatewayConnectionProfiles} />
              <Route path="*" element={<Navigate to="connection-settings" replace />} />
            </Routes>
          </Col>
        </Row>
      </Container>
    </Require>
  )
}

export default GatewayTheThingsStation
