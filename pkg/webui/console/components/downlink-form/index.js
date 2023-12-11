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
import { useSelector } from 'react-redux'

import tts from '@console/api/tts'

import Notification from '@ttn-lw/components/notification'
import SubmitButton from '@ttn-lw/components/submit-button'
import RadioButton from '@ttn-lw/components/radio-button'
import Checkbox from '@ttn-lw/components/checkbox'
import Input from '@ttn-lw/components/input'
import SubmitBar from '@ttn-lw/components/submit-bar'
import toast from '@ttn-lw/components/toast'
import Form from '@ttn-lw/components/form'
import CodeEditor from '@ttn-lw/components/code-editor'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { hexToBase64 } from '@console/lib/bytes'

import {
  selectApplicationLinkSkipPayloadCrypto,
  selectSelectedApplicationId,
} from '@console/store/selectors/applications'
import { selectSelectedDevice, selectSelectedDeviceId } from '@console/store/selectors/devices'

const m = defineMessages({
  insertMode: 'Insert Mode',
  payloadType: 'Payload type',
  bytes: 'Bytes',
  replace: 'Replace downlink queue',
  push: 'Push to downlink queue (append)',
  scheduleDownlink: 'Schedule downlink',
  downlinkSuccess: 'Downlink scheduled',
  bytesPayloadDescription: 'The desired payload bytes of the downlink message',
  jsonPayloadDescription: 'The decoded payload of the downlink message',
  invalidSessionWarning:
    'Downlinks can only be scheduled for end devices with a valid session. Please make sure your end device is properly connected to the network.',
})

const validationSchema = Yup.object({
  _mode: Yup.string().oneOf(['replace', 'push']).required(sharedMessages.validateRequired),
  _payload_type: Yup.string().oneOf(['bytes', 'json']),
  f_port: Yup.number()
    .min(1, Yup.passValues(sharedMessages.validateNumberGte))
    .max(223, Yup.passValues(sharedMessages.validateNumberLte))
    .required(sharedMessages.validateRequired),
  confirmed: Yup.bool().required(),
  frm_payload: Yup.string().when('_payload_type', {
    is: type => type === 'bytes',
    then: schema =>
      schema.test(
        'len',
        Yup.passValues(sharedMessages.validateHexLength),
        val => !Boolean(val) || val.length % 3 === 0,
      ),
    otherwise: schema => schema.strip(),
  }),
  decoded_payload: Yup.string().when('_payload_type', {
    is: type => type === 'json',
    then: schema =>
      schema.test('valid-json', sharedMessages.validateJson, json => {
        try {
          JSON.parse(json)
          return true
        } catch (e) {
          return false
        }
      }),
    otherwise: schema => schema.strip(),
  }),
})

const initialValues = {
  _mode: 'replace',
  _payload_type: 'bytes',
  f_port: 1,
  confirmed: false,
  frm_payload: '',
  decoded_payload: '',
}

const DownlinkForm = () => {
  const [payloadType, setPayloadType] = React.useState('bytes')
  const [error, setError] = useState('')
  const appId = useSelector(selectSelectedApplicationId)
  const devId = useSelector(selectSelectedDeviceId)
  const device = useSelector(selectSelectedDevice)
  const skipPayloadCrypto = useSelector(selectApplicationLinkSkipPayloadCrypto)

  const handleSubmit = useCallback(
    async (vals, { setSubmitting, resetForm }) => {
      const { _mode, _payload_type, ...values } = validationSchema.cast(vals)
      try {
        if (_payload_type === 'bytes') {
          values.frm_payload = hexToBase64(values.frm_payload)
        }

        if (_payload_type === 'json') {
          values.decoded_payload = JSON.parse(values.decoded_payload)
        }

        await tts.Applications.Devices.DownlinkQueue[_mode](appId, devId, [values])
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

  const validSession = device.session || device.pending_session
  const payloadCryptoSkipped = device.skip_payload_crypto_override ?? skipPayloadCrypto
  const deviceSimulationDisabled = !validSession || payloadCryptoSkipped

  return (
    <>
      {payloadCryptoSkipped && (
        <Notification content={sharedMessages.deviceSimulationDisabledWarning} warning small />
      )}
      {!validSession && <Notification content={m.invalidSessionWarning} warning small />}
      <IntlHelmet title={m.scheduleDownlink} />
      <Form
        error={error}
        onSubmit={handleSubmit}
        initialValues={initialValues}
        validationSchema={validationSchema}
        disabled={deviceSimulationDisabled}
      >
        <Form.SubTitle title={m.scheduleDownlink} />
        <Form.Field name="_mode" title={m.insertMode} component={RadioButton.Group}>
          <RadioButton label={m.replace} value="replace" />
          <RadioButton label={m.push} value="push" />
        </Form.Field>
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
          name="_payload_type"
          title={m.payloadType}
          component={RadioButton.Group}
          onChange={setPayloadType}
          horizontal
        >
          <RadioButton label={m.bytes} value="bytes" />
          <RadioButton label="JSON" value="json" />
        </Form.Field>
        {payloadType === 'bytes' ? (
          <Form.Field
            title={sharedMessages.payload}
            description={m.bytesPayloadDescription}
            name="frm_payload"
            component={Input}
            type="byte"
            unbounded
          />
        ) : (
          <Form.Field
            title={sharedMessages.payload}
            description={m.jsonPayloadDescription}
            language="json"
            name="decoded_payload"
            component={CodeEditor}
            minLines={14}
            maxLines={14}
          />
        )}
        <Form.Field
          label={sharedMessages.confirmedDownlink}
          name="confirmed"
          component={Checkbox}
        />
        <SubmitBar>
          <Form.Submit component={SubmitButton} message={m.scheduleDownlink} />
        </SubmitBar>
      </Form>
    </>
  )
}

export default DownlinkForm
