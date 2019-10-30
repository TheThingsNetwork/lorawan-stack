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
import { defineMessages } from 'react-intl'
import { push } from 'connected-react-router'
import queryString from 'query-string'
import { connect } from 'react-redux'
import IntlHelmet from '../../../lib/components/intl-helmet'
import PropTypes from '../../../lib/prop-types'

import Spinner from '../../../components/spinner'
import Message from '../../../lib/components/message'
import Notification from '../../../components/notification'
import api from '../../api'
import sharedMessages from '../../../lib/shared-messages'
import style from '../create-account/create-account.styl'

const m = defineMessages({
  emailConfirmed: 'Email confirmed successfully',
  confirmEmail: 'Confirm your email',
  confirmFail: 'Confirmation failed',
  goToLogin: 'Go to Login',
})

@connect(
  null,
  {
    goToLogin: () =>
      push('/login', {
        info: m.passwordChanged,
      }),
  },
)
export default class ConfirmEmail extends React.PureComponent {
  static propTypes = {
    goToLogin: PropTypes.func.isRequired,
    location: PropTypes.location.isRequired,
  }

  state = {
    error: undefined,
    success: undefined,
    fetching: true,
    confirmed: false,
  }

  handleError(error) {
    this.setState({ error: error.response, fetching: false, success: undefined })
  }

  handleSuccess() {
    this.setState({ success: m.emailConfirmed, fetching: false, confirmed: true, error: undefined })
  }

  async componentDidMount() {
    const validationData = queryString.parse(this.props.location.search)
    try {
      await api.users.validateEmail({
        token: validationData.token,
        id: validationData.reference,
      })
      this.handleSuccess()
    } catch (error) {
      this.handleError(error)
    }
  }

  render() {
    const { fetching, error, success } = this.state
    const { goToLogin } = this.props

    if (fetching) {
      return (
        <Spinner center>
          <Message content={sharedMessages.loading} />
        </Spinner>
      )
    }

    return (
      <Container className={style.fullHeight}>
        <Row justify="center" align="center" className={style.fullHeight}>
          <Col sm={12} md={8} lg={5}>
            <IntlHelmet title={m.confirmEmail} />
            {error && (
              <Notification
                large
                error={error}
                title={m.confirmFail}
                action={goToLogin}
                actionMessage={m.goToLogin}
                buttonIcon={'add_circle'}
              />
            )}
            {success && (
              <Notification
                large
                success={success}
                title={m.emailConfirmed}
                action={goToLogin}
                actionMessage={m.goToLogin}
                buttonIcon={'add_circle'}
              />
            )}
          </Col>
        </Row>
      </Container>
    )
  }
}
