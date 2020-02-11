// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import React, { Component } from 'react'
import { Container, Col, Row } from 'react-grid-system'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'

import PageTitle from '../../../components/page-title'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import WebhookForm from '../../components/webhook-form'
import toast from '../../../components/toast'
import diff from '../../../lib/diff'
import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'

import api from '../../api'

const m = defineMessages({
  editWebhook: 'Edit Webhook',
  updateSuccess: 'Successfully updated webhook',
  deleteSuccess: 'Successfully deleted webhook',
})

@withBreadcrumb('apps.single.integrations.edit', function(props) {
  const {
    appId,
    match: {
      params: { webhookId },
    },
  } = props
  return (
    <Breadcrumb
      path={`/applications/${appId}/integrations/${webhookId}`}
      content={sharedMessages.edit}
    />
  )
})
export default class ApplicationWebhookEdit extends Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    match: PropTypes.match.isRequired,
    navigateToList: PropTypes.func.isRequired,
    updateWebhook: PropTypes.func.isRequired,
    webhook: PropTypes.webhook.isRequired,
  }

  @bind
  async handleSubmit(updatedWebhook) {
    const { webhook: originalWebhook, updateWebhook } = this.props

    const patch = diff(originalWebhook, updatedWebhook, ['ids'])

    // Ensure that the header prop is always patched fully, otherwise we loose
    // old header entries.
    if ('headers' in patch) {
      patch.headers = updatedWebhook.headers
    }

    await updateWebhook(patch)
  }

  @bind
  handleSubmitSuccess() {
    toast({
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  @bind
  async handleDelete() {
    const {
      appId,
      match: {
        params: { webhookId },
      },
    } = this.props

    await api.application.webhooks.delete(appId, webhookId)
  }

  @bind
  async handleDeleteSuccess() {
    const { navigateToList } = this.props

    toast({
      message: m.deleteSuccess,
      type: toast.types.SUCCESS,
    })

    navigateToList()
  }

  render() {
    const { webhook, appId } = this.props

    return (
      <Container>
        <PageTitle title={m.editWebhook} />
        <Row>
          <Col lg={8} md={12}>
            <WebhookForm
              update
              appId={appId}
              initialWebhookValue={webhook}
              onSubmit={this.handleSubmit}
              onSubmitSuccess={this.handleSubmitSuccess}
              onDelete={this.handleDelete}
              onDeleteSuccess={this.handleDeleteSuccess}
            />
          </Col>
        </Row>
      </Container>
    )
  }
}
