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
import classnames from 'classnames'
import { defineMessages, useIntl } from 'react-intl'

import Form from '@ttn-lw/components/form'
import SubmitButton from '@ttn-lw/components/submit-button'
import CodeEditor from '@ttn-lw/components/code-editor'
import Icon from '@ttn-lw/components/icon'
import Input from '@ttn-lw/components/input'
import SafeInspector from '@ttn-lw/components/safe-inspector'

import Message from '@ttn-lw/lib/components/message'
import ErrorMessage from '@ttn-lw/lib/components/error-message'

import { isBackend, getBackendErrorName, toMessageProps } from '@ttn-lw/lib/errors/utils'
import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { base64ToHex } from '@console/lib/bytes'

import style from './test-form.styl'

const m = defineMessages({
  validResult: 'Payload is valid',
  noResult: 'No test result generated yet',
  testDecoder: 'Test decoder',
  testEncoder: 'Test encoder',
  testSubTitle: 'Test',
  testFatalError: 'An error occurred while interpreting the formatter code',
  testError: 'Test result contains user-defined error(s)',
  testWarning: 'Test result contains user-defined warning(s)',
  bytePayload: 'Byte payload',
  jsonPayload: 'JSON payload',
  emptyPayload: 'The returned payload was empty',
  decodedPayload: 'Decoded test payload',
  completeUplink: 'Complete uplink data',
  testResult: 'Test result',
  errorInformation: 'Error information',
})

const validationSchema = Yup.object({
  f_port: Yup.number()
    .min(1, Yup.passValues(sharedMessages.validateNumberGte))
    .max(223, Yup.passValues(sharedMessages.validateNumberLte)),
  payload: Yup.string().when(['$isUplink'], ([isUplink], schema) => {
    if (isUplink) {
      return schema.test(
        'len',
        Yup.passValues(sharedMessages.validateHexLength),
        payload => !Boolean(payload) || payload.length % 3 === 0,
      )
    }

    return schema.test('valid-json', Yup.passValues(sharedMessages.validateJson), json => {
      try {
        JSON.parse(json)
        return true
      } catch (e) {
        return false
      }
    })
  }),
})

const isOutputError = error => isBackend(error) && getBackendErrorName(error) === 'output_errors'

