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

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import DeviceEvents from '@console/containers/device-events'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { selectSelectedDevice, selectSelectedDeviceId } from '@console/store/selectors/devices'

@connect(function(state, props) {
  const device = selectSelectedDevice(state)
  return {
    device,
    devId: selectSelectedDeviceId(state),
    devIds: device && device.ids,
  }
})
@withBreadcrumb('device.single.data', function(props) {
  const { devId } = props
  const { appId } = props.match.params
  return (
    <Breadcrumb
      path={`/applications/${appId}/devices/${devId}/data`}
      content={sharedMessages.data}
    />
  )
})
export default class Data extends React.Component {
  static propTypes = {
    device: PropTypes.device.isRequired,
  }

  render() {
    const {
      device: { ids },
    } = this.props

    return (
      <Container>
        <IntlHelmet hideHeading title={sharedMessages.data} />
        <Row>
          <Col>
            <DeviceEvents devIds={ids} />
          </Col>
        </Row>
      </Container>
    )
  }
}
