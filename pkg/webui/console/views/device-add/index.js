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
import bind from 'autobind-decorator'
import { connect } from 'react-redux'
import { push } from 'connected-react-router'

import api from '@console/api'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'

import DeviceDataForm from '@console/components/device-data-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { getDeviceId } from '@ttn-lw/lib/selectors/id'
import PropTypes from '@ttn-lw/lib/prop-types'

import { mayEditApplicationDeviceKeys, checkFromState } from '@console/lib/feature-checks'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'

@connect(
  state => ({
    appId: selectSelectedApplicationId(state),
    mayEditKeys: checkFromState(mayEditApplicationDeviceKeys, state),
  }),
  dispatch => ({
    redirectToList: (appId, deviceId) =>
      dispatch(push(`/applications/${appId}/devices/${deviceId}`)),
  }),
)
@withBreadcrumb('devices.add', function(props) {
  const { appId } = props
  return <Breadcrumb path={`/applications/${appId}/devices/add`} content={sharedMessages.add} />
})
export default class DeviceAdd extends Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    mayEditKeys: PropTypes.bool.isRequired,
    redirectToList: PropTypes.func.isRequired,
  }

  @bind
  async handleSubmit(device) {
    const { appId } = this.props

    return api.device.create(appId, device)
  }

  @bind
  handleSubmitSuccess(device) {
    const { appId, redirectToList } = this.props
    const deviceId = getDeviceId(device)

    redirectToList(appId, deviceId)
  }

  render() {
    const { mayEditKeys } = this.props

    return (
      <Container>
        <PageTitle title={sharedMessages.addDevice} />
        <Row>
          <Col lg={8} md={12}>
            <DeviceDataForm
              onSubmit={this.handleSubmit}
              onSubmitSuccess={this.handleSubmitSuccess}
              mayEditKeys={mayEditKeys}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
