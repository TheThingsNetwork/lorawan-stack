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
import ButtonGroup from '@ttn-lw/components/button/group'

import Message from '@ttn-lw/lib/components/message'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import FORMATTER_NAMES from '@ttn-lw//lib/payload-formatter-messages'

import { address as addressRegexp } from '@console/lib/regexp'

import { getDefaultGrpcServiceFormatter, getDefaultJavascriptFormatter } from './formatter-values'
import TestForm from './test-form'

import style from './payload-formatters-form.styl'

const m = defineMessages({
  grpc: 'GRPC service',
  repository: 'Use Device Repository formatters',
  customJavascipt: 'Custom Javascript formatter',
  formatterType: 'Formatter type',
  formatterCode: 'Formatter code',
  formatterCodeReadOnly: 'Formatter code (read only)',
  grpcHost: 'GRPC host',
  grpcFieldDescription: 'The address of the service to connect to',
  appFormatter: 'Use application payload formatter',
  appFormatterWarning: 'This option will affect both uplink and downlink formatter',
  setupSubTitle: 'Setup',
  defaultFormatter:
    'Click <Link>here</Link> to modify the default payload formatter for this application. The payload formatter of this application is currently set to `{defaultFormatter}`',
  pasteRepositoryFormatter: 'Paste repository formatter',
  pasteApplicationFormatter: 'Paste application formatter',
  learnMoreAboutDeviceRepo: 'What is the Device Repository formatter option?',
  learnMoreAboutPayloadFormatters: 'Learn more about payload formatters',
  learnMoreAboutCayenne: 'What is CayenneLPP?',
  noRepositoryWarning:
    'The application formatter is set to `Repository` but this device does not have an associated formatter in the LoRaWAN Device repository. Messages for this end device will hence not be formatted.',
  confirmNavigationTitle: 'Confirm navigation',
  confirmNavigationMessage:
    'Are you sure you want to leave this page? Your current changes have not been saved yet.',
})

const FIELD_NAMES = {
  SELECT: 'types-select',
  JAVASCRIPT: 'javascript-formatter',
  GRPC: 'grpc-formatter',
  REPOSITORY: 'repository-formatter',
}

const formatterOptions = [
  { label: m.appFormatter, value: TYPES.DEFAULT },
  { label: m.repository, value: TYPES.REPOSITORY },
  { label: m.customJavascipt, value: TYPES.JAVASCRIPT },
  { label: m.grpc, value: TYPES.GRPC },
  { label: 'CayenneLPP', value: TYPES.CAYENNELPP },
  { label: sharedMessages.none, value: TYPES.NONE },
]

const validationSchema = Yup.object().shape({
  [FIELD_NAMES.SELECT]: Yup.string()
    .oneOf(Object.values(TYPES))
    .required(sharedMessages.validateRequired),
  [FIELD_NAMES.JAVASCRIPT]: Yup.string().when(FIELD_NAMES.SELECT, {
    is: TYPES.JAVASCRIPT,
    then: schema =>
      schema
        .required(sharedMessages.validateRequired)
        // See https://github.com/TheThingsNetwork/lorawan-stack/blob/v3.14/api/messages.proto#L380
        // for validation requirements.
        .max(40960, Yup.passValues(sharedMessages.validateTooLong)),
  }),
  [FIELD_NAMES.GRPC]: Yup.string()
    .matches(addressRegexp, Yup.passValues(sharedMessages.validateAddressFormat))
    .when(FIELD_NAMES.SELECT, {
      is: TYPES.GRPC,
      then: schema =>
        schema
          .required(sharedMessages.validateRequired)
          .max(40960, Yup.passValues(sharedMessages.validateTooLong)),
    }),
})

