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
import { withRouter } from 'react-router-dom'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import { connect } from 'react-redux'
import { replace } from 'connected-react-router'
import * as Yup from 'yup'

import api from '../../../api'
import sharedMessages from '../../../lib/shared-messages'

import Button from '../../../components/button'
import Field from '../../../components/field'
import Form from '../../../components/form'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'

import style from './create-account.styl'

const m = defineMessages({
  createTtnAccount: 'Create The Things Network Account',
  register: 'Register',
  confirmPassword: 'Confirm Password',
  validatePasswordMatch: 'Passwords should match',
  registrationApproved: 'You have successfully registered and can login now',
  registrationPending: 'You have successfully sent the registration request. Please wait until an admin approves it.',
})

const validationSchema = Yup.object().shape({
  user_id: Yup.string()
    .min(4)
    .max(50)
    .required(sharedMessages.validateRequired),
  name: Yup.string()
    .min(3, sharedMessages.validateTooShort)
    .max(50, sharedMessages.validateTooLong)
    .required(sharedMessages.validateRequired),
  password: Yup.string()
    .min(8)
    .required(sharedMessages.validateRequired),
  primary_email_address: Yup.string()
    .email(sharedMessages.validateEmail)
    .required(sharedMessages.validateRequired),
  password_confirm: Yup.string()
    .oneOf([ Yup.ref('password'), null ], m.validatePasswordMatch)
    .min(8)
    .required(sharedMessages.validateRequired),
})

const getSuccessMessage = function (state) {
  switch (state) {
  case 'STATE_REQUESTED':
    return m.registrationApproved
  case 'STATE_APPROVED':
    return m.registrationApproved
  default:
    return ''
  }
}

@connect()
@withRouter
@bind
export default class CreateAccount extends React.PureComponent {
  constructor (props) {
    super(props)

    this.state = {
      error: '',
      info: '',
    }
  }

  async handleSubmit (values, { setSubmitting, setErrors }) {
    try {
      const result = await api.v3.is.users.create(values)
      this.setState({
        error: '',
        info: getSuccessMessage(result.data.state),
      })

    } catch (error) {
      this.setState({
        error: error.response.data,
        info: '',
      })
    } finally {
      setSubmitting(false)
    }
  }

  handleCancel () {
    const { dispatch, location } = this.props
    const state = location.state || {}

    const back = state.back || '/oauth/login'

    dispatch(replace(back))
  }

  render () {
    const { error, info } = this.state

    return (
      <div className={style.fullHeightCenter}>
        <IntlHelmet>
          <title><Message content={m.register} /></title>
        </IntlHelmet>
        <div className={style.wrapper}>
          <h1><Message content={m.createTtnAccount} /></h1>
          <Form
            onSubmit={this.handleSubmit}
            error={error}
            info={info}
            validationSchema={validationSchema}
          >
            <Field
              required
              title={sharedMessages.userId}
              name="user_id"
              type="text"
              autoComplete="username"
              autoFocus
            />
            <Field
              title={sharedMessages.name}
              name="name"
              type="text"
              autoComplete="name"
            />
            <Field
              required
              title={sharedMessages.email}
              type="text"
              name="primary_email_address"
              autoComplete="email"
            />
            <Field
              required
              title={sharedMessages.password}
              name="password"
              type="password"
              autoComplete="new-password"
            />
            <Field
              required
              title={m.confirmPassword}
              name="password_confirm"
              type="password"
              autoComplete="new-password"
            />
            <Button type="submit" message={m.register} />
            <Button naked secondary message={sharedMessages.cancel} onClick={this.handleCancel} />
          </Form>
        </div>
      </div>
    )
  }
}
