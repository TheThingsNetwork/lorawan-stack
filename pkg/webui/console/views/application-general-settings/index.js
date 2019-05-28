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
import bind from 'autobind-decorator'
import { Col, Row, Container } from 'react-grid-system'
import { defineMessages } from 'react-intl'
import { connect } from 'react-redux'
import * as Yup from 'yup'
import { replace } from 'connected-react-router'

import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import sharedMessages from '../../../lib/shared-messages'
import Message from '../../../lib/components/message'
import Form from '../../../components/form'
import Field from '../../../components/field'
import Button from '../../../components/button'
import ModalButton from '../../../components/button/modal-button'
import diff from '../../../lib/diff'
import toast from '../../../components/toast'
import SubmitBar from '../../../components/submit-bar'

import { selectSelectedApplication } from '../../store/selectors/application'

import api from '../../api'

const m = defineMessages({
  basics: 'Basics',
  deleteApp: 'Delete application',
  modalWarning: 'Are you sure you want to delete "{appName}"? Deleting an application cannot be undone!',
  generalSettings: 'General Settings',
  updateSuccess: 'Successfully updated application',
})

const validationSchema = Yup.object().shape({
  name: Yup.string()
    .min(3, sharedMessages.validateTooShort)
    .max(50, sharedMessages.validateTooLong),
  description: Yup.string()
    .max(150, sharedMessages.validateTooLong),
})

@connect(function (state) {
  return {
    application: selectSelectedApplication(state),
  }
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

  state = {
    error: '',
  }

  async handleSubmit (values, { resetForm }) {
    const { application } = this.props

    await this.setState({ error: '' })

    const changed = diff(application, values)
    try {
      const { ids: { application_id }} = application
      await api.application.update(application_id, changed)
      resetForm({ ...values })
      toast({
        title: application_id,
        message: m.updateSuccess,
        type: toast.types.SUCCESS,
      })
    } catch (error) {
      resetForm({ ...values })
      await this.setState(error)
    }
  }

  async handleDelete () {
    const { dispatch } = this.props
    const { appId } = this.props.match.params

    await this.setState({ error: '' })

    try {
      await api.application.delete(appId)
      dispatch(replace('/console/applications'))
    } catch (error) {
      this.setState(error)
    }
  }

  render () {
    const { application } = this.props
    const { error } = this.state

    return (
      <Container>
        <Row>
          <Col lg={8} md={12}>
            <Message
              component="h2"
              content={sharedMessages.generalSettings}
            />
          </Col>
        </Row>
        <Row>
          <Col lg={8} md={12}>
            <Form
              error={error}
              horizontal
              onSubmit={this.handleSubmit}
              initialValues={{ name: application.name, description: application.description }}
              validationSchema={validationSchema}
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
              <SubmitBar>
                <Button type="submit" message={sharedMessages.saveChanges} />
                <ModalButton
                  type="button"
                  icon="delete"
                  danger
                  naked
                  message={m.deleteApp}
                  modalData={{ message: { values: { appName: application.name }, ...m.modalWarning }}}
                  onApprove={this.handleDelete}
                />
              </SubmitBar>
            </Form>
          </Col>
        </Row>
      </Container>
    )
  }
}
