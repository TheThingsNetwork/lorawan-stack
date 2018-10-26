// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
import { Col, Row } from 'react-grid-system'

import Form from '../../../components/form'
import Field from '../../../components/field'
import Button from '../../../components/button'

import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import sharedMessages from '../../../lib/shared-messages'

import style from './application-add.styl'

const initialValues = {
  app_id: '',
  name: '',
  description: '',
}

@withBreadcrumb('apps.add', function (props) {
  return (
    <Breadcrumb
      path="/console/applications/add"
      icon="add"
      content={sharedMessages.add}
    />
  )
})
export default class Add extends React.Component {

  handleSubmit (e) {
    console.log(e)
  }

  render () {
    return (
      <Row>
        <Col className={style.description} sm={12} md={4}>
          <h2>Add Application</h2>
          <p>
            Here is a text that sort of explains the process of adding an application.
            This is to help users making sense of what is actually happening.
            <br />
            We could also provide links and resources from our documentation.
            Lorem Ipsum dolor sit amet.
          </p>
        </Col>
        <Col className={style.form} sm={12} md={8}>
          <Form
            onSubmit={this.handleSubmit}
            initialValues={initialValues}
          >
            <Field
              title="Application ID"
              name="app_id"
              type="text"
              required
            />
            <Field
              title="Application Name"
              name="name"
              type="text"
              required
            />
            <Field
              title="Description"
              name="description"
              type="text"
            />
            <Button type="submit" message="Create Application" />
            <Button naked message="Cancel" />
          </Form>
        </Col>
      </Row>
    )
  }
}
