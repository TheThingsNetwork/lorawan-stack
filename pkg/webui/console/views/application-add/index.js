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
import { replace } from 'connected-react-router'
import { connect } from 'react-redux'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'

import Form from '../../../components/form'
import Field from '../../../components/field'
import Button from '../../../components/button'

import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import sharedMessages from '../../../lib/shared-messages'

import { createApplication } from '../../store/actions/applications'


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
    .matches(/^[0-9a-z]+$/i, sharedMessages.validateAlphanum)
    .min(2, sharedMessages.validateTooShort)
    .max(25, sharedMessages.validateTooLong)
    .required(sharedMessages.validateRequired),
  name: Yup.string()
    .min(2, sharedMessages.validateTooShort)
    .max(50, sharedMessages.validateTooLong)
    .required(sharedMessages.validateRequired),
  description: Yup.string(),
})

@withBreadcrumb('apps.add', function (props) {
  return (
    <Breadcrumb
      path="/console/applications/add"
      icon="add"
      content={sharedMessages.add}
    />
  )
})
@connect(({ application }) => ({ error: application.error }))
@bind
export default class Add extends React.Component {

  handleSubmit (values, { setSubmitting }) {
    const { dispatch } = this.props

    dispatch(createApplication(values))
    setSubmitting(false)
  }

  handleCancel () {
    const { dispatch } = this.props

    dispatch(replace('/console/applications'))
  }

  render () {
    const { error } = this.props

    return (
      <Container>
        <Row className={style.wrapper}>
          <Col sm={12}><h2>Add Application</h2></Col>
          <Col className={style.form} sm={12} md={8} lg={8} xl={8}>
            <Form
              error={error}
              onSubmit={this.handleSubmit}
              initialValues={initialValues}
              validationSchema={validationSchema}
              horizontal
            >
              <Field
                title={sharedMessages.appId}
                name="application_id"
                type="text"
                placeholder={m.appIdPlaceholder}
                autoFocus
                required
              />
              <Field
                title={m.applicationName}
                name="name"
                type="text"
                placeholder={m.appNamePlaceholder}
                required
              />
              <Field
                title={sharedMessages.description}
                name="description"
                placeholder={m.appDescPlaceholder}
                type="text"
              />
              <div className={style.submitBar}>
                <Button type="submit" message={m.createApplication} />
                <Button type="button" naked message={sharedMessages.cancel} onClick={this.handleCancel} />
              </div>
            </Form>
          </Col>
          <Col className={style.description} sm={12} md={4} xl={3}>
            <aside>
              <p>
                Here is a text that sort of explains the process of adding an application.
                This is to help users making sense of what is actually happening.
                <br />
                We could also provide links and resources from our documentation.
                Lorem Ipsum dolor sit amet.
              </p>
            </aside>
          </Col>
        </Row>
      </Container>
    )
  }
}
