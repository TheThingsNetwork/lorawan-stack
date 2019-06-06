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
import { connect as formConnect, getIn } from 'formik'
import bind from 'autobind-decorator'

import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'
import Field from '../../../components/field'

import {
  formatsSelector,
  errorSelector,
  fetchingSelector,
} from '../../store/selectors/webhook-formats'

import { getWebhookFormats } from '../../store/actions/webhook-formats'

const m = defineMessages({
  formatFetchingFailure: 'Could not retrieve the list of available webhook formats',
})

const formatOptions = formats => Object.keys(formats).map(key => ({ value: key, label: formats[key]}))

@formConnect
@storeConnect(function (state, props) {
  return {
    formats: formatsSelector(state, props),
    error: errorSelector(state, props),
    fetching: fetchingSelector(state, props),
  }
},
{ getWebhookFormats }
)
@bind
class WebhookFormatsSelector extends React.PureComponent {

  constructor (props) {
    super(props)

    const { name, formik } = props

    formik.registerField(name, this)
  }

  componentDidMount () {
    const { getWebhookFormats } = this.props
    getWebhookFormats()
  }

  componentWillUnmount () {
    const { formik, name } = this.props

    formik.unregisterField(name)
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

    const {
      setFieldValue,
      setFieldTouched,
    } = this.props.formik

    const fieldOptions = this.getOptions()
    const fieldError = getIn(this.props.formik.errors, name)
    const fieldTouched = getIn(this.props.formik.touched, name)
    const fieldValue = getIn(this.props.formik.values, name)

    return (
      <Field
        horizontal={horizontal}
        type="select"
        options={fieldOptions}
        name={name}
        value={fieldValue}
        required={required}
        title={title}
        autoFocus={autoFocus}
        isLoading={fetching}
        warning={Boolean(error) ? m.formatFetchingFailure : undefined}
        error={fieldError}
        touched={fieldTouched}
        menuPlacement={menuPlacement}
        setFieldTouched={setFieldTouched}
        setFieldValue={setFieldValue}
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
