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
import { Redirect } from 'react-router-dom'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import { push } from 'connected-react-router'
import { connect } from 'react-redux'
import * as Yup from 'yup'
import queryString from 'query-string'

import api from '../../api'
import sharedMessages from '../../../lib/shared-messages'
import { digit, uppercase, special } from '../../lib/regexp'

import Button from '../../../components/button'
import Form from '../../../components/form'
import Input from '../../../components/input'
import SubmitButton from '../../../components/submit-button'
import IntlHelmet from '../../../lib/components/intl-helmet'
import Message from '../../../lib/components/message'
import style from '../create-account/create-account.styl'
import Checkbox from '../../../components/checkbox'
import Spinner from '../../../components/spinner'
import PropTypes from '../../../lib/prop-types'

const m = defineMessages({
  newPassword: 'New Password',
  oldPassword: 'Old Password',
  passwordChanged: 'Password changed successfully',
  revokeAllAccess: 'Log out from all devices',
  revokeWarning: 'This will revoke access from all logged in devices',
  sessionRevoked: 'Your session is revoked',
})

const validationSchema = Yup.object().shape({
  password: Yup.string()
    .min(8)
    .matches(digit, sharedMessages.validatePasswordDigit)
    .matches(uppercase, sharedMessages.validatePasswordUppercase)
    .matches(special, sharedMessages.validatePasswordSpecial)
    .required(sharedMessages.validateRequired),
  confirm: Yup.string()
    .oneOf([Yup.ref('password'), null], sharedMessages.validatePasswordMatch)
    .min(8)
    .required(sharedMessages.validateRequired),
})

const initialValues = {
  password: '',
  confirm: '',
  password_changed: false,
  revoke_all_access: true,
}

@connect(
  state => ({
    fetching: state.user.fetching,
    user: state.user.user,
  }),
  {
    handleCancelUpdate: () => push('/'),
    handlePasswordChanged: () =>
      push('/login', {
        info: m.passwordChanged,
      }),
    handleSessionRevoked: () =>
      push('/login', {
        info: m.sessionRevoked,
      }),
  },
)
@bind
export default class UpdatePassword extends React.PureComponent {
  static propTypes = {
    fetching: PropTypes.bool.isRequired,
    user: PropTypes.object,
    handleCancelUpdate: PropTypes.func.isRequired,
    handlePasswordChanged: PropTypes.func.isRequired,
    handleSessionRevoked: PropTypes.func.isRequired,
  }

  state = {
    error: '',
    info: '',
    revoke_all_access: true,
  }

  handleRevokeAllAccess(evt) {
    this.setState({ revoke_all_access: evt.target.checked })
  }

  async handleSubmit(values, { resetForm, setSubmitting }) {
    const { user, handlePasswordChanged, handleSessionRevoked } = this.props
    const userParams = queryString.parse(this.props.location.search)
    const oldPassword = values.old_password ? values.old_password : userParams.current
    const userId = Boolean(user) ? user.ids.user_id : userParams.user

    try {
      if (Boolean(user)) {
        try {
          await api.oauth.me()
        } catch (error) {
          handleSessionRevoked()
          return
        }
      }
      await api.users.updatePassword(userId, {
        user_ids: { user_id: userId },
        new: values.password,
        old: oldPassword,
        revoke_all_access: values.revoke_all_access,
      })

      handlePasswordChanged()
    } catch (error) {
      this.setState({
        error: error.response.data,
        info: '',
      })
      setSubmitting(false)
    }
  }

  render() {
    const { user, fetching, location, handleCancelUpdate } = this.props

    const { error, info, revoke_all_access } = this.state

    if (fetching) {
      return (
        <Spinner center>
          <Message content={sharedMessages.loading} />
        </Spinner>
      )
    }

    const { user: userParam, current: currentParam } = queryString.parse(location.search)
    if (!Boolean(user) && (!Boolean(userParam) || !Boolean(currentParam))) {
      return <Redirect to={{ pathname: '/' }} />
    }

    let oldPasswordField
    if (Boolean(user)) {
      oldPasswordField = (
        <Form.Field
          component={Input}
          required
          title={m.oldPassword}
          name="old_password"
          type="password"
          autoComplete="old-password"
          autoFocus
        />
      )
    }

    return (
      <Container className={style.fullHeight}>
        <Row justify="center" align="center" className={style.fullHeight}>
          <Col sm={12} md={8} lg={5}>
            <IntlHelmet title={sharedMessages.changePassword} />
            <Message
              content={sharedMessages.changePassword}
              component="h1"
              className={style.title}
            />
            <Form
              onSubmit={this.handleSubmit}
              initialValues={initialValues}
              error={error}
              info={info}
              validationSchema={validationSchema}
              horizontal={false}
            >
              {oldPasswordField}
              <Form.Field
                component={Input}
                required
                title={m.newPassword}
                name="password"
                type="password"
                autoComplete="new-password"
                autoFocus={!Boolean(user)}
              />
              <Form.Field
                component={Input}
                required
                title={sharedMessages.confirmPassword}
                name="confirm"
                type="password"
                autoComplete="new-password"
              />
              <Form.Field
                onChange={this.handleRevokeAllAccess}
                warning={revoke_all_access ? m.revokeWarning : undefined}
                name="revoke_all_access"
                label={m.revokeAllAccess}
                component={Checkbox}
              />
              <Form.Submit component={SubmitButton} message={sharedMessages.changePassword} />
              <Button
                naked
                secondary
                message={sharedMessages.cancel}
                onClick={handleCancelUpdate}
              />
            </Form>
          </Col>
        </Row>
      </Container>
    )
  }
}
