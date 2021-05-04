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

import React, { useCallback, useState } from 'react'
import { useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'

import tts from '@account/api/tts'

import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import Checkbox from '@ttn-lw/components/checkbox'
import Button from '@ttn-lw/components/button'
import toast from '@ttn-lw/components/toast'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import createPasswordValidationSchema from '@ttn-lw/lib/create-password-validation-schema'

import { selectUserId } from '@account/store/selectors/user'
import { selectPasswordRequirements } from '@account/store/selectors/identity-server'

const m = defineMessages({
  currentPassword: 'Current password',
  newPassword: 'New password',
  newPasswordConfirm: 'Confirm new password',
  revokeAllAccess: 'Revoke all access',
  revokeAllAccessDescription:
    'Check this option if you also wish to revoke all your currently active sessions',
})

const validationSchemaOldPassword = Yup.object().shape({
  old: Yup.string().required(sharedMessages.validateRequired).default(''),
  revoke_all_access: Yup.bool().default(false),
})

const validationSchemaTemporaryPassword = Yup.object().shape({
  revoke_all_access: Yup.bool().default(false),
})

const ChangePasswordForm = ({ userId, old, cancelRoute, onSubmitSuccess }) => {
  const selectedUserId = useSelector(state => userId || selectUserId(state))
  const passwordRequirements = useSelector(selectPasswordRequirements)
  const [error, setError] = useState(undefined)
  const usesTemporaryPw = Boolean(old)
  const baseValidationSchema = usesTemporaryPw
    ? validationSchemaTemporaryPassword
    : validationSchemaOldPassword

  const validationSchema = baseValidationSchema.concat(
    createPasswordValidationSchema(passwordRequirements),
  )

  const handleSubmit = useCallback(
    async (values, { resetForm, setSubmitting }) => {
      setError(undefined)
      try {
        await tts.Users.updatePasswordById(selectedUserId, {
          old: values.old || old,
          new: values.password,
          revoke_all_access: values.revoke_all_access,
        })

        resetForm({ values: validationSchema.cast({}) })
        onSubmitSuccess(values.revoke_all_access)
      } catch (error) {
        setSubmitting(false)
        setError(error)
      }
    },
    [selectedUserId, old, onSubmitSuccess, validationSchema],
  )

  return (
    <Form
      initialValues={validationSchema.cast({ old })}
      validationSchema={validationSchema}
      onSubmit={handleSubmit}
      error={error}
      validateOnChange
    >
      {!usesTemporaryPw && (
        <Form.Field
          name="old"
          component={Input}
          title={m.currentPassword}
          type="password"
          required
        />
      )}
      <Form.Field
        name="password"
        component={Input}
        title={m.newPassword}
        type="password"
        required
      />
      <Form.Field
        name="confirmPassword"
        component={Input}
        title={m.newPasswordConfirm}
        type="password"
        required
      />
      <Form.Field
        name="revoke_all_access"
        component={Checkbox}
        title={m.revokeAllAccess}
        description={m.revokeAllAccessDescription}
      />
      {usesTemporaryPw ? (
        <>
          <Form.Submit component={SubmitButton} message={sharedMessages.changePassword} />
          <Button.Link to={cancelRoute} naked secondary message={sharedMessages.cancel} />
        </>
      ) : (
        <SubmitBar>
          <div>
            <Form.Submit component={SubmitButton} message={sharedMessages.changePassword} />
          </div>
        </SubmitBar>
      )}
    </Form>
  )
}

ChangePasswordForm.propTypes = {
  cancelRoute: PropTypes.string,
  old: PropTypes.string,
  onSubmitSuccess: PropTypes.func,
  userId: PropTypes.string,
}

ChangePasswordForm.defaultProps = {
  old: undefined,
  cancelRoute: undefined,
  onSubmitSuccess: () => {
    toast({
      title: m.success,
      message: sharedMessages.passwordChanged,
      type: toast.types.SUCCESS,
    })
  },
  userId: undefined,
}

export default ChangePasswordForm
