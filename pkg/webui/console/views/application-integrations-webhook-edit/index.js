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

import React, { Component } from 'react'
import { Container, Col, Row } from 'react-grid-system'
import bind from 'autobind-decorator'
import { connect } from 'react-redux'
import { defineMessages } from 'react-intl'
import { replace } from 'connected-react-router'

import PageTitle from '../../../components/page-title'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import WebhookForm from '../../components/webhook-form'
import toast from '../../../components/toast'
import diff from '../../../lib/diff'
import sharedMessages from '../../../lib/shared-messages'
import withRequest from '../../../lib/components/with-request'

import {
  selectSelectedWebhook,
  selectWebhookFetching,
  selectWebhookError,
} from '../../store/selectors/webhooks'
import { selectSelectedApplicationId } from '../../store/selectors/applications'
import { getWebhook } from '../../store/actions/webhooks'

import api from '../../api'
import PropTypes from '../../../lib/prop-types'

const m = defineMessages({
  editWebhook: 'Edit Webhook',
  updateSuccess: 'Successfully updated webhook',
  deleteSuccess: 'Successfully deleted webhook',
})

const webhookEntitySelector = [
  'base_url',
  'format',
  'headers',
  'uplink_message',
  'join_accept',
  'downlink_ack',
  'downlink_nack',
  'downlink_sent',
  'downlink_failed',
  'downlink_queued',
  'location_solved',
]

@connect(
  state => ({
    appId: selectSelectedApplicationId(state),
    webhook: selectSelectedWebhook(state),
    fetching: selectWebhookFetching(state),
    error: selectWebhookError(state),
  }),
  function(dispatch, { match }) {
    const { appId, webhookId } = match.params
    return {
      getWebhook: () => dispatch(getWebhook(appId, webhookId, webhookEntitySelector)),
      navigateToList: () => dispatch(replace(`/applications/${appId}/integrations/webhooks`)),
    }
  },
)
@withRequest(
  ({ getWebhook }) => getWebhook(),
  ({ fetching, webhook }) => fetching || !Boolean(webhook),
)
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
      icon="general_settings"
      content={sharedMessages.edit}
    />
  )
})
@bind
export default class ApplicationWebhookEdit extends Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    match: PropTypes.match.isRequired,
    navigateToList: PropTypes.func.isRequired,
    webhook: PropTypes.webhook.isRequired,
  }

  async handleSubmit(updatedWebhook) {
    const {
      appId,
      match: {
        params: { webhookId },
      },
      webhook: originalWebhook,
    } = this.props
    const patch = diff(originalWebhook, updatedWebhook, ['ids'])

    // Ensure that the header prop is always patched fully, otherwise we loose
    // old header entries.
    if ('headers' in patch) {
      patch.headers = updatedWebhook.headers
    }

    await api.application.webhooks.update(appId, webhookId, patch)
  }

  handleSubmitSuccess() {
    toast({
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  async handleDelete() {
    const {
      appId,
      match: {
        params: { webhookId },
      },
    } = this.props

    await api.application.webhooks.delete(appId, webhookId)
  }

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
