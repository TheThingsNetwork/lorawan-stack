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
import SubmitButton from '../../../components/submit-button'
import Message from '../../../lib/components/message'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import IntlHelmet from '../../../lib/components/intl-helmet'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import sharedMessages from '../../../lib/shared-messages'
import { id as applicationIdRegexp } from '../../lib/regexp'
import SubmitBar from '../../../components/submit-bar'

import api from '../../api'

import style from './application-add.styl'

const m = defineMessages({
  applicationName: 'Application Name',
  appIdPlaceholder: 'my-new-application',
  appNamePlaceholder: 'My New Application',
  appDescPlaceholder: 'Description for my new application',
  createApplication: 'Create Application',
})

const initialValues = {
  application_id: '',
  name: '',
  description: '',
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
})

@withBreadcrumb('apps.add', function(props) {
  return <Breadcrumb path="/applications/add" icon="add" content={sharedMessages.add} />
})
@connect(
  ({ user }) => ({
    userId: user.user.ids.user_id,
  }),
  dispatch => ({
    navigateToApplication: appId => dispatch(push(`/applications/${appId}`)),
  }),
)
@bind
export default class Add extends React.Component {
  state = {
    error: '',
  }

  async handleSubmit(values, { resetForm }) {
    const { userId, navigateToApplication } = this.props

    await this.setState({ error: '' })

    try {
      const result = await api.application.create(userId, {
        ids: { application_id: values.application_id },
        name: values.name,
        description: values.description,
      })

      const {
        ids: { application_id: appId },
      } = result
      navigateToApplication(appId)
    } catch (error) {
      const { application_id, name, description } = values
      resetForm({ application_id, name, description })

      await this.setState({ error })
    }
  }

  render() {
    const { error } = this.state
    return (
      <Container>
        <Row className={style.wrapper}>
          <Col sm={12}>
            <IntlHelmet title={sharedMessages.addApplication} />
            <Message component="h2" content={sharedMessages.addApplication} />
          </Col>
          <Col className={style.form} sm={12} md={8} lg={8} xl={8}>
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
