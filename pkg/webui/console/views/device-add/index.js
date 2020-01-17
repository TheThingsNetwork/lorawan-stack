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

import PageTitle from '../../../components/page-title'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import DeviceDataForm from '../../components/device-data-form'
import sharedMessages from '../../../lib/shared-messages'
import { selectSelectedApplicationId } from '../../store/selectors/applications'
import { getDeviceId } from '../../../lib/selectors/id'
import PropTypes from '../../../lib/prop-types'
import api from '../../api'
import style from './device-add-single.styl'

@connect(
  state => ({
    appId: selectSelectedApplicationId(state),
  }),
  dispatch => ({
    redirectToList: (appId, deviceId) =>
      dispatch(push(`/applications/${appId}/devices/${deviceId}`)),
  }),
)
@withBreadcrumb('devices.add', function(props) {
  const { appId } = props
  return (
    <Breadcrumb
      path={`/applications/${appId}/devices/add`}
      icon="add"
      content={sharedMessages.add}
    />
  )
})
export default class DeviceAdd extends Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    redirectToList: PropTypes.func.isRequired,
  }

  @bind
  async handleSubmit(values) {
    const { appId } = this.props
    const { activation_mode, ...device } = values

    return api.device.create(appId, device, {
      otaa: activation_mode === 'otaa',
    })
  }

  @bind
  handleSubmitSuccess(device) {
    const { appId, redirectToList } = this.props
    const deviceId = getDeviceId(device)

    redirectToList(appId, deviceId)
  }

  render() {
    return (
      <Container>
        <PageTitle title={sharedMessages.addDevice} />
        <Row>
          <Col className={style.form} lg={8} md={12}>
            <DeviceDataForm
              onSubmit={this.handleSubmit}
              onSubmitSuccess={this.handleSubmitSuccess}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
