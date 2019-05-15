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
import { connect } from 'react-redux'
import { Container, Col, Row } from 'react-grid-system'
import * as Yup from 'yup'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import { push } from 'connected-react-router'

import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import sharedMessages from '../../../lib/shared-messages'
import Message from '../../../lib/components/message'
import Form from '../../../components/form'
import Field from '../../../components/field'
import SubmitBar from '../../../components/submit-bar'
import Button from '../../../components/button'
import { id as gatewayIdRegexp, address as addressRegexp } from '../../lib/regexp'
import IntlHelmet from '../../../lib/components/intl-helmet'
import FrequencyPlansSelect from '../../containers/freq-plans-select'
import { withEnv } from '../../../lib/components/env'

import api from '../../api'

import { userIdSelector } from '../../store/selectors/user'

import style from './gateway-add.styl'

const m = defineMessages({
  dutyCycle: 'Enforce Duty Cycle',
  gatewayIdPlaceholder: 'my-new-gateway',
  createGateway: 'Create Gateway',
  gsServerAddressDescription: 'The address of the Gateway Server to connect to',
})

const validationSchema = Yup.object().shape({
  ids: Yup.object().shape({
    gateway_id: Yup.string()
      .matches(gatewayIdRegexp, sharedMessages.validateAlphanum)
      .min(2, sharedMessages.validateTooShort)
      .max(36, sharedMessages.validateTooLong)
      .required(sharedMessages.validateRequired),
    eui: Yup.string()
      .length(8 * 2, sharedMessages.validateTooShort),
  }),
  name: Yup.string()
    .min(2, sharedMessages.validateTooShort)
    .max(50, sharedMessages.validateTooLong),
  description: Yup.string()
    .max(2000, sharedMessages.validateTooLong),
  frequency_plan_id: Yup.string()
    .required(sharedMessages.validateRequired),
  gateway_server_address: Yup.string()
    .matches(addressRegexp, sharedMessages.validateAddressFormat),
})

@withEnv
@withBreadcrumb('gateways.add', function () {
  return (
    <Breadcrumb
      path="/console/gateways/add"
      icon="add"
      content={sharedMessages.add}
    />
  )
})
@connect(function (state, props) {
  const userId = userIdSelector(state, props)

  return { userId }
},
dispatch => ({
  createSuccess: gtwId => dispatch(push(`/console/gateways/${gtwId}`)),
}))
@bind
export default class GatewayAdd extends React.Component {

  state = {
    error: '',
  }

  async handleSubmit (values, { resetForm }) {
    const { userId, createSuccess } = this.props
    const { ids: { gateway_id }} = values

    await this.setState({ error: '' })

    try {
      await api.gateway.create(userId, values)

      createSuccess(gateway_id)
    } catch (error) {
      resetForm(values)

      await this.setState({ error })
    }
  }

  render () {
    const { error } = this.state
    const { env } = this.props

    const gs = env.config.gs

    let gsServerAddress = ''
    if (gs.enabled) {
      gsServerAddress = new URL(gs.base_url).host
    }

    const initialValues = {
      ids: {
        gateway_id: undefined,
      },
      enforce_duty_cycle: true,
      gateway_server_address: gsServerAddress,
      frequency_plan_id: undefined,
    }

    return (
      <Container>
        <Row className={style.wrapper}>
          <Col sm={12}>
            <IntlHelmet
              title={sharedMessages.addGateway}
            />
            <Message component="h2" content={sharedMessages.addGateway} />
          </Col>
          <Col sm={12} md={8}>
            <Form
              submitEnabledWhenInvalid
              error={error}
              onSubmit={this.handleSubmit}
              initialValues={initialValues}
              validationSchema={validationSchema}
              horizontal
            >
              <Message
                component="h4"
                content={sharedMessages.generalSettings}
              />
              <Field
                title={sharedMessages.gatewayID}
                name="ids.gateway_id"
                type="text"
                placeholder={m.gatewayIdPlaceholder}
                required
                autoFocus
              />
              <Field
                title={sharedMessages.gatewayEUI}
                name="ids.eui"
                type="byte"
                min={8}
                max={8}
                placeholder={sharedMessages.gatewayEUI}
              />
              <Field
                title={sharedMessages.gatewayName}
                name="name"
              />
              <Field
                title={sharedMessages.gatewayDescription}
                name="description"
                type="textarea"
              />
              <Field
                title={sharedMessages.gatewayServerAddress}
                description={m.gsServerAddressDescription}
                placeholder={sharedMessages.addressPlaceholder}
                name="gateway_server_address"
                type="text"
              />
              <Message
                component="h4"
                content={sharedMessages.lorawanOptions}
              />
              <FrequencyPlansSelect
                horizontal
                source="gs"
                name="frequency_plan_id"
                menuPlacement="top"
                required
              />
              <Field
                title={m.dutyCycle}
                name="enforce_duty_cycle"
                type="checkbox"
              />
              <SubmitBar>
                <Button type="submit" message={m.createGateway} />
              </SubmitBar>
            </Form>
          </Col>
        </Row>
      </Container>
    )
  }
}
