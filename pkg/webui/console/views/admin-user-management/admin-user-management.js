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

import React, { Component } from 'react'
import { Container, Col, Row } from 'react-grid-system'

import UsersTable from '../../containers/users-table'
import sharedMessages from '../../../lib/shared-messages'
import IntlHelmet from '../../../lib/components/intl-helmet'

import PAGE_SIZES from '../../constants/page-sizes'

export default class UserManagement extends Component {
  render() {
    return (
      <Container>
        <Row>
          <Col>
            <IntlHelmet title={sharedMessages.userManagement} />
            <UsersTable pageSize={PAGE_SIZES.REGULAR} />
          </Col>
        </Row>
      </Container>
    )
  }
}
