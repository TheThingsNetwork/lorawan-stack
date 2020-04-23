// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import { connect } from 'react-redux'
import { Container, Col, Row } from 'react-grid-system'

import IntlHelmet from '../../../lib/components/intl-helmet'
import DownlinkForm from '../../components/downlink-form'
import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'
import withFeatureRequirement from '../../lib/components/with-feature-requirement'
import { mayScheduleDownlinks } from '../../lib/feature-checks'
import { selectSelectedApplicationId } from '../../store/selectors/applications'
import { selectSelectedDeviceId } from '../../store/selectors/devices'

@connect(state => ({
  appId: selectSelectedApplicationId(state),
  devId: selectSelectedDeviceId(state),
}))
@withFeatureRequirement(mayScheduleDownlinks, {
  redirect: ({ appId, devId }) => `/applications/${appId}/devices/${devId}`,
})
export default class DeviceMessages extends Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    devId: PropTypes.string.isRequired,
  }

  render() {
    const { appId, devId } = this.props
    return (
      <Container>
        <IntlHelmet title={sharedMessages.messages} />
        <Row>
          <Col lg={8} md={12}>
            <DownlinkForm appId={appId} devId={devId} />
          </Col>
        </Row>
      </Container>
    )
  }
}
