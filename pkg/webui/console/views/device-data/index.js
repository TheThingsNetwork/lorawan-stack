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
import bind from 'autobind-decorator'

import IntlHelmet from '../../../lib/components/intl-helmet'
import sharedMessages from '../../../lib/shared-messages'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import DeviceEvents from '../../containers/device-events'

import { getDeviceId } from '../../../lib/selectors/id'

@connect(function({ device, application }, props) {
  return {
    device: device.device,
    devId: getDeviceId(device.device),
    devIds: device.device && device.device.ids,
  }
})
@withBreadcrumb('device.single.data', function(props) {
  const { devId } = props
  const { appId } = props.match.params
  return (
    <Breadcrumb
      path={`/applications/${appId}/devices/${devId}/data`}
      icon="data"
      content={sharedMessages.data}
    />
  )
})
@bind
export default class Data extends React.Component {
  render() {
    const { devIds } = this.props

    return (
      <Container>
        <IntlHelmet hideHeading title={sharedMessages.data} />
        <Row>
          <Col>
            <DeviceEvents devIds={devIds} />
          </Col>
        </Row>
      </Container>
    )
  }
}
