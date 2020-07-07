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
import { Redirect } from 'react-router-dom'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import { push } from 'connected-react-router'
import { connect } from 'react-redux'
import queryString from 'query-string'

import api from '@account/api'

import Button from '@ttn-lw/components/button'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitButton from '@ttn-lw/components/submit-button'
import Checkbox from '@ttn-lw/components/checkbox'

import Message from '@ttn-lw/lib/components/message'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import style from '@account/views/front/front.styl'

import Yup from '@ttn-lw/lib/yup'
import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const m = defineMessages({
  updatePassword: 'Update password',
  newPassword: 'New password',
  oldPassword: 'Old password',
  passwordChanged: 'Password changed',
  revokeAccess: 'Revoke access',
  logoutAllDevices: 'Sign out from all devices',
  revokeWarning: 'This will revoke access from all signed in devices',
  sessionRevoked: 'Your password was changed and all active sessions were revoked',
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

const siteName = selectApplicationSiteName()

const initialValues = {
  password: '',
  confirm: '',
  password_changed: false,
  revoke_all_access: true,
}

@connect(
  undefined,
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
export default class UpdatePassword extends React.PureComponent {
  static propTypes = {
    handleCancelUpdate: PropTypes.func.isRequired,
    handlePasswordChanged: PropTypes.func.isRequired,
    handleSessionRevoked: PropTypes.func.isRequired,
    location: PropTypes.location.isRequired,
  }

  state = {
    error: '',
    info: '',
    revoke_all_access: true,
  }

  @bind
  handleRevokeAllAccess(evt) {
    this.setState({ revoke_all_access: evt.target.checked })
  }

  @bind
  async handleSubmit(values, { resetForm, setSubmitting }) {
    const { handlePasswordChanged, handleSessionRevoked } = this.props
    const userParams = queryString.parse(this.props.location.search)
    const oldPassword = values.old_password ? values.old_password : userParams.current
    const userId = userParams.user

    try {
      try {
        await api.account.me()
      } catch (error) {
        handleSessionRevoked()
        return
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
    const { location, handleCancelUpdate } = this.props

    const { error, info, revoke_all_access } = this.state

    const { user: userParam, current: currentParam } = queryString.parse(location.search)
    if (!Boolean(userParam) || !Boolean(currentParam)) {
      return <Redirect to={{ pathname: '/' }} />
    }

    return (
      <div className={style.form}>
        <IntlHelmet title={m.forgotPassword} />
        <h1 className={style.title}>
          {siteName}
          <br />
          <Message component="strong" content={m.updatePassword} />
        </h1>
        <hr className={style.hRule} />
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
            title={m.newPassword}
            name="password"
            type="password"
            autoComplete="new-password"
            autoFocus
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
          <Button naked secondary message={sharedMessages.cancel} onClick={handleCancelUpdate} />
        </Form>
      </div>
    )
  }
}
