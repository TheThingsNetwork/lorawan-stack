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

import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Input from '@ttn-lw/components/input'
import Form from '@ttn-lw/components/form'
import Notification from '@ttn-lw/components/notification'
import Checkbox from '@ttn-lw/components/checkbox'

import Yup from '@ttn-lw/lib/yup'
import diff from '@ttn-lw/lib/diff'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import glossaryIds from '@ttn-lw/lib/constants/glossary-ids'

import { generate16BytesKey } from '@console/lib/device-utils'

import messages from '../messages'

const validationSchema = Yup.object()
  .shape({
    skip_payload_crypto: Yup.boolean().default(false),
    session: Yup.object().when(
      ['skip_payload_crypto', '$mayEditKeys'],
      (skipPayloadCrypto, mayEditKeys, schema) => {
        if (skipPayloadCrypto || !mayEditKeys) {
          return schema.strip()
        }

        return schema.shape({
          keys: Yup.object().shape({
            app_s_key: Yup.object().shape({
              key: Yup.string()
                .length(16 * 2, Yup.passValues(sharedMessages.validateLength)) // A 16 Byte hex.
                .required(sharedMessages.validateRequired),
            }),
          }),
        })
      },
    ),
  })
  .noUnknown()

const ApplicationServerForm = React.memo(props => {
  const { device, onSubmit, onSubmitSuccess, mayEditKeys, mayReadKeys } = props

  const [error, setError] = React.useState('')

  const validationContext = React.useMemo(() => ({ mayEditKeys }), [mayEditKeys])
  const initialValues = React.useMemo(() => {
    const { session = {}, skip_payload_crypto } = device
    const {
      keys = {
        app_s_key: {
          key: '',
        },
      },
    } = session

    return validationSchema.cast(
      {
        skip_payload_crypto,
        session: {
          keys: {
            app_s_key: keys.app_s_key,
          },
        },
      },
      { context: validationContext },
    )
  }, [device, validationContext])

  const formRef = React.useRef(null)
  const sessionRef = React.useRef(device.session)

  const [skipCrypto, setSkipCrypto] = React.useState(device.skip_payload_crypto || false)
  const handleSkipCryptoChange = React.useCallback(
    evt => {
      const { checked } = evt.target
      const { setValues, values } = formRef.current

      setSkipCrypto(checked)
      if (checked) {
        setValues(
          validationSchema.cast(
            {
              ...values,
              skip_payload_crypto: checked,
              session: {
                keys: {
                  app_s_key: {
                    key: '',
                  },
                },
              },
            },
            { context: validationContext },
          ),
        )
      } else {
        setValues(
          validationSchema.cast(
            {
              ...values,
              skip_payload_crypto: checked,
              // Reset initial app_s_key value.
              session: sessionRef.current || '',
            },
            { context: validationContext },
          ),
        )
      }
    },
    [validationContext],
  )

  const onFormSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const castedValues = validationSchema.cast(values, { context: validationContext })
      const updatedValues = diff(initialValues, castedValues)

      setError('')
      try {
        await onSubmit(updatedValues)
        resetForm({ values: castedValues })
        onSubmitSuccess()
      } catch (err) {
        setSubmitting(false)
        setError(err)
      }
    },
    [initialValues, onSubmit, onSubmitSuccess, validationContext],
  )

  // Notify the user that the session keys might be there, but since there are
  // no rights to read the keys we cannot display them.
  const showResetNotification = !mayReadKeys && mayEditKeys && !Boolean(device.session)

  return (
    <Form
      validationSchema={validationSchema}
      validationContext={validationContext}
      initialValues={initialValues}
      onSubmit={onFormSubmit}
      error={error}
      enableReinitialize
      formikRef={formRef}
    >
      {showResetNotification && <Notification content={messages.keysResetWarning} info small />}
      <Form.Field
        autoFocus
        title={sharedMessages.skipCryptoTitle}
        name="skip_payload_crypto"
        description={sharedMessages.skipCryptoDescription}
        component={Checkbox}
        onChange={handleSkipCryptoChange}
      />
      <Form.Field
        required
        title={sharedMessages.appSKey}
        name="session.keys.app_s_key.key"
        type="byte"
        min={16}
        max={16}
        disabled={!mayEditKeys || skipCrypto}
        placeholder={skipCrypto ? sharedMessages.skipCryptoPlaceholder : undefined}
        component={Input.Generate}
        mayGenerateValue={mayEditKeys && !skipCrypto}
        onGenerateValue={generate16BytesKey}
        glossaryId={glossaryIds.APP_SESSION_KEY}
      />
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
      </SubmitBar>
    </Form>
  )
})

ApplicationServerForm.propTypes = {
  device: PropTypes.device.isRequired,
  mayEditKeys: PropTypes.bool.isRequired,
  mayReadKeys: PropTypes.bool.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onSubmitSuccess: PropTypes.func.isRequired,
}

export default ApplicationServerForm
