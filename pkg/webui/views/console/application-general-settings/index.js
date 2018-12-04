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

import { withBreadcrumb } from '../../../components/breadcrumbs/context'

import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import sharedMessages from '../../../lib/shared-messages'
import Message from '../../../lib/components/message'
import Form from '../../../components/form'
import Field from '../../../components/field'
import Button from '../../../components/button'
import ModalButton from '../../../containers/modal-button'

import style from './application-general-settings.styl'

const m = defineMessages({
  basics: 'Basics',
  deleteApp: 'Delete application',
  modalWarning: 'Are you sure you want to delete "{appId}"? Deleting an application cannot be undone!',
  generalSettings: 'General Settings',
})

@withBreadcrumb('apps.single.general-settings', function (props) {
  const { match } = props
  const appId = match.params.appId

  return (
    <Breadcrumb
      path={`/console/applications/${appId}/general-settings`}
      icon="general_settings"
      content={m.generalSettings}
    />
  )
})
@bind
export default class ApplicationGeneralSettings extends React.Component {

  handleSubmit (e) {
  }

  handleDelete () {
  }

  render () {
    const { match } = this.props
    const appId = match.params.appId

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
                <ModalButton
                  type="button"
                  icon="delete"
                  danger
                  naked
                  message={m.deleteApp}
                  modalData={{ message: { values: { appId }, ...m.modalWarning }}}
                  onApprove={this.handleDelete}
                />
              </div>
            </Form>
          </Col>
        </Row>
      </div>
    )
  }
}
