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
import api from '../../api'

import { updateDevice } from '../../store/actions/device'
import { attachPromise } from '../../store/actions/lib'
import { selectSelectedApplicationId } from '../../store/selectors/applications'
import { selectSelectedDevice, selectSelectedDeviceId } from '../../store/selectors/device'

@connect(
  function(state) {
    return {
      device: selectSelectedDevice(state),
      devId: selectSelectedDeviceId(state),
      appId: selectSelectedApplicationId(state),
    }
  },
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
@bind
export default class DeviceGeneralSettings extends React.Component {
  state = {
    error: '',
  }

  async handleSubmit(values) {
    const { device, appId, updateDevice } = this.props
    const { activation_mode, ...updatedDevice } = values

    // Clean values based on activation mode
    if (activation_mode === 'otaa') {
      delete updatedDevice.mac_settings
      delete updatedDevice.session
    } else {
      delete updatedDevice.ids.join_eui
      delete updatedDevice.ids.dev_eui
      delete updatedDevice.root_keys
      delete updatedDevice.resets_join_nonces
    }

    const {
      ids: { device_id: deviceId },
    } = this.props.device
    const changed = diff(device, updatedDevice, ['updated_at', 'created_at'])

    return updateDevice(appId, deviceId, changed)
  }

  async handleDelete() {
    const { appId, device } = this.props
    const {
      ids: { device_id: deviceId },
    } = device

    return api.device.delete(appId, deviceId)
  }

  render() {
    const { device, onDeleteSuccess } = this.props
    const { error } = this.state

    return (
      <Container>
        <IntlHelmet title={sharedMessages.generalSettings} />
        <Row>
          <Col sm={12} md={12} lg={8} xl={8}>
            <DeviceDataForm
              error={error}
              onSubmit={this.handleSubmit}
              onDelete={this.handleDelete}
              onDeleteSuccess={onDeleteSuccess}
              initialValues={device}
              update
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
