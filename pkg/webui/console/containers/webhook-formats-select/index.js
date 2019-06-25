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
import { connect as storeConnect } from 'react-redux'
import { defineMessages } from 'react-intl'
import bind from 'autobind-decorator'

import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'
import Field from '../../../components/form/field'
import Select from '../../../components/select'

import {
  selectWebhookFormats,
  selectWebhookFormatsError,
  selectWebhookFormatsFetching,
} from '../../store/selectors/webhook-formats'

import { getWebhookFormats } from '../../store/actions/webhook-formats'

const m = defineMessages({
  formatFetchingFailure: 'Could not retrieve the list of available webhook formats',
})

const formatOptions = formats => Object.keys(formats).map(key => ({ value: key, label: formats[key]}))

@storeConnect(function (state) {
  return {
    formats: selectWebhookFormats(state),
    error: selectWebhookFormatsError(state),
    fetching: selectWebhookFormatsFetching(state),
  }
},
{ getWebhookFormats }
)
@bind
class WebhookFormatsSelector extends React.PureComponent {

  componentDidMount () {
    const { getWebhookFormats } = this.props
    getWebhookFormats()
  }

  getOptions () {
    const { formats } = this.props

    return formatOptions(formats)
  }

  render () {
    const {
      name,
      required,
      title,
      autoFocus,
      horizontal,
      error,
      fetching,
      menuPlacement,
    } = this.props

    const fieldOptions = this.getOptions()

    return (
      <Field
        component={Select}
        horizontal={horizontal}
        type="select"
        options={fieldOptions}
        name={name}
        required={required}
        title={title}
        autoFocus={autoFocus}
        isLoading={fetching}
        warning={Boolean(error) ? m.formatFetchingFailure : undefined}
        menuPlacement={menuPlacement}
      />
    )
  }
}

WebhookFormatsSelector.propTypes = {
  name: PropTypes.string.isRequired,
  required: PropTypes.bool.isRequired,
  title: PropTypes.message,
  autoFocus: PropTypes.bool,
  horizontal: PropTypes.bool,
  menuPlacement: PropTypes.oneOf([ 'top', 'bottom', 'auto' ]),
}

WebhookFormatsSelector.defaultProps = {
  title: sharedMessages.webhookFormat,
  autoFocus: false,
  horizontal: false,
  menuPlacement: 'auto',
}

export default WebhookFormatsSelector
