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
import { useLocation } from 'react-router-dom'
import Query from 'query-string'
import { defineMessages } from 'react-intl'

import api from '@account/api'

import Button from '@ttn-lw/components/button'
import ButtonGroup from '@ttn-lw/components/button/group'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitButton from '@ttn-lw/components/submit-button'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import style from '@account/views/front/front.styl'

import Yup from '@ttn-lw/lib/yup'
import {
  selectApplicationRootPath,
  selectApplicationSiteName,
  selectApplicationSiteTitle,
} from '@ttn-lw/lib/selectors/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { userId as userIdRegexp } from '@ttn-lw/lib/regexp'

import { selectEnableUserRegistration } from '@account/lib/selectors/app-config'

const m = defineMessages({
  createAccount: 'Create an account',
  forgotPassword: 'Forgot password?',
  loginToContinue: 'Please login to continue',
  loginFailed: 'Login failed',
  accountDeleted: 'Account deleted',
})

const appRoot = selectApplicationRootPath()
const siteName = selectApplicationSiteName()
const siteTitle = selectApplicationSiteTitle()
const enableUserRegistration = selectEnableUserRegistration()

const validationSchema = Yup.object().shape({
  user_id: Yup.string()
    .min(2, Yup.passValues(sharedMessages.validateTooShort))
    .max(36, Yup.passValues(sharedMessages.validateTooLong))
    .matches(userIdRegexp, Yup.passValues(sharedMessages.validateIdFormat))
    .required(sharedMessages.validateRequired)
    .trim(),
  password: Yup.string().required(sharedMessages.validateRequired),
})

const url = (location, omitQuery = false) => {
  const query = Query.parse(location.search)

  const next = query.n || appRoot

  if (omitQuery) {
    return next.split('?')[0]
  }

  return next
}

const Login = () => {
  const [error, setError] = useState(undefined)
  const location = useLocation()

  const handleSubmit = useCallback(
    async (values, { setSubmitting }) => {
      try {
        setError(undefined)

        const castedValues = validationSchema.cast(values)
        await api.account.login(castedValues)

        window.location = url(location)
      } catch (error) {
        setError(error)
        setSubmitting(false)
      }
    },
    [location],
  )

  const initialValues = {
    user_id: '',
    password: '',
  }

  let info
  const next = url(location)

  if (location.state && location.state.info) {
    info = location.state.info
  } else if (!next || (next !== appRoot && !Boolean(error))) {
    info = m.loginToContinue
  } else if ('account-deleted' in Query.parse(location.search)) {
    info = m.accountDeleted
  }

  return (
    <div className={style.form}>
      <IntlHelmet title={sharedMessages.login} />
      <h1 className={style.title}>
        {siteName}
        <br />
        <span className={style.subTitle}>{siteTitle}</span>
      </h1>
      <hr className={style.hRule} />
      <Form
        onSubmit={handleSubmit}
        initialValues={initialValues}
        error={error}
        errorTitle={m.loginFailed}
        info={info}
        validationSchema={validationSchema}
        horizontal={false}
      >
        <Form.Field
          title={sharedMessages.userId}
          name="user_id"
          component={Input}
          autoFocus
          required
        />
        <Form.Field
          title={sharedMessages.password}
          component={Input}
          name="password"
          type="password"
          required
        />
        <ButtonGroup>
          <Form.Submit
            component={SubmitButton}
            message={sharedMessages.login}
            className={style.submitButton}
            error={Boolean(error)}
          />
          {enableUserRegistration && (
            <Button.Link to={`/register${location.search}`} message={m.createAccount} />
          )}
          <Button.Link naked message={m.forgotPassword} to={`/forgot-password${location.search}`} />
        </ButtonGroup>
      </Form>
    </div>
  )
}

export default Login
