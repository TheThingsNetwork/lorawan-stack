// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
import { defineMessages } from 'react-intl'
import { useNavigate, useSearchParams } from 'react-router-dom'

import tts from '@account/api/tts'

import Button from '@ttn-lw/components/button'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitButton from '@ttn-lw/components/submit-button'
import ButtonGroup from '@ttn-lw/components/button/group'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import Message from '@ttn-lw/lib/components/message'

import style from '@account/views/front/front.styl'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { userId as userIdRegexp } from '@ttn-lw/lib/regexp'
import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'

const m = defineMessages({
  forgotPassword: 'Forgot password',
  passwordRequested:
    'An email with reset instruction was sent to the email address associated with your username. Please check your spam folder as well.',
  send: 'Send',
  resetPassword: 'Reset password',
  resetPasswordDescription: 'Please enter your User ID to receive an email with reset instructions',
})

const validationSchema = Yup.object().shape({
  user_id: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(36, Yup.passValues(sharedMessages.validateTooLong))
    .matches(userIdRegexp, Yup.passValues(sharedMessages.validateIdFormat))
    .required(sharedMessages.validateRequired),
})

const initialValues = { user_id: '' }

const siteName = selectApplicationSiteName()

const ForgotPassword = () => {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [error, setError] = useState(undefined)

  const handleSubmit = useCallback(
    async (values, { setSubmitting }) => {
      try {
        setError(undefined)
        await tts.Users.createTemporaryPassword(values.user_id)
        navigate(`/login?${searchParams.toString()}`, {
          state: {
            info: m.passwordRequested,
          },
        })
      } catch (error) {
        setSubmitting(false)
        setError(error)
      }
    },
    [navigate, searchParams],
  )

  return (
    <div className={style.form}>
      <IntlHelmet title={m.forgotPassword} />
      <h1 className={style.title}>
        {siteName}
        <br />
        <Message component="strong" content={m.resetPassword} />
      </h1>
      <hr className={style.hRule} />
      <Message content={m.resetPasswordDescription} component="p" className={style.description} />
      <Form
        onSubmit={handleSubmit}
        initialValues={initialValues}
        error={error}
        validationSchema={validationSchema}
      >
        <Form.Field
          title={sharedMessages.userId}
          name="user_id"
          component={Input}
          autoFocus
          required
        />
        <ButtonGroup>
          <Form.Submit component={SubmitButton} message={m.send} className={style.submitButton} />
          <Button.Link
            naked
            message={sharedMessages.cancel}
            to={`/login?${searchParams.toString()}`}
          />
        </ButtonGroup>
      </Form>
    </div>
  )
}

export default ForgotPassword
