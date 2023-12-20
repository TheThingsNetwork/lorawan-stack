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
import { Routes, Route, Navigate, useParams } from 'react-router-dom'
import { Container, Col, Row } from 'react-grid-system'
import { useSelector } from 'react-redux'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import ApplicationUplinkPayloadFormatters from '@console/containers/application-payload-formatters/uplink'
import ApplicationDownlinkPayloadFormatters from '@console/containers/application-payload-formatters/downlink'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  maySetApplicationPayloadFormatters,
  mayViewApplicationLink,
} from '@console/lib/feature-checks'
import { checkFromState } from '@account/lib/feature-checks'

import { getApplicationLink } from '@console/store/actions/link'

const ApplicationPayloadFormatters = () => {
  const { appId } = useParams()
  const mayViewLink = useSelector(state => checkFromState(mayViewApplicationLink, state))
  const getLink = mayViewLink ? getApplicationLink(appId, ['default_formatters']) : []

  return (
    <Require
      featureCheck={maySetApplicationPayloadFormatters}
      otherwise={{ redirect: `/applications/${appId}` }}
    >
      <RequireRequest requestAction={getLink}>
        <ApplicationPayloadFormattersInner />
      </RequireRequest>
    </Require>
  )
}

const ApplicationPayloadFormattersInner = () => {
  const { appId } = useParams()

  useBreadcrumbs(
    'apps.single.payload-formatters',
    <Breadcrumb
      path={`/applications/${appId}/payload-formatters`}
      content={sharedMessages.payloadFormatters}
    />,
  )

  return (
    <Container>
      <Row>
        <Col>
          <Routes>
            <Route index element={<Navigate to="uplink" replace />} />
            <Route path="uplink" Component={ApplicationUplinkPayloadFormatters} />
            <Route path="downlink" Component={ApplicationDownlinkPayloadFormatters} />
          </Routes>
        </Col>
      </Row>
    </Container>
  )
}

export default ApplicationPayloadFormatters
