// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
import { Row, Col, Container } from 'react-grid-system'

import sharedMessages from '../../../lib/shared-messages'
import IntlHelmet from '../../../lib/components/intl-helmet'

import ApplicationsTable from '../../containers/applications-table'

import PAGE_SIZES from '../../constants/page-sizes'

export default class List extends React.Component {
  render () {
    return (
      <Container>
        <Row>
          <IntlHelmet title={sharedMessages.applications} />
          <Col sm={12}>
            <ApplicationsTable pageSize={PAGE_SIZES.REGULAR} />
          </Col>
        </Row>
      </Container>
    )
  }
}
