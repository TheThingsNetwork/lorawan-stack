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
import { defineMessages } from 'react-intl'

import PAYLOAD_FORMATTER_TYPES from '@console/constants/formatter-types'

import api from '@console/api'

import Notification from '@ttn-lw/components/notification'
import PageTitle from '@ttn-lw/components/page-title'
import toast from '@ttn-lw/components/toast'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'

import PayloadFormattersForm from '@console/components/payload-formatters-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { updateApplicationLinkSuccess } from '@console/store/actions/link'

import {
  selectApplicationIsLinked,
  selectApplicationLinkFormatters,
  selectSelectedApplicationId,
} from '@console/store/selectors/applications'

import style from './application-payload-formatters.styl'

const m = defineMessages({
  infoText:
    'These payload formatters are executed on uplink messages from all end devices in this application. Note: end device level payload formatters have precedence.',
})
@connect(
  function(state) {
    const formatters = selectApplicationLinkFormatters(state) || {}

    return {
      appId: selectSelectedApplicationId(state),
      linked: selectApplicationIsLinked(state) || false,
      formatters,
    }
  },
  { updateLinkSuccess: updateApplicationLinkSuccess },
)
@withBreadcrumb('apps.single.payload-formatters.uplink', function(props) {
  const { appId } = props

  return (
    <Breadcrumb
      path={`/applications/${appId}/payload-formatters/uplink`}
      content={sharedMessages.uplink}
    />
  )
})
@bind
class ApplicationPayloadFormatters extends React.PureComponent {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    formatters: PropTypes.formatters.isRequired,
    linked: PropTypes.bool.isRequired,
    updateLinkSuccess: PropTypes.func.isRequired,
  }

  async onSubmit(values) {
    const { appId, formatters } = this.props

    return await api.application.link.set(appId, {
      default_formatters: {
        down_formatter: formatters.down_formatter || PAYLOAD_FORMATTER_TYPES.NONE,
        down_formatter_parameter: formatters.down_formatter_parameter || '',
        up_formatter: values.type,
        up_formatter_parameter: values.parameter,
      },
    })
  }

  onSubmitSuccess(link) {
    const { appId, updateLinkSuccess } = this.props
    toast({
      title: appId,
      message: sharedMessages.payloadFormattersUpdateSuccess,
      type: toast.types.SUCCESS,
    })
    updateLinkSuccess(link)
  }

  render() {
    const { formatters, linked } = this.props

    const applicationFormatterInfo = (
      <Notification className={style.notification} small info content={m.infoText} />
    )

    return (
      <React.Fragment>
        <PageTitle title={sharedMessages.payloadFormattersUplink} />
        {linked && applicationFormatterInfo}
        <PayloadFormattersForm
          uplink
          linked={linked}
          onSubmit={this.onSubmit}
          onSubmitSuccess={this.onSubmitSuccess}
          initialType={formatters.up_formatter || PAYLOAD_FORMATTER_TYPES.NONE}
          initialParameter={formatters.up_formatter_parameter || ''}
        />
      </React.Fragment>
    )
  }
}

export default ApplicationPayloadFormatters