class PayloadFormattersForm extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      type: props.initialType,
      isSubmitting: false,
      error: undefined,
      testResult: {},
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

    this.setState({ error: undefined, isSubmitting: true })

    const {
      [FIELD_NAMES.SELECT]: type,
      [FIELD_NAMES.JAVASCRIPT]: javascriptParameter,
      [FIELD_NAMES.GRPC]: grpcParameter,
    } = values

    const resetValues = {
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
      this.setState({ isSubmitting: false })
      await onSubmitSuccess(result)
    } catch (error) {
      resetForm({ values: resetValues })

      this.setState({ error, isSubmitting: false })
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
      const testResult = await onTestSubmit({
        f_port,
        payload,
        parameter,
        formatter,
      })
      this.setState({ testResult })
    } catch (error) {
      this.setState({ testResult: error })
    }
  }

  @bind
  pasteAppPayloadFormatter() {
    const { defaultParameter, uplink } = this.props
    this.formRef?.current?.setFieldValue(
      FIELD_NAMES.JAVASCRIPT,
      defaultParameter || getDefaultJavascriptFormatter(uplink),
    )
  }

  @bind
  pasteRepoPayloadFormatters() {
    const { repoFormatters } = this.props
    this.formRef?.current?.setFieldValue(
      FIELD_NAMES.JAVASCRIPT,
      repoFormatters?.formatter_parameter,
    )
  }

  get formatter() {
    const { defaultType, repoFormatters } = this.props
    const { type } = this.state
    const hasRepoFormatter =
      repoFormatters !== undefined && Object.keys(repoFormatters).length !== 0
    const showParameter =
      type === TYPES.JAVASCRIPT ||
      (type === TYPES.DEFAULT && defaultType === 'FORMATTER_JAVASCRIPT')
    const showRepositoryParameter =
      (type === TYPES.REPOSITORY && hasRepoFormatter) ||
      (type === TYPES.DEFAULT && defaultType === 'FORMATTER_REPOSITORY')

    if (showParameter) {
      return (
        <>
          <Form.Field
            required
            component={CodeEditor}
            name={FIELD_NAMES.JAVASCRIPT}
            title={m.formatterCode}
            height="10rem"
            minLines={25}
            maxLines={25}
          />
          {type === TYPES.JAVASCRIPT && (
            <ButtonGroup>
              {defaultType !== 'FORMATTER_NONE' && (
                <Button
                  type="button"
                  message={m.pasteApplicationFormatter}
                  secondary
                  onClick={this.pasteAppPayloadFormatter}
                />
              )}
              {hasRepoFormatter && (
                <Button
                  type="button"
                  message={m.pasteRepositoryFormatter}
                  secondary
                  onClick={this.pasteRepoPayloadFormatters}
                />
              )}
            </ButtonGroup>
          )}
        </>
      )
    } else if (type === TYPES.GRPC) {
      return (
        <Form.Field
          required
          component={Input}
          title={m.grpcHost}
          name={FIELD_NAMES.GRPC}
          type="text"
          placeholder={sharedMessages.addressPlaceholder}
          description={m.grpcFieldDescription}
          autoComplete="on"
        />
      )
    } else if (type === TYPES.CAYENNELPP) {
      return (
        <Link.DocLink path="/integrations/payload-formatters/device-repo/cayenne" secondary>
          <Message content={m.learnMoreAboutCayenne} />
        </Link.DocLink>
      )
    } else if (showRepositoryParameter) {
      if (!hasRepoFormatter) {
        return <Notification warning content={m.noRepositoryWarning} small />
      }
      return (
        <>
          <Form.Field
            readOnly
            component={CodeEditor}
            title={m.formatterCodeReadOnly}
            name={FIELD_NAMES.REPOSITORY}
            type="text"
            height="10rem"
            minLines={25}
            maxLines={25}
            value={repoFormatters?.formatter_parameter}
          />
          <Link.DocLink path="/integrations/payload-formatters/device-repo/" secondary>
            <Message content={m.learnMoreAboutDeviceRepo} />
          </Link.DocLink>
        </>
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
      repoFormatters,
    } = this.props

    const { error, type, testResult, isSubmitting } = this.state

    const initialValues = {
      [FIELD_NAMES.SELECT]: type,
      [FIELD_NAMES.JAVASCRIPT]:
        initialType === TYPES.JAVASCRIPT ? initialParameter : getDefaultJavascriptFormatter(uplink),
      [FIELD_NAMES.GRPC]:
        initialType === TYPES.GRPC ? initialParameter : getDefaultGrpcServiceFormatter(uplink),
    }
    const hasRepoFormatter =
      repoFormatters !== undefined && Object.keys(repoFormatters).length !== 0
    let options = allowReset
      ? formatterOptions
      : formatterOptions.filter(o => o.value !== TYPES.DEFAULT)
    if (!hasRepoFormatter && allowReset) {
      options = options.filter(o => o.value !== TYPES.REPOSITORY)
    }
    const defaultFormatter = FORMATTER_NAMES[defaultType].defaultMessage

    return (
      <>
        <Row>
          <Col sm={12} lg={this._showTestSection() ? 6 : 12}>
            <Form
              onSubmit={this.handleSubmit}
              initialValues={initialValues}
              validationSchema={validationSchema}
              error={error}
              formikRef={this.formRef}
              id="payload-formatter-form"
            >
              {() => (
                <>
                  <Form.SubTitle title={m.setupSubTitle} />
                  <Form.Field
                    name={FIELD_NAMES.SELECT}
                    title={m.formatterType}
                    component={Select}
                    options={options}
                    onChange={this.onTypeChange}
                    warning={type === TYPES.DEFAULT ? m.appFormatterWarning : undefined}
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
                  {/*
                      // TODO: Refactor to use data API and re-enable prompt.
                      // NOTE: Unfortunately react router v6 requires us to do further
                      // refactoring to use the data API to be able to use `usePrompt`
                      // again, which is required to make the Prompt component work.
                      // For now we will disable the prompt.
                  <Prompt
                    when={Boolean(touched['javascript-formatter'] || touched['grpc-formatter'])}
                    modal={{
                      title: m.confirmNavigationTitle,
                      message: m.confirmNavigationMessage,
                      buttonMessage: m.confirmNavigationTitle,
                    }}
                  />
                  */}
                </>
              )}
            </Form>
          </Col>
          {this._showTestSection() && (
            <Col sm={12} lg={6}>
              <TestForm
                className={style.testForm}
                onSubmit={this.handleTestSubmit}
                uplink={uplink}
                testResult={testResult}
              />
              <Link.DocLink path="/integrations/payload-formatters" secondary>
                <Message content={m.learnMoreAboutPayloadFormatters} />
              </Link.DocLink>
            </Col>
          )}
        </Row>
        <Row>
          <Col sm={12}>
            <SubmitBar>
              <SubmitButton
                message={sharedMessages.saveChanges}
                form="payload-formatter-form"
                isSubmitting={isSubmitting}
                isValidating={false}
              />
            </SubmitBar>
          </Col>
        </Row>
      </>
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
  repoFormatters: PropTypes.shape({
    formatter_parameter: PropTypes.string,
  }),
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
  repoFormatters: undefined,
}

export default injectIntl(PayloadFormattersForm)
