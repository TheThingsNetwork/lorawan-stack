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
import { injectIntl, defineMessages } from 'react-intl'
import { Col, Row } from 'react-grid-system'
import { connect } from 'react-redux'

import TYPES from '@console/constants/formatter-types'

import Select from '@ttn-lw/components/select'
import Form from '@ttn-lw/components/form'
import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Input from '@ttn-lw/components/input'
import CodeEditor from '@ttn-lw/components/code-editor'
import Link from '@ttn-lw/components/link'
import Notification from '@ttn-lw/components/notification'
import Button from '@ttn-lw/components/button'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { address as addressRegexp } from '@console/lib/regexp'

import { selectVersionIds } from '@console/store/selectors/devices'

import { getDefaultGrpcServiceFormatter, getDefaultJavascriptFormatter } from './formatter-values'
import TestForm from './test-form'

import style from './payload-formatters-form.styl'

const m = defineMessages({
  grpc: 'GRPC service',
  repository: 'Repository',
  formatterType: 'Formatter type',
  formatterParameter: 'Formatter parameter',
  grpcFieldDescription: 'The address of the service to connect to',
  appFormatter: 'Use application payload formatter  ',
  appFormatterWarning: 'This option will affect both uplink and downlink formatter',
  setupSubTitle: 'Setup',
  defaultFormatter:
    'Click <Link>here</Link> to modify the default payload formatter for this application. The payload formatter of this application is currently set to `{defaultFormatter}`',
  pasteRepositoryFormatter: 'Paste repository formatter',
  pasteApplicationFormatter: 'Paste application formatter',
})

const FIELD_NAMES = {
  SELECT: 'types-select',
  JAVASCRIPT: 'javascript-formatter',
  GRPC: 'grpc-formatter',
}

const formatterOptionsWithReset = [
  { label: m.appFormatter, value: TYPES.DEFAULT },
  { label: sharedMessages.none, value: TYPES.NONE },
  { label: 'Javascript', value: TYPES.JAVASCRIPT },
  { label: m.grpc, value: TYPES.GRPC },
  { label: 'CayenneLPP', value: TYPES.CAYENNELPP },
  { label: m.repository, value: TYPES.REPOSITORY },
]
const formatterOptions = formatterOptionsWithReset.slice(1, formatterOptionsWithReset.length)

const validationSchema = Yup.object().shape({
  [FIELD_NAMES.SELECT]: Yup.string()
    .oneOf(Object.values(TYPES))
    .required(sharedMessages.validateRequired),
  [FIELD_NAMES.JAVASCRIPT]: Yup.string().when(FIELD_NAMES.SELECT, {
    is: TYPES.JAVASCRIPT,
    then: Yup.string()
      .required(sharedMessages.validateRequired)
      // See https://github.com/TheThingsNetwork/lorawan-stack/blob/v3.14/api/messages.proto#L380
      // for validation requirements.
      .max(40960, Yup.passValues(sharedMessages.validateTooLong)),
  }),
  [FIELD_NAMES.GRPC]: Yup.string()
    .matches(addressRegexp, Yup.passValues(sharedMessages.validateAddressFormat))
    .when(FIELD_NAMES.SELECT, {
      is: TYPES.GRPC,
      then: Yup.string()
        .required(sharedMessages.validateRequired)
        .max(40960, Yup.passValues(sharedMessages.validateTooLong)),
    }),
})

@connect(state => {
  const version_ids = selectVersionIds(state)
  return {
    version_ids,
  }
})
class PayloadFormattersForm extends React.Component {
  static propTypes = {
    version_ids: PropTypes.shape({}),
  }
  static defaultProps = {
    version_ids: {},
  }

  constructor(props) {
    super(props)
    this.state = {
      type: props.initialType,
      error: undefined,
      test: {
        result: undefined,
        warning: undefined,
        error: undefined,
      },
    }

    this.formRef = React.createRef(null)
  }

  @bind
  onTypeChange(type) {
    const { onTypeChange } = this.props

    this.setState({ type }, () => onTypeChange(type))
  }

