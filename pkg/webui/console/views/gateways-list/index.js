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
import { Container, Row, Col } from 'react-grid-system'

import PAGE_SIZES from '@ttn-lw/constants/page-sizes'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import GatewaysTable from '@console/containers/gateways-table'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import getCookie from '@console/lib/table-utils'

const GatewaysList = () => {
  const gtwPageSize = getCookie('gateways-list-page-size')
  const gtwParam = `?page-size=${gtwPageSize ? gtwPageSize : PAGE_SIZES.REGULAR}`
  useBreadcrumbs(
    'gtws.list',
    <Breadcrumb path={`/gateways${gtwParam}`} content={sharedMessages.list} />,
  )

  return (
    <Container>
      <Row>
        <IntlHelmet title={sharedMessages.gateways} />
        <Col>
          <GatewaysTable />
        </Col>
      </Row>
    </Container>
  )
}

export default GatewaysList
