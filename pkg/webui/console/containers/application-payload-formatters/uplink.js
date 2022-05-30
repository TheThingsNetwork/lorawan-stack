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
import { defineMessages } from 'react-intl'

import PAYLOAD_FORMATTER_TYPES from '@console/constants/formatter-types'

import Notification from '@ttn-lw/components/notification'
import PageTitle from '@ttn-lw/components/page-title'
import toast from '@ttn-lw/components/toast'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import ErrorNotification from '@ttn-lw/components/error-notification'

import PayloadFormattersForm from '@console/components/payload-formatters-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import { isNotFoundError } from '@ttn-lw/lib/errors/utils'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { checkFromState, mayViewApplicationLink } from '@console/lib/feature-checks'

import { updateApplicationLink, updateApplicationLinkSuccess } from '@console/store/actions/link'

import {
  selectApplicationLinkError,
  selectApplicationLinkFormatters,
  selectSelectedApplicationId,
} from '@console/store/selectors/applications'

import style from './application-payload-formatters.styl'

const m = defineMessages({
  title: 'Default uplink payload formatter',
  infoText:
    'You can use the "Payload formatter" tab of individual end devices to test uplink payload formatters and to define individual payload formatter settings per end device.',
  uplinkResetWarning:
    'You do not have sufficient rights to view the current uplink payload formatter. Only overwriting is allowed.',
})
@connect(
  state => {
    const formatters = selectApplicationLinkFormatters(state) || {}
    const mayViewLink = checkFromState(mayViewApplicationLink, state)
    const linkError = selectApplicationLinkError(state)

    return {
      appId: selectSelectedApplicationId(state),
      formatters,
      mayViewLink,
      linkError,
    }
  },
  {
    updateLinkSuccess: updateApplicationLinkSuccess,
    submitPayloadFormatters: (id, pf) => attachPromise(updateApplicationLink(id, pf)),
  },
)
@withBreadcrumb('apps.single.payload-formatters.uplink', props => {
  const { appId } = props

  return (
    <Breadcrumb
      path={`/applications/${appId}/payload-formatters/uplink`}
      content={sharedMessages.uplink}
    />
  )
})
class ApplicationPayloadFormatters extends React.PureComponent {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    formatters: PropTypes.formatters.isRequired,
    linkError: PropTypes.error,
    mayViewLink: PropTypes.bool.isRequired,
    submitPayloadFormatters: PropTypes.func.isRequired,
    updateLinkSuccess: PropTypes.func.isRequired,
  }

  static defaultProps = {
    linkError: undefined,
  }

  constructor(props) {
    super(props)

    const { formatters } = props

    this.state = {
      type: formatters.up_formatter || PAYLOAD_FORMATTER_TYPES.NONE,
    }
  }

  @bind
  async onSubmit(values) {
    const { appId, formatters, submitPayloadFormatters } = this.props

    return await submitPayloadFormatters(appId, {
      default_formatters: {
        down_formatter: formatters.down_formatter || PAYLOAD_FORMATTER_TYPES.NONE,
        down_formatter_parameter: formatters.down_formatter_parameter || '',
        up_formatter: values.type,
        up_formatter_parameter: values.parameter,
      },
    })
  }

  @bind
  onSubmitSuccess(link) {
    const { appId, updateLinkSuccess } = this.props
    toast({
      title: appId,
      message: sharedMessages.payloadFormattersUpdateSuccess,
      type: toast.types.SUCCESS,
    })
    updateLinkSuccess(link)
  }

  @bind
  onTypeChange(type) {
    this.setState({ type })
  }

  render() {
    const { formatters, mayViewLink, linkError } = this.props
    const { type } = this.state

    const isNoneType = type === PAYLOAD_FORMATTER_TYPES.NONE
    const hasError = Boolean(linkError) && !isNotFoundError(linkError)

    return (
      <React.Fragment>
        <PageTitle title={m.title} />
        {hasError && <ErrorNotification content={linkError} small />}
        {!isNoneType && (
          <Notification className={style.notification} small info content={m.infoText} />
        )}
        {!mayViewLink && <Notification content={m.uplinkResetWarning} info small />}
        <PayloadFormattersForm
          uplink
          onSubmit={this.onSubmit}
          onSubmitSuccess={this.onSubmitSuccess}
          initialType={formatters.up_formatter || PAYLOAD_FORMATTER_TYPES.NONE}
          initialParameter={formatters.up_formatter_parameter || ''}
          onTypeChange={this.onTypeChange}
        />
      </React.Fragment>
    )
  }
}

export default ApplicationPayloadFormatters
