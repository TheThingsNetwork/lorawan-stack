// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import { connect } from 'react-redux'

import withRequest from '@ttn-lw/lib/components/with-request'

import { getWebhookTemplateId } from '@ttn-lw/lib/selectors/id'

import { getWebhook } from '@console/store/actions/webhooks'

import {
  selectSelectedWebhook,
  selectWebhookFetching,
  selectWebhookError,
} from '@console/store/selectors/webhooks'
import { selectWebhookTemplateById } from '@console/store/selectors/webhook-templates'
import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import { selectWebhooksHealthStatusEnabled } from '@console/store/selectors/application-server'

const webhookEntitySelector = [
  'base_url',
  'format',
  'headers',
  'downlink_api_key',
  'uplink_message',
  'join_accept',
  'downlink_ack',
  'downlink_nack',
  'downlink_sent',
  'downlink_failed',
  'downlink_queued',
  'downlink_queue_invalidated',
  'location_solved',
  'service_data',
  'template_ids',
  'health_status',
]

const mapStateToProps = state => {
  const healthStatusEnabled = selectWebhooksHealthStatusEnabled(state)
  const webhook = selectSelectedWebhook(state)
  const webhookTemplateId = getWebhookTemplateId(webhook)
  const webhookTemplate = Boolean(webhookTemplateId)
    ? selectWebhookTemplateById(state, webhookTemplateId)
    : undefined
  return {
    appId: selectSelectedApplicationId(state),
    webhook,
    webhookTemplate,
    healthStatusEnabled,
    fetching: selectWebhookFetching(state),
    error: selectWebhookError(state),
  }
}

const mapDispatchToProps = (dispatch, { match }) => {
  const { appId, webhookId } = match.params
  return {
    getWebhook: () => dispatch(getWebhook(appId, webhookId, webhookEntitySelector)),
  }
}

export default WebhookEdit =>
  connect(
    mapStateToProps,
    mapDispatchToProps,
  )(withRequest(({ getWebhook }) => getWebhook())(WebhookEdit))
