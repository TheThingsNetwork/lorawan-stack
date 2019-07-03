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
import { defineMessages } from 'react-intl'
import * as Yup from 'yup'

import Form from '../../../components/form'
import Radio from '../../../components/radio-button'
import SubmitButton from '../../../components/submit-button'
import SubmitBar from '../../../components/submit-bar'
import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'
import { address as addressRegexp } from '../../lib/regexp'
import Input from '../../../components/input'
import CodeEditor from '../../../components/code-editor'
import TYPES from '../../constants/formatter-types'
import {
  getDefaultGrpcServiceFormatter,
  getDefaultJavascriptFormatter,
} from './formatter-values'

const m = defineMessages({
  grpc: 'GRPC Service',
  repository: 'Repository',
  formatterType: 'Formatter Type',
  formatterParameter: 'Formatter Parameter',
  grpcDescription: 'The address of the service to connect to',
})

const FIELD_NAMES = {
  RADIO: 'types-radio',
  JAVASCRIPT: 'javascript-formatter',
  GRPC: 'grpc-formatter',
}

const validationSchema = Yup.object().shape({
  [FIELD_NAMES.RADIO]: Yup.string().oneOf(Object.values(TYPES)),
  [FIELD_NAMES.JAVASCRIPT]: Yup.string()
    .when('types-radio', {
      is: TYPES.JAVASCRIPT,
      then: Yup.string().required(sharedMessages.validateRequired),
    }),
  [FIELD_NAMES.GRPC]: Yup.string()
    .matches(addressRegexp, sharedMessages.validateAddressFormat)
    .when( FIELD_NAMES.RADIO, {
      is: TYPES.GRPC,
      then: Yup.string().required(sharedMessages.validateRequired),
    }),
})

@bind
class PayloadFormattersForm extends React.Component {

  constructor (props) {
    super(props)

    this.state = {
      type: props.initialType,
      error: '',
    }
  }

  onTypeChange (type) {
    this.setState({ type })
  }

  async handleSubmit (values, { resetForm }) {
    const {
      onSubmit,
      onSubmitSuccess,
      onSubmitFailure,
    } = this.props

    const {
      [FIELD_NAMES.RADIO]: type,
      [FIELD_NAMES.JAVASCRIPT]: javascriptParameter,
      [FIELD_NAMES.GRPC]: grpcParameter,
    } = values

    const resetValues = {
      [FIELD_NAMES.RADIO]: type,
    }

    let parameter = ''
    switch (type) {
    case TYPES.JAVASCRIPT:
      parameter = javascriptParameter
      resetValues[FIELD_NAMES.JAVASCRIPT] = javascriptParameter
      break
    case TYPES.GRPC:
      parameter = grpcParameter
      resetValues[FIELD_NAMES.GRPC] = grpcParameter
      break
    default:
      parameter = undefined
    }

    try {
      const result = await onSubmit({ type, parameter })
      resetForm(resetValues)
      await onSubmitSuccess(result)
    } catch (error) {
      resetForm(resetValues)

      await this.setState({ error })
      await onSubmitFailure(error)
    }
  }

  get formatter () {
    const { linked } = this.props

    if (!linked) {
      return null
    }

    const { type } = this.state

    switch (type) {
    case TYPES.JAVASCRIPT:
      return (
        <Form.Field
          required
          component={CodeEditor}
          horizontal
          name={FIELD_NAMES.JAVASCRIPT}
          title={m.formatterParameter}
          height="10rem"
          minLines={15}
          maxLines={15}
        />
      )
    case TYPES.GRPC:
      return (
        <Form.Field
          horizontal
          required
          component={Input}
          title={m.formatterParameter}
          name={FIELD_NAMES.GRPC}
          type="text"
          placeholder={sharedMessages.addressPlaceholder}
          description={m.grpcDescription}
        />
      )
    default:
      return null
    }
  }

  render () {
    const {
      initialType,
      initialParameter,
      linked,
      uplink,
    } = this.props

    const initialValues = {
      [FIELD_NAMES.RADIO]: initialType,
      [FIELD_NAMES.JAVASCRIPT]: initialType === TYPES.JAVASCRIPT
        ? initialParameter
        : getDefaultJavascriptFormatter(uplink),
      [FIELD_NAMES.GRPC]: initialType === TYPES.GRPC
        ? initialParameter
        : getDefaultGrpcServiceFormatter(uplink),
    }

    return (
      <div>
        <Form
          disabled={!linked}
          submitEnabledWhenInvalid
          horizontal
          onSubmit={this.handleSubmit}
          initialValues={initialValues}
          validationSchema={validationSchema}
        >
          <Form.Field
            name={FIELD_NAMES.RADIO}
            title={m.formatterType}
            component={Radio.Group}
            onChange={this.onTypeChange}
          >
            <Radio
              label={sharedMessages.none}
              value={TYPES.NONE}
            />
            <Radio
              label="Javascript"
              value={TYPES.JAVASCRIPT}
            />
            <Radio
              label={m.grpc}
              value={TYPES.GRPC}
            />
            <Radio
              label="CayenneLPP"
              value={TYPES.CAYENNELPP}
            />
            <Radio
              label={m.repository}
              value={TYPES.REPOSITORY}
            />
          </Form.Field>
          {this.formatter}
          <SubmitBar>
            <Form.Submit
              component={SubmitButton}
              message={sharedMessages.saveChanges}
            />
          </SubmitBar>
        </Form>
      </div>
    )
  }
}

PayloadFormattersForm.propTypes = {
  initialType: PropTypes.oneOf(Object.values(TYPES)).isRequired,
  initialParameter: PropTypes.string,
  onSubmit: PropTypes.func.isRequired,
  onSubmitSuccess: PropTypes.func,
  onSubmitFailure: PropTypes.func,
  linked: PropTypes.bool.isRequired,
  uplink: PropTypes.bool.isRequired,
}

PayloadFormattersForm.defaultProps = {
  initialParameter: '',
  onSubmitSuccess: () => null,
  onSubmitFailure: () => null,
}

export default PayloadFormattersForm
