// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import React, { useState, useCallback } from 'react'
import { defineMessages } from 'react-intl'

import api from '@console/api'

import SubmitButton from '@ttn-lw/components/submit-button'
import RadioButton from '@ttn-lw/components/radio-button'
import Checkbox from '@ttn-lw/components/checkbox'
import Input from '@ttn-lw/components/input'
import SubmitBar from '@ttn-lw/components/submit-bar'
import toast from '@ttn-lw/components/toast'
import Form from '@ttn-lw/components/form'

import Message from '@ttn-lw/lib/components/message'

import Yup from '@ttn-lw/lib/yup'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { hexToBase64 } from '@console/lib/bytes'

const m = defineMessages({
  insertMode: 'Insert Mode',
  insertModeDescription: 'Messages can either replace the downlink queue or append to it',
  lengthMustBeEven: 'Payload must be complete',
  replace: 'Replace',
  push: 'Push',
  validateFPort: 'FPort must be between 1 and 223',
  confirmation: 'Confirmation',
  confirmed: 'Confirmed',
  scheduleDownlink: 'Schedule Downlink',
  downlinkSuccess: 'Downlink was scheduled successfully',
  payloadDescription: 'The desired payload bytes of the downlink message',
})

const validationSchema = Yup.object({
  _mode: Yup.string()
    .oneOf(['replace', 'push'])
    .required(sharedMessages.validateRequired),
  f_port: Yup.number()
    .min(1, m.validateFPort)
    .max(223, m.validateFPort)
    .required(sharedMessages.validateRequired),
  confirmed: Yup.bool().required(),
  frm_payload: Yup.string().test(
    'len',
    m.lengthMustBeEven,
    val => !Boolean(val) || val.length % 2 === 0,
  ),
})

const DownlinkForm = ({ appId, devId }) => {
  const [error, setError] = useState('')
  const handleSubmit = useCallback(
    async (vals, { setSubmitting, resetForm }) => {
      const { _mode, ...values } = validationSchema.cast(vals)
      try {
        values.frm_payload = hexToBase64(values.frm_payload)
        await api.downlinkQueue[_mode](appId, devId, [values])
        toast({
          title: sharedMessages.success,
          type: toast.types.SUCCESS,
          message: m.downlinkSuccess,
        })
        setSubmitting(false)
      } catch (err) {
        setError(err)
        resetForm({ values: vals })
      }
    },
    [appId, devId],
  )

  const initialValues = {
    _mode: 'replace',
    f_port: 1,
    confirmed: false,
    frm_payload: '',
  }

  return (
    <Form
      error={error}
      onSubmit={handleSubmit}
      initialValues={initialValues}
      validationSchema={validationSchema}
    >
      <Message component="h4" content={m.scheduleDownlink} />
      <Form.Field
        name="_mode"
        title={m.insertMode}
        component={RadioButton.Group}
        description={m.insertModeDescription}
      >
        <RadioButton label={m.replace} value="replace" />
        <RadioButton label={m.push} value="push" />
      </Form.Field>
      <Form.Field
        title={m.confirmation}
        label={m.confirmed}
        name="confirmed"
        component={Checkbox}
      />
      <Form.Field
        title="FPort"
        name="f_port"
        component={Input}
        type="number"
        min={1}
        max={223}
        required
      />
      <Form.Field
        title={sharedMessages.payload}
        description={m.payloadDescription}
        name="frm_payload"
        component={Input}
        type="byte"
        required
        unbounded
      />
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={m.scheduleDownlink} />
      </SubmitBar>
    </Form>
  )
}

DownlinkForm.propTypes = {
  appId: PropTypes.string.isRequired,
  devId: PropTypes.string.isRequired,
}

export default DownlinkForm