const TestForm = props => {
  const {
    className,
    onSubmit,
    uplink,
    testResult,
    testResult: {
      decoded_payload: decodedPayload,
      decoded_payload_warnings: decodedPayloadWarnings,
      normalized_payload_warnings: normalizedPayloadWarnings,
      frm_payload: framePayload,
    },
  } = props

  const { formatMessage } = useIntl()

  const initialValues = React.useMemo(() => ({ payload: '', f_port: 1 }), [])
  const validationContext = React.useMemo(
    () => ({
      isUplink: uplink,
    }),
    [uplink],
  )

  const hasTestError = isBackend(testResult)
  const hasFatalError = !isOutputError(testResult)
  const hasTestWarning =
    (decodedPayloadWarnings instanceof Array && decodedPayloadWarnings.length !== 0) ||
    (normalizedPayloadWarnings instanceof Array && normalizedPayloadWarnings.length !== 0)
  const hasPayload = decodedPayload !== undefined

  const showTestError = hasTestError
  const showTestWarning = !hasTestError && hasTestWarning
  const showTestValid = !hasTestError && !hasTestWarning && hasPayload
  let testOutput
  if (uplink) {
    if (showTestError) {
      const errorMessage = toMessageProps(testResult)
      testOutput = formatMessage(errorMessage.content, errorMessage.values)
    } else {
      testOutput = showTestWarning || hasPayload ? JSON.stringify(testResult, null, 2) : ''
    }
  } else {
    testOutput =
      showTestError || showTestWarning
        ? JSON.stringify(testResult, null, 2)
        : 'decoded_payload' in testResult && !('frm_payload' in testResult)
        ? formatMessage(m.emptyPayload)
        : ''
  }
  let infoIcon = 'info'
  let infoMessage = m.noResult
  if (showTestError) {
    infoIcon = 'error'
    infoMessage = hasFatalError ? m.testFatalError : m.testError
  } else if (showTestWarning) {
    infoIcon = 'warning'
    infoMessage = m.testWarning
  } else if (showTestValid) {
    infoIcon = 'valid'
    infoMessage = m.validResult
  }

  return (
    <div className={className}>
      <Form
        onSubmit={onSubmit}
        initialValues={initialValues}
        validationSchema={validationSchema}
        validationContext={validationContext}
      >
        <Form.SubTitle title={m.testSubTitle} />
        <Form.FieldContainer horizontal className={style.topRow}>
          {uplink ? (
            <Form.Field
              title={m.bytePayload}
              name="payload"
              type="byte"
              component={Input}
              className={style.payload}
              unbounded
            />
          ) : (
            <Form.Field
              className={style.payload}
              title={m.jsonPayload}
              language="json"
              name="payload"
              component={CodeEditor}
              minLines={15}
              maxLines={15}
            />
          )}
          <Form.Field
            className={style.fPort}
            title="FPort"
            name="f_port"
            type="number"
            component={Input}
            min={1}
            max={223}
          />
          {uplink && (
            <Form.Submit
              className={style.submitButton}
              component={SubmitButton}
              message={uplink ? m.testDecoder : m.testEncoder}
              primary={false}
            />
          )}
        </Form.FieldContainer>
        {!uplink && (
          <Form.Submit
            className="mb-cs-m"
            component={SubmitButton}
            message={uplink ? m.testDecoder : m.testEncoder}
            primary={false}
          />
        )}
        {uplink ? (
          <>
            {!showTestError && (
              <Form.InfoField title={m.decodedPayload}>
                <CodeEditor
                  value={JSON.stringify(decodedPayload, null, 2)}
                  language="json"
                  name="test_result"
                  minLines={12}
                  maxLines={12}
                  readOnly
                  showGutter={false}
                />
              </Form.InfoField>
            )}
            <Form.InfoField title={showTestError ? m.errorInformation : m.completeUplink}>
              <CodeEditor
                value={testOutput}
                language="json"
                name="test_result"
                minLines={11}
                maxLines={11}
                readOnly
                showGutter={false}
              />
            </Form.InfoField>
          </>
        ) : (
          <Form.InfoField title={showTestError ? m.errorInformation : m.testResult}>
            {!showTestError && (
              <SafeInspector
                data={framePayload !== undefined ? base64ToHex(framePayload) : ''}
                initiallyVisible
                className="mb-cs-m"
                hideable={false}
              />
            )}
            <CodeEditor
              value={testOutput}
              language="json"
              name="test_result"
              minLines={showTestError ? 9 : 6}
              maxLines={showTestError ? 9 : 6}
              readOnly
              showGutter={false}
            />
          </Form.InfoField>
        )}
        <div
          className={classnames(style.infoSection, {
            [style.infoSectionError]: showTestError,
            [style.infoSectionWarning]: showTestWarning,
            [style.infoSectionValid]: showTestValid,
          })}
        >
          {showTestError && (
            <>
              <Icon className={style.icon} icon={infoIcon} nudgeUp />
              <ErrorMessage className={style.message} content={infoMessage} />
            </>
          )}
          {(showTestValid || showTestWarning) && (
            <>
              <Icon className={style.icon} icon={infoIcon} nudgeUp />
              <Message className={style.message} content={infoMessage} />
            </>
          )}
        </div>
      </Form>
    </div>
  )
}

TestForm.propTypes = {
  className: PropTypes.string,
  onSubmit: PropTypes.func.isRequired,
  testResult: PropTypes.shape({
    decoded_payload: PropTypes.PropTypes.shape({}),
    decoded_payload_warnings: PropTypes.arrayOf(PropTypes.message),
    normalized_payload: PropTypes.arrayOf(PropTypes.PropTypes.shape({})),
    normalized_payload_warnings: PropTypes.arrayOf(PropTypes.message),
    frm_payload: PropTypes.string,
  }).isRequired,
  uplink: PropTypes.bool.isRequired,
}
TestForm.defaultProps = {
  className: undefined,
}

export default TestForm
