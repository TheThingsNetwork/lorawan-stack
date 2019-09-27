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
import * as Yup from 'yup'
import { connect } from 'react-redux'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import { push } from 'connected-react-router'

import Form from '../../../components/form'
import Input from '../../../components/input'
import Checkbox from '../../../components/checkbox'
import SubmitButton from '../../../components/submit-button'
import toast from '../../../components/toast'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import PropTypes from '../../../lib/prop-types'
import sharedMessages from '../../../lib/shared-messages'
import { id as applicationIdRegexp, address } from '../../lib/regexp'
import SubmitBar from '../../../components/submit-bar'
import { getApplicationId } from '../../../lib/selectors/id'
import { selectAsConfig } from '../../../lib/selectors/env'

import api from '../../api'

import style from './application-add.styl'

const m = defineMessages({
  applicationName: 'Application Name',
  appIdPlaceholder: 'my-new-application',
  appNamePlaceholder: 'My New Application',
  appDescPlaceholder: 'Description for my new application',
  createApplication: 'Create Application',
  linkAutomatically: 'Link automatically',
  linkFailure: 'There was a problem while linking the application',
  linkFailureTitle: 'Application link failed',
})

const initialValues = {
  application_id: '',
  name: '',
  description: '',
  link: true,
  network_server_address: '',
}

const validationSchema = Yup.object().shape({
  application_id: Yup.string()
    .matches(applicationIdRegexp, sharedMessages.validateAlphanum)
    .min(2, sharedMessages.validateTooShort)
    .max(25, sharedMessages.validateTooLong)
    .required(sharedMessages.validateRequired),
  name: Yup.string()
    .min(2, sharedMessages.validateTooShort)
    .max(50, sharedMessages.validateTooLong),
  description: Yup.string(),
  network_server_address: Yup.string().when('link', {
    is: true,
    then: Yup.string().matches(address, sharedMessages.validateFormat),
  }),
})

@connect(
  ({ user }) => ({
    userId: user.user.ids.user_id,
    asEnabled: selectAsConfig().enabled,
  }),
  dispatch => ({
    navigateToApplication: appId => dispatch(push(`/applications/${appId}`)),
  }),
)
@bind
export default class Add extends React.Component {
  static propTypes = {
    asEnabled: PropTypes.bool.isRequired,
    navigateToApplication: PropTypes.func.isRequired,
    userId: PropTypes.string.isRequired,
  }
  state = {
    error: '',
    link: true,
  }

  async handleSubmit(values, { resetForm }) {
    const { userId, navigateToApplication, asEnabled } = this.props

    await this.setState({ error: '' })

    try {
      const result = await api.application.create(userId, {
        ids: { application_id: values.application_id },
        name: values.name,
        description: values.description,
      })

      const appId = getApplicationId(result)

      if (asEnabled && values.link) {
        try {
          const key = {
            name: 'Application Server Linking',
            rights: ['RIGHT_APPLICATION_LINK'],
          }
          const { key: api_key } = await api.application.apiKeys.create(appId, key)
          await api.application.link.set(appId, {
            api_key,
            network_server_address: values.network_server_address,
          })
        } catch (err) {
          toast({
            title: m.linkFailureTitle,
            message: m.linkFailure,
            type: toast.types.ERROR,
          })
        }
      }

      navigateToApplication(appId)
    } catch (error) {
      const { application_id, name, description } = values
      resetForm({ application_id, name, description })

      await this.setState({ error })
    }
  }

  handleLinkChange(event) {
    this.setState({
      link: event.target.checked,
    })
  }

  get linkingBit() {
    const { link } = this.state

    return (
      <React.Fragment>
        <Form.Field
          onChange={this.handleLinkChange}
          title={m.linkAutomatically}
          name="link"
          component={Checkbox}
        />
        <Form.Field
          component={Input}
          description={sharedMessages.nsEmptyDefault}
          name="network_server_address"
          title={sharedMessages.nsAddress}
          disabled={!link}
        />
      </React.Fragment>
    )
  }

  render() {
    const { error } = this.state
    const { asEnabled } = this.props

    return (
      <Container>
        <Row className={style.wrapper}>
          <Col sm={12}>
            <IntlHelmet title={sharedMessages.addApplication} />
            <Message component="h2" content={sharedMessages.addApplication} />
          </Col>
          <Col className={style.form} md={10} lg={9}>
            <Form
              error={error}
              onSubmit={this.handleSubmit}
              initialValues={initialValues}
              validationSchema={validationSchema}
            >
              <Form.Field
                title={sharedMessages.appId}
                name="application_id"
                placeholder={m.appIdPlaceholder}
                autoFocus
                required
                component={Input}
              />
              <Form.Field
                title={m.applicationName}
                name="name"
                placeholder={m.appNamePlaceholder}
                component={Input}
              />
              <Form.Field
                title={sharedMessages.description}
                name="description"
                placeholder={m.appDescPlaceholder}
                component={Input}
              />
              {asEnabled && this.linkingBit}
              <SubmitBar>
                <Form.Submit message={m.createApplication} component={SubmitButton} />
              </SubmitBar>
            </Form>
          </Col>
        </Row>
      </Container>
    )
  }
}
