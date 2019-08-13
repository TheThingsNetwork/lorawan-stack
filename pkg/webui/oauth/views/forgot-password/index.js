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
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import { push } from 'connected-react-router'
import { connect } from 'react-redux'
import * as Yup from 'yup'

import api from '../../api'
import sharedMessages from '../../../lib/shared-messages'

import Button from '../../../components/button'
import Form from '../../../components/form'
import Input from '../../../components/input'
import SubmitButton from '../../../components/submit-button'
import IntlHelmet from '../../../lib/components/intl-helmet'
import Message from '../../../lib/components/message'
import PropTypes from '../../../lib/prop-types'

import style from './forgot-password.styl'

const m = defineMessages({
  loginPage: 'Login Page',
  forgotPassword: 'Forgot Password',
  passwordRequested: 'You will receive an email with reset instructions shortly.',
  goToLogin: 'Go to Login',
  send: 'Send',
  resetPasswordDescription:
    'Please enter your username to receive an email with reset instructions',
  requestTempPassword: 'Reset Password',
})

const validationSchema = Yup.object().shape({
  user_id: Yup.string().required(sharedMessages.validateRequired),
})

const initialValues = { user_id: '' }

@connect(
  null,
  {
    handleCancel: () => push('/login'),
  },
)
@bind
export default class ForgotPassword extends React.PureComponent {
  static propTypes = {
    handleCancel: PropTypes.func.isRequired,
  }

  state = {
    error: '',
    info: '',
    requested: false,
  }

  async handleSubmit(values, { setSubmitting }) {
    try {
      await api.users.resetPassword(values.user_id)
      this.setState({
        error: '',
        info: m.passwordRequested,
        requested: true,
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

  render() {
    const { error, info, requested } = this.state
    const { handleCancel } = this.props
    const cancelButtonText = requested ? m.goToLogin : sharedMessages.cancel

    return (
      <Container className={style.fullHeight}>
        <Row justify="center" align="center" className={style.fullHeight}>
          <Col sm={12} md={8} lg={5}>
            <IntlHelmet title={m.forgotPassword} />
            <Message content={m.requestTempPassword} component="h1" className={style.title} />
            <Message
              content={m.resetPasswordDescription}
              component="h4"
              className={style.description}
            />
            <Form
              onSubmit={this.handleSubmit}
              initialValues={initialValues}
              error={error}
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
              <Form.Submit component={SubmitButton} message={m.send} />
              <Button naked secondary message={cancelButtonText} onClick={handleCancel} />
            </Form>
          </Col>
        </Row>
      </Container>
    )
  }
}
