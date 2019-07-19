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
import { connect } from 'react-redux'
import { Col, Row, Container } from 'react-grid-system'

import IntlHelmet from '../../../lib/components/intl-helmet'
import sharedMessages from '../../../lib/shared-messages'
import DevicesTable from '../../containers/devices-table'

import { selectSelectedApplication } from '../../store/selectors/applications'

import PAGE_SIZES from '../../constants/page-sizes'

@connect(function (state, props) {
  return {
    application: selectSelectedApplication(state),
  }
})
class ApplicationDeviceList extends React.Component {
  render () {
    return (
      <Container>
        <Row>
          <IntlHelmet title={sharedMessages.devices} />
          <Col sm={12}>
            <DevicesTable pageSize={PAGE_SIZES.REGULAR} />
          </Col>
        </Row>
      </Container>
    )
  }
}

export default ApplicationDeviceList
