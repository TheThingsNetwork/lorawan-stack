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
import { defineMessages } from 'react-intl'
import { connect } from 'react-redux'
import { Container, Col, Row } from 'react-grid-system'
import bind from 'autobind-decorator'
import { push } from 'connected-react-router'

import FormSubmit from '../../../components/form/submit'
import SubmitButton from '../../../components/submit-button'
import GatewayDataForm from '../../components/gateway-data-form'

import sharedMessages from '../../../lib/shared-messages'
import Message from '../../../lib/components/message'
import PropTypes from '../../../lib/prop-types'
import IntlHelmet from '../../../lib/components/intl-helmet'
import { withEnv } from '../../../lib/components/env'

import api from '../../api'

import { selectUserId } from '../../store/selectors/user'

import style from './gateway-add.styl'

const m = defineMessages({
  createGateway: 'Create Gateway',
})

@withEnv
@connect(
  function(state) {
    const userId = selectUserId(state)

    return { userId }
  },
  dispatch => ({
    createSuccess: gtwId => dispatch(push(`/gateways/${gtwId}`)),
  }),
)
export default class GatewayAdd extends React.Component {
  static propTypes = {
    createSuccess: PropTypes.func.isRequired,
    env: PropTypes.env.isRequired,
    userId: PropTypes.string.isRequired,
  }

  state = {
    error: '',
  }

  @bind
  async handleSubmit(values, { resetForm }) {
    const { userId, createSuccess } = this.props
    const { owner_id, ...gateway } = values
    const {
      ids: { gateway_id },
    } = gateway

    await this.setState({ error: '' })

    try {
      await api.gateway.create(owner_id, gateway, userId === owner_id)

      createSuccess(gateway_id)
    } catch (error) {
      resetForm(values)

      await this.setState({ error })
    }
  }

  render() {
    const { error } = this.state
    const {
      env: {
        config: { stack },
      },
      userId,
    } = this.props

    const initialValues = {
      ids: {
        gateway_id: undefined,
      },
      enforce_duty_cycle: true,
      gateway_server_address: stack.gs.enabled ? new URL(stack.gs.base_url).hostname : '',
      frequency_plan_id: undefined,
      owner_id: userId,
    }

    return (
      <Container>
        <Row>
          <Col>
            <IntlHelmet title={sharedMessages.addGateway} />
            <Message className={style.title} component="h2" content={sharedMessages.addGateway} />
          </Col>
        </Row>
        <Row>
          <Col md={10} lg={9}>
            <GatewayDataForm
              error={error}
              onSubmit={this.handleSubmit}
              initialValues={initialValues}
              update={false}
            >
              <FormSubmit component={SubmitButton} message={m.createGateway} />
            </GatewayDataForm>
          </Col>
        </Row>
      </Container>
    )
  }
}
