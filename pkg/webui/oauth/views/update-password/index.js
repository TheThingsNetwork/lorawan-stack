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

import api from '@oauth/api'

import Button from '@ttn-lw/components/button'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitButton from '@ttn-lw/components/submit-button'
import Checkbox from '@ttn-lw/components/checkbox'
import Spinner from '@ttn-lw/components/spinner'

import Message from '@ttn-lw/lib/components/message'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import style from '@oauth/views/create-account/create-account.styl'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const m = defineMessages({
  newPassword: 'New Password',
  oldPassword: 'Old Password',
  passwordChanged: 'Password changed',
  revokeAccess: 'Revoke Access',
  logoutAllDevices: 'Log out from all end devices',
  revokeWarning: 'This will revoke access from all logged in devices',
  sessionRevoked: 'Session revoked',
})

const validationSchema = Yup.object().shape({
  password: Yup.string()
    .min(8, Yup.passValues(sharedMessages.validateTooShort))
    .required(sharedMessages.validateRequired),
  confirm: Yup.string()
    .oneOf([Yup.ref('password'), null], sharedMessages.validatePasswordMatch)
    .min(8, Yup.passValues(sharedMessages.validateTooShort))
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
    handleCancelUpdate: PropTypes.func.isRequired,
    handlePasswordChanged: PropTypes.func.isRequired,
    handleSessionRevoked: PropTypes.func.isRequired,
    location: PropTypes.location.isRequired,
    user: PropTypes.user,
  }

  static defaultProps = {
    user: undefined,
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
          <Message content={sharedMessages.fetching} />
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
                title={m.revokeAccess}
                name="revoke_all_access"
                label={m.logoutAllDevices}
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
