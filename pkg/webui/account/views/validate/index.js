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
import { defineMessages } from 'react-intl'
import queryString from 'query-string'

import api from '@account/api'

import Spinner from '@ttn-lw/components/spinner'
import ErrorNotification from '@ttn-lw/components/error-notification'
import Notification from '@ttn-lw/components/notification'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import Logo from '@console/containers/logo'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { isNotFoundError, createFrontendError } from '@ttn-lw/lib/errors/utils'

import style from './validate.styl'

const m = defineMessages({
  validateSuccess: 'Contact info validated',
  validateFail: 'There was an error and the contact info could not be validated',
  goToLogin: 'Go to login',
  tokenNotFoundTitle: 'Token not found',
  tokenNotFoundMessage:
    'The validation token was not found. This could mean that the contact info has already been validated. Otherwise, please contact an administrator.',
  contactInfoValidation: 'Contact info validation',
})

export default class Validate extends React.PureComponent {
  static propTypes = {
    location: PropTypes.location.isRequired,
  }

  state = {
    error: undefined,
    success: undefined,
    fetching: true,
  }

  handleError(error) {
    this.setState({ error, fetching: false, success: undefined })
  }

  handleSuccess() {
    this.setState({ success: m.validateSuccess, fetching: false, error: undefined })
  }

  async componentDidMount() {
    const validationData = queryString.parse(this.props.location.search)
    try {
      await api.users.validate({
        token: validationData.token,
        id: validationData.reference,
      })
      this.handleSuccess()
    } catch (error) {
      if (error.response && error.response.data) {
        if (isNotFoundError(error.response.data)) {
          this.handleError(createFrontendError(m.tokenNotFoundTitle, m.tokenNotFoundMessage))
        } else {
          this.handleError(error.response.data)
        }
      } else {
        this.handleError(error)
      }
    }
  }

  render() {
    const { fetching, error, success } = this.state

    if (fetching) {
      return (
        <Spinner center>
          <Message content={sharedMessages.fetching} />
        </Spinner>
      )
    }

    return (
      <div className={style.center}>
        <Logo className={style.logo} />
        <Message component="h3" content={m.contactInfoValidation} className={style.heading} />
        {error && <ErrorNotification content={error} buttonIcon={'error'} small />}
        {success && (
          <Notification
            large
            success
            content={success}
            title={m.validateSuccess}
            buttonIcon={'check_circle'}
          />
        )}
        <Link secondary to="/login">
          <Message className={style.goToLogin} content={m.goToLogin} />
        </Link>
      </div>
    )
  }
}