  @bind
  async handleSubmit(values, { resetForm }) {
    const { onSubmit, onSubmitSuccess, onSubmitFailure, uplink } = this.props

    this.setState({ error: undefined })

    const {
      [FIELD_NAMES.SELECT]: type,
      [FIELD_NAMES.JAVASCRIPT]: javascriptParameter,
      [FIELD_NAMES.GRPC]: grpcParameter,
    } = values

    const resetValues = {
      test: values.test,
      [FIELD_NAMES.SELECT]: type,
    }

    let parameter
    switch (type) {
      case TYPES.JAVASCRIPT:
        parameter = javascriptParameter
        resetValues[FIELD_NAMES.JAVASCRIPT] = javascriptParameter
        resetValues[FIELD_NAMES.GRPC] = getDefaultGrpcServiceFormatter(uplink)
        break
      case TYPES.GRPC:
        parameter = grpcParameter
        resetValues[FIELD_NAMES.JAVASCRIPT] = getDefaultJavascriptFormatter(uplink)
        resetValues[FIELD_NAMES.GRPC] = grpcParameter
        break
      default:
        resetValues[FIELD_NAMES.GRPC] = getDefaultGrpcServiceFormatter(uplink)
        resetValues[FIELD_NAMES.JAVASCRIPT] = getDefaultJavascriptFormatter(uplink)
        break
    }

    try {
      const result = await onSubmit({ type, parameter })
      resetForm({ values: resetValues })
      await onSubmitSuccess(result)
    } catch (error) {
      resetForm({ values: resetValues })

      this.setState({ error })
      await onSubmitFailure(error)
    }
  }

  @bind
  async handleTestSubmit(values) {
    const { onTestSubmit, defaultType, defaultParameter } = this.props
    const { values: formatterValues } = this.formRef.current

    const { payload, f_port } = values
    const {
      [FIELD_NAMES.SELECT]: selectedFormatter,
      [FIELD_NAMES.JAVASCRIPT]: javascriptParameter,
      [FIELD_NAMES.GRPC]: grpcParameter,
    } = formatterValues

    let parameter
    let formatter = selectedFormatter
    switch (selectedFormatter) {
      case TYPES.JAVASCRIPT:
        parameter = javascriptParameter
        break
      case TYPES.GRPC:
        parameter = grpcParameter
        break
      case TYPES.DEFAULT:
        parameter = defaultParameter
        formatter = defaultType
        break
    }

    try {
      const { payload: decodedPayload, warnings = [] } = await onTestSubmit({
        f_port,
        payload,
        parameter,
        formatter,
      })
      this.setState({ test: { result: decodedPayload, warning: warnings[0], error: undefined } })
    } catch (error) {
      this.setState({ test: { result: undefined, warning: undefined, error } })
    }
  }

  @bind
  pastePayloadFormatter(app) {
    const { defaultParameter, uplink } = this.props
    const repositoryFormatter = ''
    const applicationFormatter = defaultParameter
      ? defaultParameter
      : getDefaultJavascriptFormatter(uplink)
    if (app && this.formRef !== null) {
      return () =>
        this.formRef?.current?.setFieldValue(FIELD_NAMES.JAVASCRIPT, applicationFormatter)
    }

    return () => this.formRef?.current?.setFieldValue(FIELD_NAMES.JAVASCRIPT, repositoryFormatter)
  }

  get formatter() {
    const { defaultType } = this.props
    const { type } = this.state
    const showParameter =
      type === TYPES.JAVASCRIPT ||
      (type === TYPES.DEFAULT && defaultType === 'FORMATTER_JAVASCRIPT')
    const isReadOnly = type === TYPES.DEFAULT && defaultType === 'FORMATTER_JAVASCRIPT'

    if (showParameter) {
      return (
        <>
          <Form.Field
            readOnly={isReadOnly}
            required
            component={CodeEditor}
            name={FIELD_NAMES.JAVASCRIPT}
            title={m.formatterParameter}
            height="10rem"
            minLines={15}
            maxLines={15}
          />
          {type === TYPES.JAVASCRIPT && (
            <>
              <Button
                type="button"
                message={m.pasteApplicationFormatter}
                secondary
                icon={'payload_formats'}
                onClick={this.pastePayloadFormatter(true)}
              />
              <Button
                type="button"
                message={m.pasteRepositoryFormatter}
                secondary
                icon={'payload_formats'}
                onClick={this.pastePayloadFormatter(false)}
              />
            </>
          )}
        </>
      )
    } else if (type === TYPES.GRPC) {
      return (
        <Form.Field
          required
          component={Input}
          title={m.formatterParameter}
          name={FIELD_NAMES.GRPC}
          type="text"
          placeholder={sharedMessages.addressPlaceholder}
          description={m.grpcFieldDescription}
          autoComplete="on"
        />
      )
    }

    return null
  }

