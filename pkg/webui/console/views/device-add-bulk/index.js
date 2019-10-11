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
import { connect } from 'react-redux'

import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import BulkDeviceCreator from '../../containers/device-bulk-creator'
import sharedMessages from '../../../lib/shared-messages'
import { selectSelectedApplicationId } from '../../store/selectors/applications'

import style from './device-add-bulk.styl'

@connect(state => ({
  appId: selectSelectedApplicationId(state),
}))
@withBreadcrumb('devices.add.bulk', function(props) {
  const { appId } = props
  return (
    <Breadcrumb
      path={`/applications/${appId}/devices/add/bulk`}
      icon="bulk_creation"
      content={sharedMessages.bulk}
    />
  )
})
export default class DeviceAddBulk extends Component {
  render() {
    return (
      <Container>
        <Row>
          <Col>
            <IntlHelmet title={sharedMessages.addDeviceBulk} />
            <Message
              className={style.title}
              component="h2"
              content={sharedMessages.addDeviceBulk}
            />
          </Col>
        </Row>
        <Row>
          <Col lg={8} md={12}>
            <BulkDeviceCreator />
          </Col>
        </Row>
      </Container>
    )
  }
}
