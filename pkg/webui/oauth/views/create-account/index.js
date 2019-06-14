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
import { Container, Col, Row } from 'react-grid-system'
import { withRouter } from 'react-router-dom'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import { connect } from 'react-redux'
import { replace } from 'connected-react-router'
import * as Yup from 'yup'

import api from '../../api'
import sharedMessages from '../../../lib/shared-messages'

import Button from '../../../components/button'
import Input from '../../../components/input'
import Form from '../../../components/form'
import SubmitButton from '../../../components/submit-button'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'

import style from './create-account.styl'

const m = defineMessages({
  createAccount: 'Create TTN Stack Account',
  register: 'Register',
  goToLogin: 'Go to login',
  confirmPassword: 'Confirm Password',
  validatePasswordMatch: 'Passwords should match',
  validatePasswordDigit: 'Should contain at least one digit',
  validatePasswordUppercase: 'Should contain at least one uppercase letter',
  validatePasswordSpecial: 'Should contain at least one special character',
  registrationApproved: 'You have successfully registered and can login now',
  registrationPending: 'You have successfully sent the registration request. Please wait until an admin approves it.',
})

const digit = /(?=.*[\d])/
const uppercase = /(?=.*[A-Z])/
const special = /(?=.*[!@#$%^&*])/

const validationSchema = Yup.object().shape({
  user_id: Yup.string()
    .min(2)
    .max(36)
    .required(sharedMessages.validateRequired),
  name: Yup.string()
    .min(3, sharedMessages.validateTooShort)
    .max(50, sharedMessages.validateTooLong),
  password: Yup.string()
    .min(8)
    .matches(digit, m.validatePasswordDigit)
    .matches(uppercase, m.validatePasswordUppercase)
    .matches(special, m.validatePasswordSpecial)
    .required(sharedMessages.validateRequired),
  primary_email_address: Yup.string()
    .email(sharedMessages.validateEmail)
    .required(sharedMessages.validateRequired),
  password_confirm: Yup.string()
    .oneOf([ Yup.ref('password'), null ], m.validatePasswordMatch)
    .min(8)
    .required(sharedMessages.validateRequired),
})

const initialValues = {
  user_id: '',
  name: '',
  primary_email_address: '',
  password: '',
  password_confirm: '',
}

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
      registered: false,
    }
  }

  async handleSubmit (values, { setSubmitting, setErrors }) {
    try {
      const { user_id, ...rest } = values
      const result = await api.users.register({
        user: { ids: { user_id }, ...rest },
      })

      this.setState({
        error: '',
        info: getSuccessMessage(result.data.state),
        registered: true,
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
    const { error, info, registered } = this.state
    const cancelButtonText = registered ? m.goToLogin : sharedMessages.cancel

    return (
      <Container className={style.fullHeight}>
        <Row justify="center" align="center" className={style.fullHeight}>
          <Col sm={12} md={8} lg={5}>
            <IntlHelmet title={m.register} />
            <Message content={m.createAccount} component="h1" className={style.title} />
            <Form
              onSubmit={this.handleSubmit}
              initialValues={initialValues}
              error={error}
              info={info}
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
                title={m.confirmPassword}
                name="password_confirm"
                type="password"
                autoComplete="new-password"
                component={Input}
              />
              <Form.Submit
                component={SubmitButton}
                message={m.register}
              />
              <Button naked secondary message={cancelButtonText} onClick={this.handleCancel} />
            </Form>
          </Col>
        </Row>
      </Container>
    )
  }
}
