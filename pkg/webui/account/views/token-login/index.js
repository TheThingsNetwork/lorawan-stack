// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

const m = defineMessages({
  loginToken: 'Login Token',
})

const appRoot = selectApplicationRootPath()
const siteName = selectApplicationSiteName()
const siteTitle = selectApplicationSiteTitle()

const validationSchema = Yup.object().shape({
  token: Yup.string().required(sharedMessages.validateRequired),
})

const url = (location, omitQuery = false) => {
  const query = Query.parse(location.search)

  const next = query.n || appRoot

  if (omitQuery) {
    return next.split('?')[0]
  }

  return next
}

const TokenLogin = () => {
  const [error, setError] = useState(undefined)
  const location = useLocation()
  const { token: tokenParam } = Query.parse(location.search)

  const handleSubmit = useCallback(
    async (values, { setSubmitting }) => {
      try {
        setError(undefined)
        await api.account.tokenLogin(values)

        window.location = url(location)
      } catch (error) {
        setError(error)
        setSubmitting(false)
      }
    },
    [location],
  )

  const initialValues = {
    token: tokenParam ? tokenParam : '',
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
        errorTitle={sharedMessages.loginFailed}
        validationSchema={validationSchema}
        horizontal={false}
      >
        <Form.Field title={m.loginToken} component={Input} name="token" type="password" required />
        <div className={style.buttons}>
          <Form.Submit
            component={SubmitButton}
            message={sharedMessages.login}
            className={style.submitButton}
          />
        </div>
      </Form>
    </div>
  )
}

export default TokenLogin