  @bind
  _showTestSection() {
    const { allowTest, defaultType } = this.props
    const { type } = this.state

    // Show the testing section if:
    // 1. This payload formatters form is linked to this end device.
    if (!allowTest) {
      return false
    }
    // 2. This end device is set to use the application level formatter and it is not set to `NONE`.
    if (type === TYPES.DEFAULT) {
      return defaultType !== TYPES.NONE
    }
    // 3. The end device formatter is not set to `NONE`.
    return type !== TYPES.NONE
  }

  render() {
    const {
      initialType,
      initialParameter,
      uplink,
      allowReset,
      defaultType,
      appId,
      isDefaultType,
      version_ids,
    } = this.props
    console.log(version_ids)
    const { error, type, test } = this.state

    const initialValues = {
      [FIELD_NAMES.SELECT]: type,
      [FIELD_NAMES.JAVASCRIPT]:
        initialType === TYPES.JAVASCRIPT ? initialParameter : getDefaultJavascriptFormatter(uplink),
      [FIELD_NAMES.GRPC]:
        initialType === TYPES.GRPC ? initialParameter : getDefaultGrpcServiceFormatter(uplink),
    }
    const options = allowReset ? formatterOptionsWithReset : formatterOptions
    const defaultFormatter = defaultType.replace('FORMATTER_', '').toLowerCase()

    return (
      <Row>
        <Col sm={12} lg={this._showTestSection() ? 6 : 12}>
          <Form
            submitEnabledWhenInvalid
            onSubmit={this.handleSubmit}
            initialValues={initialValues}
            validationSchema={validationSchema}
            error={error}
            formikRef={this.formRef}
          >
            <Form.SubTitle title={m.setupSubTitle} />
            <Form.Field
              disabled={type === TYPES.DEFAULT && defaultType === 'FORMATTER_REPOSITORY'}
              name={FIELD_NAMES.SELECT}
              title={m.formatterType}
              component={Select}
              options={options}
              onChange={this.onTypeChange}
              warning={
                type === TYPES.DEFAULT || type === TYPES.NONE ? m.appFormatterWarning : undefined
              }
              inputWidth="m"
              required
            />
            {isDefaultType && (
              <Notification
                small
                info
                content={m.defaultFormatter}
                convertBackticks
                messageValues={{
                  Link: msg => (
                    <Link
                      secondary
                      key="manual-link"
                      to={`/applications/${appId}/payload-formatters/uplink`}
                    >
                      {msg}
                    </Link>
                  ),
                  defaultFormatter,
                }}
              />
            )}
            {this.formatter}
            <SubmitBar>
              <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
            </SubmitBar>
          </Form>
        </Col>
        {this._showTestSection() && (
          <Col sm={12} lg={6}>
            <TestForm
              className={style.testForm}
              onSubmit={this.handleTestSubmit}
              uplink={uplink}
              payload={test.result}
              warning={test.warning}
              error={test.error}
            />
          </Col>
        )}
      </Row>
    )
  }
}

PayloadFormattersForm.propTypes = {
  allowReset: PropTypes.bool,
  allowTest: PropTypes.bool,
  appId: PropTypes.string,
  defaultParameter: PropTypes.string,
  defaultType: PropTypes.string,
  initialParameter: PropTypes.string,
  initialType: PropTypes.oneOf(Object.values(TYPES)).isRequired,
  intl: PropTypes.shape({
    formatMessage: PropTypes.func.isRequired,
  }).isRequired,
  isDefaultType: PropTypes.bool,
  onSubmit: PropTypes.func.isRequired,
  onSubmitFailure: PropTypes.func,
  onSubmitSuccess: PropTypes.func,
  onTestSubmit: PropTypes.func,
  onTypeChange: PropTypes.func,
  uplink: PropTypes.bool.isRequired,
}

PayloadFormattersForm.defaultProps = {
  initialParameter: '',
  defaultParameter: '',
  onSubmitSuccess: () => null,
  onSubmitFailure: () => null,
  allowReset: false,
  allowTest: false,
  onTestSubmit: () => null,
  defaultType: TYPES.NONE,
  onTypeChange: () => null,
  appId: undefined,
  isDefaultType: undefined,
}

export default injectIntl(PayloadFormattersForm)
