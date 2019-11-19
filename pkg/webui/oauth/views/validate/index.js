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
import PropTypes from '../../../lib/prop-types'
import Spinner from '../../../components/spinner'
import Message from '../../../lib/components/message'
import ErrorNotification from '../../../components/error-notification'
import Notification from '../../../components/notification'
import api from '../../api'
import sharedMessages from '../../../lib/shared-messages'

const m = defineMessages({
  validateSuccess: 'Contact info validated successfully',
  validateFail: 'Contact info validation failed',
  goToLogin: 'Go to Login',
})

@connect(
  null,
  {
    goToLogin: () => push('/login'),
  },
)
export default class Validate extends React.PureComponent {
  static propTypes = {
    goToLogin: PropTypes.func.isRequired,
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
      this.handleError(error)
    }
  }

  render() {
    const { fetching, error, success } = this.state
    const { goToLogin } = this.props

    if (fetching) {
      return (
        <Spinner center>
          <Message content={sharedMessages.fetching} />
        </Spinner>
      )
    }

    return (
      <Container>
        <Row justify="center" align="center">
          <Col sm={12} md={8} lg={5}>
            {error && (
              <ErrorNotification
                content={error}
                title={m.validateFail}
                action={goToLogin}
                actionMessage={m.goToLogin}
                buttonIcon={'error'}
              />
            )}
            {success && (
              <Notification
                large
                success
                content={success}
                title={m.validateSuccess}
                action={goToLogin}
                actionMessage={m.goToLogin}
                buttonIcon={'check_circle'}
              />
            )}
          </Col>
        </Row>
      </Container>
    )
  }
}
