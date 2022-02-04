// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import PAYLOAD_FORMATTER_TYPES from '@console/constants/formatter-types'
import tts from '@console/api/tts'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import toast from '@ttn-lw/components/toast'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import withRequest from '@ttn-lw/lib/components/with-request'

import PayloadFormattersForm from '@console/components/payload-formatters-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { hexToBase64 } from '@console/lib/bytes'

import { updateDevice } from '@console/store/actions/devices'
import { getRepositoryPayloadFormatters } from '@console/store/actions/device-repository'

import {
  selectSelectedApplicationId,
  selectApplicationLink,
} from '@console/store/selectors/applications'
import {
  selectSelectedDeviceId,
  selectSelectedDeviceFormatters,
  selectSelectedDevice,
} from '@console/store/selectors/devices'
import { selectDeviceRepoPayloadFromatters } from '@console/store/selectors/device-repository'

@connect(
  state => ({
    appId: selectSelectedApplicationId(state),
    device: selectSelectedDevice(state),
    devId: selectSelectedDeviceId(state),
    link: selectApplicationLink(state),
    formatters: selectSelectedDeviceFormatters(state),
    decodeUplink: tts.As.decodeUplink,
    repositoryPayloadFormatters: selectDeviceRepoPayloadFromatters(state),
  }),
  dispatch => ({
    updateDevice: attachPromise(updateDevice),
    getRepositoryPayloadFormatters: (appId, versionIds) =>
      dispatch(getRepositoryPayloadFormatters(appId, versionIds)),
  }),
)
@withRequest(({ appId, device, getRepositoryPayloadFormatters }) =>
  getRepositoryPayloadFormatters(appId, device.version_ids),
)
@withBreadcrumb('device.single.payload-formatters.uplink', props => {
  const { appId, devId } = props

  return (
    <Breadcrumb
      path={`/applications/${appId}/devices/${devId}/payload-formatters/uplink`}
      content={sharedMessages.uplink}
    />
  )
})
class DevicePayloadFormatters extends React.PureComponent {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    decodeUplink: PropTypes.func.isRequired,
    devId: PropTypes.string.isRequired,
    device: PropTypes.device.isRequired,
    formatters: PropTypes.formatters,
    link: PropTypes.shape({
      default_formatters: PropTypes.shape({
        up_formatter: PropTypes.string,
        up_formatter_parameter: PropTypes.string,
      }),
    }).isRequired,
    repositoryPayloadFormatters: PropTypes.shape({
      formatter_parameter: PropTypes.string,
    }),
    updateDevice: PropTypes.func.isRequired,
  }

  static defaultProps = {
    formatters: undefined,
    repositoryPayloadFormatters: undefined,
  }

  constructor(props) {
    super(props)

    const { formatters } = props

    this.state = {
      type: Boolean(formatters)
        ? formatters.up_formatter || PAYLOAD_FORMATTER_TYPES.NONE
        : PAYLOAD_FORMATTER_TYPES.DEFAULT,
    }
  }

  @bind
  async onSubmit(values) {
    const { appId, devId, formatters: initialFormatters, updateDevice } = this.props

    if (values.type === PAYLOAD_FORMATTER_TYPES.DEFAULT) {
      return updateDevice(appId, devId, {
        formatters: null,
      })
    }

    const formatters = { ...(initialFormatters || {}) }

    return updateDevice(appId, devId, {
      formatters: {
        down_formatter: formatters.down_formatter || PAYLOAD_FORMATTER_TYPES.NONE,
        down_formatter_parameter: formatters.down_formatter_parameter,
        up_formatter: values.type,
        up_formatter_parameter: values.parameter,
      },
    })
  }

  @bind
  async onSubmitSuccess() {
    const { devId } = this.props
    toast({
      title: devId,
      message: sharedMessages.payloadFormattersUpdateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  @bind
  async onTestSubmit(data) {
    const { appId, devId, decodeUplink, device } = this.props
    const { f_port, payload, formatter, parameter } = data
    const { version_ids } = device

    const { uplink } = await decodeUplink(appId, devId, {
      uplink: {
        f_port,
        frm_payload: hexToBase64(payload),
        // `rx_metadata` and `settings` fields are required by the validation middleware in AS.
        // These fields won't affect the result of decoding an uplink message.
        rx_metadata: [
          { gateway_ids: { gateway_id: 'test' }, rssi: 42, channel_rssi: 42, snr: 4.2 },
        ],
        settings: { data_rate: { lora: { bandwidth: 125000, spreading_factor: 7 } } },
      },
      version_ids: Object.keys(version_ids).length > 0 ? version_ids : undefined,
      formatter,
      parameter,
    })

    return {
      payload: uplink.decoded_payload,
      warnings: uplink.decoded_payload_warnings,
    }
  }

  @bind
  onTypeChange(type) {
    this.setState({ type })
  }

  render() {
    const { formatters, link, repositoryPayloadFormatters } = this.props
    const { type } = this.state
    const { default_formatters = {} } = link

    const formatterType = Boolean(formatters)
      ? formatters.up_formatter || PAYLOAD_FORMATTER_TYPES.NONE
      : PAYLOAD_FORMATTER_TYPES.DEFAULT
    const formatterParameter = Boolean(formatters) ? formatters.up_formatter_parameter : undefined
    const appFormatterType = Boolean(default_formatters.up_formatter)
      ? default_formatters.up_formatter
      : PAYLOAD_FORMATTER_TYPES.NONE
    const appFormatterParameter = Boolean(default_formatters.up_formatter_parameter)
      ? default_formatters.up_formatter_parameter
      : undefined

    const isDefaultType = type === PAYLOAD_FORMATTER_TYPES.DEFAULT

    return (
      <React.Fragment>
        <IntlHelmet title={sharedMessages.payloadFormattersUplink} />
        <PayloadFormattersForm
          uplink
          linked
          allowReset
          allowTest
          onSubmit={this.onSubmit}
          onSubmitSuccess={this.onSubmitSuccess}
          onTestSubmit={this.onTestSubmit}
          title={sharedMessages.payloadFormattersUplink}
          initialType={formatterType}
          initialParameter={formatterParameter}
          defaultType={appFormatterType}
          defaultParameter={appFormatterParameter}
          onTypeChange={this.onTypeChange}
          isDefaultType={isDefaultType}
          repoFormatters={repositoryPayloadFormatters}
        />
      </React.Fragment>
    )
  }
}

export default DevicePayloadFormatters
