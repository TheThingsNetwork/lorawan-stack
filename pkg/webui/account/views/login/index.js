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
import { withRouter } from 'react-router-dom'
import bind from 'autobind-decorator'
import Query from 'query-string'
import { defineMessages } from 'react-intl'

import api from '@account/api'

import Button from '@ttn-lw/components/button'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitButton from '@ttn-lw/components/submit-button'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import style from '@account/views/front/front.styl'

import Yup from '@ttn-lw/lib/yup'
import PropTypes from '@ttn-lw/lib/prop-types'
import {
  selectApplicationRootPath,
  selectApplicationSiteName,
  selectApplicationSiteTitle,
} from '@ttn-lw/lib/selectors/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { id as userRegexp } from '@ttn-lw/lib/regexp'

const m = defineMessages({
  createAccount: 'Create an account',
  forgotPassword: 'Forgot password?',
  loginToContinue: 'Please login to continue',
  loginFailed: 'Login failed',
})

const appRoot = selectApplicationRootPath()
const siteName = selectApplicationSiteName()
const siteTitle = selectApplicationSiteTitle()

const validationSchema = Yup.object().shape({
  user_id: Yup.string()
    .min(3, Yup.passValues(sharedMessages.validateTooShort))
    .max(36, Yup.passValues(sharedMessages.validateTooLong))
    .matches(userRegexp, Yup.passValues(sharedMessages.validateIdFormat))
    .required(sharedMessages.validateRequired),
  password: Yup.string().required(sharedMessages.validateRequired),
})

@withRouter
export default class Login extends React.PureComponent {
  static propTypes = {
    location: PropTypes.location.isRequired,
  }

  constructor(props) {
    super(props)
    this.state = {
      error: '',
    }
  }

  @bind
  async handleSubmit(values, { setSubmitting }) {
    try {
      await api.account.login(values)

      window.location = url(this.props.location)
    } catch (error) {
      this.setState({
        error,
      })
      setSubmitting(false)
    }
  }

  render() {
    const initialValues = {
      user_id: '',
      password: '',
    }

    let info
    const { location } = this.props
    const next = url(location)

    if (location.state && location.state.info) {
      info = location.state.info
    } else if (!next || (next !== appRoot && !Boolean(this.state.error))) {
      info = m.loginToContinue
    }

    return (
      <React.Fragment>
        <div className={style.form}>
          <IntlHelmet title={sharedMessages.login} />
          <h1 className={style.title}>
            {siteName}
            <br />
            <span className={style.subTitle}>{siteTitle}</span>
          </h1>
          <hr className={style.hRule} />
          <Form
            onSubmit={this.handleSubmit}
            initialValues={initialValues}
            error={this.state.error}
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
            <div className={style.buttons}>
              <Form.Submit
                component={SubmitButton}
                message={sharedMessages.login}
                className={style.submitButton}
                alwaysEnabled
              />
              <Button.Link
                to={`/register${location.search}`}
                secondary
                message={m.createAccount}
                className={style.registerButton}
              />
              <Button.Link naked secondary message={m.forgotPassword} to="/forgot-password" />
            </div>
          </Form>
        </div>
      </React.Fragment>
    )
  }
}

function url(location, omitQuery = false) {
  const query = Query.parse(location.search)

  const next = query.n || appRoot

  if (omitQuery) {
    return next.split('?')[0]
  }

  return next
}
