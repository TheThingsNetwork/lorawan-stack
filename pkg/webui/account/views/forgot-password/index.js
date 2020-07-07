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
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import { push } from 'connected-react-router'
import { connect } from 'react-redux'

import api from '@account/api'

import Button from '@ttn-lw/components/button'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitButton from '@ttn-lw/components/submit-button'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import Message from '@ttn-lw/lib/components/message'

import style from '@account/views/front/front.styl'

import Yup from '@ttn-lw/lib/yup'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import { id as userRegexp } from '@ttn-lw/lib/regexp'
import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'

const m = defineMessages({
  forgotPassword: 'Forgot password',
  passwordRequested:
    'An email with reset instruction was sent to the email address associated with your username. Please check your spam folder as well.',
  goToLogin: 'Go to login',
  send: 'Send',
  requestTempPassword: 'Reset password',
  resetPasswordDescription: 'Please enter your User ID to receive an email with reset instructions',
})

const validationSchema = Yup.object().shape({
  user_id: Yup.string()
    .min(3, Yup.passValues(sharedMessages.validateTooShort))
    .max(36, Yup.passValues(sharedMessages.validateTooLong))
    .matches(userRegexp, Yup.passValues(sharedMessages.validateIdFormat))
    .required(sharedMessages.validateRequired),
})

const initialValues = { user_id: '' }

const siteName = selectApplicationSiteName()

@connect(
  undefined,
  {
    handleCancel: () => push('/login'),
  },
)
export default class ForgotPassword extends React.PureComponent {
  static propTypes = {
    handleCancel: PropTypes.func.isRequired,
  }

  state = {
    error: '',
    info: '',
    requested: false,
  }

  @bind
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
        error,
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
      <div className={style.form}>
        <IntlHelmet title={m.forgotPassword} />
        <h1 className={style.title}>
          {siteName}
          <br />
          <Message component="strong" content={m.requestTempPassword} />
        </h1>
        <hr className={style.hRule} />
        <Message content={m.resetPasswordDescription} component="p" className={style.description} />
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
          <Form.Submit
            component={SubmitButton}
            message={m.send}
            className={style.submitButton}
            alwaysEnabled
          />
          <Button naked secondary message={cancelButtonText} onClick={handleCancel} />
        </Form>
      </div>
    )
  }
}
