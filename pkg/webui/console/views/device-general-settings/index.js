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
import { defineMessages } from 'react-intl'

import sharedMessages from '../../../lib/shared-messages'
import diff from '../../../lib/diff'

import DeviceDataForm from '../../containers/device-data-form'
import IntlHelmet from '../../../lib/components/intl-helmet'
import toast from '../../../components/toast'

import { updateDevice } from '../../store/actions/device'

import { selectSelectedApplicationId } from '../../store/selectors/applications'

const m = defineMessages({
  updateSuccess: 'Successfully updated end device',
})

@connect(function (state) {
  return {
    device: state.device.device,
    appId: selectSelectedApplicationId(state),
  }
}, { updateDevice })
@bind
export default class DeviceGeneralSettings extends React.Component {

  state = {
    error: '',
  }

  async handleSubmit (values, { setSubmitting, resetForm }) {
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

    await this.setState({ error: '' })
    try {
      const { ids: { device_id: deviceId }} = this.props.device
      const changed = diff(device, updatedDevice)

      await updateDevice(appId, deviceId, changed)
      resetForm()
      toast({
        title: deviceId,
        message: m.updateSuccess,
        type: toast.types.SUCCESS,
      })
    } catch (error) {
      resetForm()
      const err = error instanceof Error ? sharedMessages.genericError : error
      await this.setState({ error: err })
    }
  }

  render () {
    const { device } = this.props
    const { error } = this.state

    return (
      <Container>
        <IntlHelmet
          title={sharedMessages.generalSettings}
        />
        <Row>
          <Col sm={12} md={12} lg={8} xl={8}>
            <DeviceDataForm
              error={error}
              onSubmit={this.handleSubmit}
              initialValues={device}
              update
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
