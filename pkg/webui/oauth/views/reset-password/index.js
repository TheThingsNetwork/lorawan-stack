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
import { replace } from 'connected-react-router'
import { connect } from 'react-redux'
import * as Yup from 'yup'

import api from '../../api'
import sharedMessages from '../../../lib/shared-messages'

import Button from '../../../components/button'
import Form from '../../../components/form'
import Input from '../../../components/input'
import SubmitButton from '../../../components/submit-button'
import Logo from '../../../components/logo'
import IntlHelmet from '../../../lib/components/intl-helmet'
import Message from '../../../lib/components/message'

import style from '../login/login.styl'

const m = defineMessages({
    loginPage: 'Login Page',
    passwordRequested: 'New temporary password has been sent to user email',
    resetPassword: 'Reset Password',
    requestTempPassword: 'Request a new temporary password',
    stackAccount: 'TTN Stack Account',
  })

const validationSchema = Yup.object().shape({
    user_id: Yup.string()
    .required(sharedMessages.validateRequired),
  })

@connect()
@withRouter
@bind
export default class ResetPassword extends React.PureComponent {
    constructor (props) {
        super(props)
        this.state = {
            error: '',
            info: '',
        }
    }

    async handleSubmit (values, {setSubmitting, setErrors }) {
        try {
            const response = await api.users.reset(values.user_id)
            this.setState({
                error: '',
                info: m.passwordRequested,
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

    navigateToLogin () {
        const { dispatch, location } = this.props
        dispatch(replace('/oauth/login'))
      }

    render() {
        const { error , info } = this.state
        const initialUserId = {
            user_id: '',
        }

        return(
            <div className={style.fullHeightCenter}>
              <IntlHelmet title={m.resetPassword} />
              <div>
                  <div className={style.left}>
                      <div>
                        <Logo />
                        <Message content={m.requestTempPassword} />
                      </div>
                  </div>
                  <div className={style.right}>
                      <h1><Message content={m.stackAccount} /></h1>
                      <Form
                        onSubmit={this.handleSubmit}
                        initialValues={initialUserId}
                        error={this.state.error}
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
                            message={m.resetPassword}
                          />
                          <Button naked secondary message={m.loginPage} onClick={this.navigateToLogin} />
                      </Form>
                  </div>
              </div>
            </div>
        )
    }
}   
