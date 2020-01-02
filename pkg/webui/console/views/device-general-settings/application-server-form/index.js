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

import SubmitButton from '../../../../components/submit-button'
import SubmitBar from '../../../../components/submit-bar'
import Input from '../../../../components/input'
import Form from '../../../../components/form'
import DevAddrInput from '../../../containers/dev-addr-input'

import diff from '../../../../lib/diff'
import m from '../../../components/device-data-form/messages'
import randomByteString from '../../../lib/random-bytes'
import PropTypes from '../../../../lib/prop-types'
import sharedMessages from '../../../../lib/shared-messages'

const random16BytesString = () => randomByteString(32)
const toUndefined = value => (!Boolean(value) ? undefined : value)

const validationSchema = Yup.object().shape({
  session: Yup.object().shape({
    dev_addr: Yup.string()
      .length(4 * 2, m.validate8) // 4 Byte hex
      .required(sharedMessages.validateRequired),
    keys: Yup.object().shape({
      app_s_key: Yup.object().shape({
        key: Yup.string()
          .emptyOrLength(16 * 2, m.validate32) // 16 Byte hex
          .transform(toUndefined)
          .default(random16BytesString),
      }),
    }),
  }),
})

const ApplicationServerForm = React.memo(props => {
  const { device, onSubmit, onSubmitSuccess } = props

  const [error, setError] = React.useState('')

  const initialValues = React.useMemo(() => {
    const { session = {}, ids } = device
    const {
      keys = {
        app_s_key: {},
      },
    } = session

    return {
      session: {
        dev_addr: session.dev_addr || ids.dev_addr,
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

  return (
    <Form
      validationSchema={validationSchema}
      initialValues={initialValues}
      onSubmit={onFormSubmit}
      error={error}
      enableReinitialize
    >
      <DevAddrInput
        title={sharedMessages.devAddr}
        name="session.dev_addr"
        placeholder={m.leaveBlankPlaceholder}
        description={m.deviceAddrDescription}
        disabled
        required
      />
      <Form.Field
        title={sharedMessages.appSKey}
        name="session.keys.app_s_key.key"
        type="byte"
        min={16}
        max={16}
        placeholder={m.leaveBlankPlaceholder}
        description={m.appSKeyDescription}
        component={Input}
      />
      <SubmitBar>
        <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
      </SubmitBar>
    </Form>
  )
})

ApplicationServerForm.propTypes = {
  device: PropTypes.device.isRequired,
  onSubmit: PropTypes.func.isRequired,
  onSubmitSuccess: PropTypes.func.isRequired,
}

export default ApplicationServerForm
