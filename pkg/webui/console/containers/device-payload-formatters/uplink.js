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
import bind from 'autobind-decorator'
import { connect } from 'react-redux'

import PropTypes from '../../../lib/prop-types'
import sharedMessages from '../../../lib/shared-messages'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import PayloadFormattersForm from '../../components/payload-formatters-form'
import IntlHelmet from '../../../lib/components/intl-helmet'
import PAYLOAD_FORMATTER_TYPES from '../../constants/formatter-types'
import toast from '../../../components/toast'

import { updateDevice } from '../../store/actions/device'
import { attachPromise } from '../../store/actions/lib'
import { selectSelectedApplicationId } from '../../store/selectors/applications'
import {
  selectSelectedDeviceId,
  selectSelectedDeviceFormatters,
  selectUpdateDeviceError,
} from '../../store/selectors/device'

@connect(function (state) {
  const formatters = selectSelectedDeviceFormatters(state) || {}

  return {
    appId: selectSelectedApplicationId(state),
    devId: selectSelectedDeviceId(state),
    error: selectUpdateDeviceError(state),
    formatters,
  }
},
{ updateDevice: attachPromise(updateDevice) })
@withBreadcrumb('device.single.payload-formatters.uplink', function (props) {
  const { appId, devId } = props

  return (
    <Breadcrumb
      path={`/applications/${appId}/devices/${devId}/payload-formatters/uplink`}
      icon="uplink"
      content={sharedMessages.uplink}
    />
  )
})
@bind
class DevicePayloadFormatters extends React.PureComponent {

  static propTypes = {
    appId: PropTypes.string.isRequired,
    devId: PropTypes.string.isRequired,
    formatters: PropTypes.object.isRequired,
    updateDevice: PropTypes.func.isRequired,
  }

  async onSubmit (values) {
    const { appId, devId, formatters, updateDevice } = this.props
    return updateDevice(appId, devId, {
      formatters: {
        down_formatter: formatters.down_formatter || PAYLOAD_FORMATTER_TYPES.NONE,
        down_formatter_parameter: formatters.down_formatter_parameter || '',
        up_formatter: values.type,
        up_formatter_parameter: values.parameter,
      },
    })
  }

  async onSubmitSuccess () {
    const { devId } = this.props
    toast({
      title: devId,
      message: sharedMessages.payloadFormattersUpdateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  render () {
    const { formatters, error } = this.props

    return (
      <React.Fragment>
        <IntlHelmet title={sharedMessages.payloadFormattersUplink} />
        <PayloadFormattersForm
          uplink
          linked
          error={error}
          onSubmit={this.onSubmit}
          onSubmitSuccess={this.onSubmitSuccess}
          title={sharedMessages.payloadFormattersUplink}
          initialType={formatters.up_formatter || PAYLOAD_FORMATTER_TYPES.NONE}
          initialParameter={formatters.up_formatter_parameter || ''}
        />
      </React.Fragment>
    )
  }
}

export default DevicePayloadFormatters
