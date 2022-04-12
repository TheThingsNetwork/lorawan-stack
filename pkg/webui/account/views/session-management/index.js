// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
import { Container, Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import PAGE_SIZES from '@ttn-lw/constants/page-sizes'

import PageTitle from '@ttn-lw/components/page-title'

import UserSessionsTable from '@account/containers/sessions-table'

const m = defineMessages({
  sessionManagement: 'Session management',
})

const SessionManagement = () => (
  <Container>
    <Row>
      <Col>
        <PageTitle title={m.sessionManagement} hideHeading />
        <UserSessionsTable pageSize={PAGE_SIZES.REGULAR} />
      </Col>
    </Row>
  </Container>
)

export default SessionManagement
