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
import { replace } from 'connected-react-router'
import { bindActionCreators } from 'redux'
import { connect } from 'react-redux'
import { Col, Row, Container } from 'react-grid-system'
import bind from 'autobind-decorator'

import sharedMessages from '../../../lib/shared-messages'
import diff from '../../../lib/diff'
import DeviceDataForm from '../../components/device-data-form'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import IntlHelmet from '../../../lib/components/intl-helmet'
import PropTypes from '../../../lib/prop-types'
import api from '../../api'

import { updateDevice } from '../../store/actions/device'
import { attachPromise } from '../../store/actions/lib'
import { selectSelectedApplicationId } from '../../store/selectors/applications'
import { selectSelectedDevice, selectSelectedDeviceId } from '../../store/selectors/device'

@connect(
  state => ({
    device: selectSelectedDevice(state),
    devId: selectSelectedDeviceId(state),
    appId: selectSelectedApplicationId(state),
  }),
  dispatch => ({
    ...bindActionCreators({ updateDevice: attachPromise(updateDevice) }, dispatch),
    onDeleteSuccess: appId => dispatch(replace(`/applications/${appId}/devices`)),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    onDeleteSuccess: () => dispatchProps.onDeleteSuccess(stateProps.appId),
  }),
)
@withBreadcrumb('device.single.general-settings', function(props) {
  const { devId, appId } = props
  return (
    <Breadcrumb
      path={`/applications/${appId}/devices/${devId}/general-settings`}
      icon="general_settings"
      content={sharedMessages.generalSettings}
    />
  )
})
export default class DeviceGeneralSettings extends React.Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    device: PropTypes.device.isRequired,
    onDeleteSuccess: PropTypes.func.isRequired,
    updateDevice: PropTypes.func.isRequired,
  }

  state = {
    error: '',
  }

  @bind
  async handleSubmit(values) {
    const { device, appId, updateDevice } = this.props
    const { activation_mode, ...updatedDevice } = values

    const {
      ids: { device_id: deviceId },
    } = device
    const changed = diff(device, updatedDevice, ['updated_at', 'created_at'])

    return updateDevice(appId, deviceId, changed)
  }

  @bind
  async handleDelete() {
    const { appId, device } = this.props
    const {
      ids: { device_id: deviceId },
    } = device

    return api.device.delete(appId, deviceId)
  }

  render() {
    const { device: initialValues, onDeleteSuccess } = this.props
    const { error } = this.state

    return (
      <Container>
        <IntlHelmet title={sharedMessages.generalSettings} />
        <Row>
          <Col lg={8} md={12}>
            <DeviceDataForm
              error={error}
              onSubmit={this.handleSubmit}
              onDelete={this.handleDelete}
              onDeleteSuccess={onDeleteSuccess}
              initialValues={initialValues}
              update
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
