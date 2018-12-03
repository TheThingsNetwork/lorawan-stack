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
import bind from 'autobind-decorator'
import { Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import sharedMessages from '../../../lib/shared-messages'
import Message from '../../../lib/components/message'
import Form from '../../../components/form'
import Field from '../../../components/field'
import Button from '../../../components/button'

import style from './application-general-settings.styl'

const m = defineMessages({
  basics: 'Basics',
  deleteApp: 'Delete Application',
  modalWarning: 'Are you sure you want to delete this Application? This cannot be undone.',
})

@bind
export default class ApplicationGeneralSettings extends React.Component {

  handleSubmit (e) {
  }

  handleDelete () {
  }

  render () {
    return (
      <div>
        <Row justify="center">
          <Col sm={8}>
            <Message
              component="h2"
              content={sharedMessages.generalSettings}
            />
          </Col>
        </Row>
        <Row justify="center">
          <Col sm={8}>
            <Form
              horizontal
              onSubmit={this.handleSubmit}
            >
              <Message
                component="h4"
                content={m.basics}
              />
              <Field
                title={sharedMessages.name}
                required
                name="name"
                type="text"
              />
              <Field
                title={sharedMessages.description}
                name="description"
                type="text"
              />
              <div className={style.submitBar}>
                <div>
                  <Button type="submit" message={sharedMessages.saveChanges} />
                  <Button type="button" naked secondary message={sharedMessages.cancel} />
                </div>
                <Button
                  type="button"
                  icon="delete"
                  danger
                  naked
                  message={m.deleteApp}
                  modalApprove={{ message: m.modalWarning }}
                  onClick={this.handleDelete}
                />
              </div>
            </Form>
          </Col>
        </Row>
      </div>
    )
  }
}
