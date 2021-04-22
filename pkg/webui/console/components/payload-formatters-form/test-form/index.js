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
import { defineMessages } from 'react-intl'

import Form from '@ttn-lw/components/form'
import SubmitButton from '@ttn-lw/components/submit-button'
import CodeEditor from '@ttn-lw/components/code-editor'
import Icon from '@ttn-lw/components/icon'
import Input from '@ttn-lw/components/input'
import SubmitBar from '@ttn-lw/components/submit-bar'

import Message from '@ttn-lw/lib/components/message'
import ErrorMessage from '@ttn-lw/lib/components/error-message'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './test-form.styl'

const m = defineMessages({
  jsonError: '{field} must be a valid JSON object',
  validResult: 'Payload is valid',
  noResult: 'No test result generated yet',
  testDecoder: 'Test decoder',
  testEncoder: 'Test encoder',
  testSubTitle: 'Test',
  bytePayload: 'Byte payload',
  jsonPayload: 'JSON payload',
})

const validationSchema = Yup.object({
  f_port: Yup.number()
    .min(1, Yup.passValues(sharedMessages.validateNumberGte))
    .max(223, Yup.passValues(sharedMessages.validateNumberLte)),
  payload: Yup.string().when(['$isUplink'], (isUplink, schema) => {
    if (isUplink) {
      return schema.test(
        'len',
        Yup.passValues(sharedMessages.validateHexLength),
        payload => !Boolean(payload) || payload.length % 2 === 0,
      )
    }

    return schema.test('valid-json', Yup.passValues(m.jsonError), json => {
      try {
        JSON.parse(json)
        return true
      } catch (e) {
        return false
      }
    })
  }),
})

const TestForm = props => {
  const { className, onSubmit, uplink, payload, warning, error } = props

  const initialValues = React.useMemo(() => ({ payload: '', f_port: 1 }), [])
  const validationContext = React.useMemo(
    () => ({
      isUplink: uplink,
    }),
    [uplink],
  )

  const hasTestError = Boolean(error)
  const hasTestWarning = Boolean(warning)
  const hasPayload = Boolean(payload)

  const showTestError = hasTestError
  const showTestWarning = !hasTestError && hasTestWarning
  const showTestValid = !hasTestError && !hasTestWarning && hasPayload

  let infoIcon = 'info'
  let infoMessage = m.noResult
  if (showTestError) {
    infoIcon = 'error'
    infoMessage = error
  } else if (showTestWarning) {
    infoIcon = 'warning'
    infoMessage = warning
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
        <Form.FieldContainer horizontal>
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
              title={m.jsonPayload}
              language="json"
              name="payload"
              component={CodeEditor}
              minLines={14}
              maxLines={14}
            />
          )}
          <Form.Field
            className={style.fPort}
            inputWidth="xxs"
            title="FPort"
            name="f_port"
            type="number"
            component={Input}
            min={1}
            max={223}
            autoWidth
          />
        </Form.FieldContainer>
        <hr className={style.hRule} />
        <div
          className={classnames(style.infoSection, {
            [style.infoSectionError]: showTestError,
            [style.infoSectionWarning]: showTestWarning,
            [style.infoSectionValid]: showTestValid,
          })}
        >
          <Icon className={style.icon} icon={infoIcon} nudgeUp />
          {showTestError ? (
            <ErrorMessage className={style.message} content={infoMessage} />
          ) : (
            <Message className={style.message} content={infoMessage} />
          )}
        </div>
        {uplink ? (
          <CodeEditor
            value={hasPayload ? JSON.stringify(payload, null, 2) : undefined}
            language="json"
            name="test_result"
            minLines={14}
            maxLines={14}
            readOnly
            showGutter={false}
          />
        ) : (
          <Input value={hasPayload ? payload : undefined} type="byte" unbounded readOnly />
        )}
        <SubmitBar>
          <Form.Submit
            component={SubmitButton}
            message={uplink ? m.testDecoder : m.testEncoder}
            secondary
          />
        </SubmitBar>
      </Form>
    </div>
  )
}

TestForm.propTypes = {
  className: PropTypes.string,
  error: PropTypes.error,
  onSubmit: PropTypes.func.isRequired,
  payload: PropTypes.oneOfType([PropTypes.string, PropTypes.shape({})]),
  uplink: PropTypes.bool.isRequired,
  warning: PropTypes.message,
}
TestForm.defaultProps = {
  payload: undefined,
  error: undefined,
  warning: undefined,
  className: undefined,
}

export default TestForm
