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

import React from 'react'
import { defineMessages } from 'react-intl'

import Notification from '@ttn-lw/components/notification'
import SubmitButton from '@ttn-lw/components/submit-button'
import Input from '@ttn-lw/components/input'
import SubmitBar from '@ttn-lw/components/submit-bar'
import toast from '@ttn-lw/components/toast'
import Form from '@ttn-lw/components/form'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import Yup from '@ttn-lw/lib/yup'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { hexToBase64 } from '@console/lib/bytes'

const m = defineMessages({
  simulateUplink: 'Simulate uplink',
  payloadDescription: 'The desired payload bytes of the uplink message',
  uplinkSuccess: 'Uplink sent',
})

const validationSchema = Yup.object({
  f_port: Yup.number()
    .min(1, Yup.passValues(sharedMessages.validateNumberGte))
    .max(223, Yup.passValues(sharedMessages.validateNumberLte))
    .required(sharedMessages.validateRequired),
  frm_payload: Yup.string().test(
    'len',
    Yup.passValues(sharedMessages.validateHexLength),
    payload => !Boolean(payload) || payload.length % 3 === 0,
  ),
})

const initialValues = { f_port: 1, frm_payload: '' }

const UplinkForm = props => {
  const { simulateUplink, device, skipPayloadCrypto } = props

  const [error, setError] = React.useState('')

  const handleSubmit = React.useCallback(
    async (values, { setSubmitting, resetForm }) => {
      try {
        await simulateUplink({
          f_port: values.f_port,
          frm_payload: hexToBase64(values.frm_payload),
          // `rx_metadata` and `settings` fields are required by the validation middleware in AS.
          // These fields won't affect the result of simulating an uplink message.
          rx_metadata: [
            { gateway_ids: { gateway_id: 'test' }, rssi: 42, channel_rssi: 42, snr: 4.2 },
          ],
          settings: {
            data_rate: { lora: { bandwidth: 125000, spreading_factor: 7 } },
            frequency: 868000000,
          },
        })
        toast({
          title: sharedMessages.success,
          type: toast.types.SUCCESS,
          message: m.uplinkSuccess,
        })
        setSubmitting(false)
      } catch (error) {
        setError(error)
        resetForm({ values })
      }
    },
    [simulateUplink],
  )

  const deviceSimulationDisabled = device.skip_payload_crypto_override ?? skipPayloadCrypto

  return (
    <>
      {deviceSimulationDisabled && (
        <Notification content={sharedMessages.deviceSimulationDisabledWarning} warning small />
      )}
      <IntlHelmet title={m.simulateUplink} />
      <Form
        error={error}
        initialValues={initialValues}
        validationSchema={validationSchema}
        onSubmit={handleSubmit}
        disabled={deviceSimulationDisabled}
      >
        <Form.SubTitle title={m.simulateUplink} />
        <Form.Field
          name="f_port"
          title="FPort"
          component={Input}
          type="number"
          min={1}
          max={223}
          required
        />
        <Form.Field
          name="frm_payload"
          title={sharedMessages.payload}
          description={m.payloadDescription}
          component={Input}
          type="byte"
          unbounded
        />
        <SubmitBar>
          <Form.Submit component={SubmitButton} message={m.simulateUplink} />
        </SubmitBar>
      </Form>
    </>
  )
}

UplinkForm.propTypes = {
  device: PropTypes.device.isRequired,
  simulateUplink: PropTypes.func.isRequired,
  skipPayloadCrypto: PropTypes.bool,
}

UplinkForm.defaultProps = {
  skipPayloadCrypto: false,
}

export default UplinkForm
