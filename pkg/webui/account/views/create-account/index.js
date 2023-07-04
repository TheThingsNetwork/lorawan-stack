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

import React, { useState, useCallback } from 'react'
import { useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'
import { Navigate, useNavigate, useSearchParams } from 'react-router-dom'

import tts from '@account/api/tts'

import Button from '@ttn-lw/components/button'
import Input from '@ttn-lw/components/input'
import Form from '@ttn-lw/components/form'
import SubmitButton from '@ttn-lw/components/submit-button'
import Spinner from '@ttn-lw/components/spinner'
import ButtonGroup from '@ttn-lw/components/button/group'

import Message from '@ttn-lw/lib/components/message'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import style from '@account/views/front/front.styl'

import Yup from '@ttn-lw/lib/yup'
import { selectApplicationSiteName, selectEnableUserRegistration } from '@ttn-lw/lib/selectors/env'
import { userId as userIdRegexp } from '@ttn-lw/lib/regexp'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import createPasswordValidationSchema from '@ttn-lw/lib/create-password-validation-schema'
import useRequest from '@ttn-lw/lib/hooks/use-request'

import { getIsConfiguration } from '@account/store/actions/identity-server'

import { selectPasswordRequirements } from '@account/store/selectors/identity-server'

const m = defineMessages({
  registrationApproved: 'You have successfully registered and can login now',
  createAccount: 'Create account',
  createANewAccount: 'Create a new account',
  registrationPending:
    'You have successfully sent the registration request. You will receive a confirmation once the account has been approved.',
})

const baseValidationSchema = Yup.object().shape({
  user_id: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(36, Yup.passValues(sharedMessages.validateTooLong))
    .matches(userIdRegexp, Yup.passValues(sharedMessages.validateIdFormat))
    .required(sharedMessages.validateRequired),
  name: Yup.string()
    .min(3, Yup.passValues(sharedMessages.validateTooShort))
    .max(50, Yup.passValues(sharedMessages.validateTooLong)),
  primary_email_address: Yup.string()
    .email(sharedMessages.validateEmail)
    .required(sharedMessages.validateRequired),
})

const initialValues = {
  user_id: '',
  name: '',
  primary_email_address: '',
  password: '',
  confirmPassword: '',
}

const siteName = selectApplicationSiteName()
const enableUserRegistration = selectEnableUserRegistration()

const getSuccessMessage = state => {
  switch (state) {
    case undefined:
      // Zero value is swallowed by the backend, but means STATE_REQUESTED
      return m.registrationPending
    case 'STATE_APPROVED':
      return m.registrationApproved
    default:
      return ''
  }
}

const CreateAccount = () => {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const [fetching, isConfigError] = useRequest(getIsConfiguration())
  if (Boolean(isConfigError)) {
    throw isConfigError
  }

  const [error, setError] = useState(undefined)
  const passwordRequirements = useSelector(selectPasswordRequirements)

  const handleSubmit = useCallback(
    async (values, { setSubmitting }) => {
      try {
        setError(undefined)
        const { user_id, ...rest } = values
        const invitation_token = searchParams.get('invitation_token') || ''
        const result = await tts.Users.create({ ids: { user_id }, ...rest }, invitation_token)

        navigate(`/login?${searchParams.toString()}`, {
          state: {
            info: getSuccessMessage(result.state),
          },
        })
      } catch (error) {
        setSubmitting(false)
        setError(error)
      }
    },
    [navigate, searchParams],
  )

  if (!enableUserRegistration) {
    return <Navigate to={`/login?${searchParams.toString()}`} />
  }

  if (fetching) {
    return (
      <Spinner center>
        <Message content={sharedMessages.fetching} />
      </Spinner>
    )
  }

  const validationSchema = baseValidationSchema.concat(
    createPasswordValidationSchema(passwordRequirements),
  )

  return (
    <>
      <div className={style.form}>
        <IntlHelmet title={m.createANewAccount} />
        <h1 className={style.title}>
          {siteName}
          <br />
          <Message content={m.createANewAccount} component="strong" />
        </h1>
        <hr className={style.hRule} />
        <Form
          onSubmit={handleSubmit}
          initialValues={initialValues}
          error={error}
          validationSchema={validationSchema}
          horizontal={false}
        >
          <Form.Field
            component={Input}
            required
            title={sharedMessages.userId}
            name="user_id"
            autoComplete="username"
            autoFocus
          />
          <Form.Field
            title={sharedMessages.name}
            name="name"
            component={Input}
            autoComplete="name"
          />
          <Form.Field
            required
            title={sharedMessages.email}
            component={Input}
            name="primary_email_address"
            autoComplete="email"
          />
          <Form.Field
            required
            title={sharedMessages.password}
            name="password"
            type="password"
            component={Input}
            autoComplete="new-password"
          />
          <Form.Field
            required
            title={sharedMessages.confirmPassword}
            name="confirmPassword"
            type="password"
            autoComplete="new-password"
            component={Input}
          />
          <ButtonGroup>
            <Form.Submit
              component={SubmitButton}
              message={m.createAccount}
              className={style.submitButton}
            />
            <Button.Link
              to={`/login?${searchParams.toString()}`}
              naked
              message={sharedMessages.login}
            />
          </ButtonGroup>
        </Form>
      </div>
    </>
  )
}

export default CreateAccount
