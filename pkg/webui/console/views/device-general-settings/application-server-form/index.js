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
import * as Yup from 'yup'

import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Input from '@ttn-lw/components/input'
import Form from '@ttn-lw/components/form'
import Notification from '@ttn-lw/components/notification'

import m from '@console/components/device-data-form/messages'

import diff from '@ttn-lw/lib/diff'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import randomByteString from '@console/lib/random-bytes'

import messages from '../messages'

const random16BytesString = () => randomByteString(32)

const validationSchema = Yup.object().shape({
  session: Yup.object().shape({
    keys: Yup.object().shape({
      app_s_key: Yup.object().shape({
        key: Yup.string().emptyOrLength(16 * 2, Yup.passValues(sharedMessages.validateLength)), // A 16 Byte hex.
      }),
    }),
  }),
})

const ApplicationServerForm = React.memo(props => {
  const { device, onSubmit, onSubmitSuccess, mayEditKeys, mayReadKeys } = props

  const [error, setError] = React.useState('')

  const initialValues = React.useMemo(() => {
    const { session = {} } = device
    const {
      keys = {
        app_s_key: {
          key: '',
        },
      },
    } = session

    return {
      session: {
        keys: {
          app_s_key: keys.app_s_key,
        },
      },
    }
  }, [device])

  const onFormSubmit = React.useCallback(
    async (values, { resetForm, setSubmitting }) => {
      const castedValues = validationSchema.cast(values)
      const updatedValues = diff(initialValues, castedValues)

      setError('')
      try {
        await onSubmit(updatedValues)
        resetForm(castedValues)
        onSubmitSuccess()
      } catch (err) {
        setSubmitting(false)
        setError(err)
      }
    },
    [initialValues, onSubmit, onSubmitSuccess],
  )

  // Notify the user that the session keys might be there, but since there are
  // no rights to read the keys we cannot display them.
  const showResetNotification = !mayReadKeys && mayEditKeys && !Boolean(device.session)

  return (
    <Form
      validationSchema={validationSchema}
      initialValues={initialValues}
      onSubmit={onFormSubmit}
      error={error}
      enableReinitialize
      disabled={!mayEditKeys}
    >
      {showResetNotification && <Notification content={messages.keysResetWarning} info small />}
      <Form.Field
        title={sharedMessages.appSKey}
        name="session.keys.app_s_key.key"
        type="byte"
        min={16}
        max={16}
        description={m.appSKeyDescription}
        component={Input.Generate}
        mayGenerateValue={mayEditKeys}
        onGenerateValue={random16BytesString}
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
