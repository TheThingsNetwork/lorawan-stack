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

import { connect } from 'react-redux'
import { replace } from 'connected-react-router'

import withRequest from '@ttn-lw/lib/components/with-request'

import { getWebhookTemplateId } from '@ttn-lw/lib/selectors/id'

import { getWebhook, updateWebhook } from '@console/store/actions/webhooks'
import { attachPromise } from '@console/store/actions/lib'

import {
  selectSelectedWebhook,
  selectWebhookFetching,
  selectWebhookError,
} from '@console/store/selectors/webhooks'
import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import { selectWebhookTemplateById } from '@console/store/selectors/webhook-templates'

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
  'location_solved',
  'template_ids',
]

const mapStateToProps = state => {
  const webhook = selectSelectedWebhook(state)
  const webhookTemplateId = getWebhookTemplateId(webhook)
  const webhookTemplate = Boolean(webhookTemplateId)
    ? selectWebhookTemplateById(state, webhookTemplateId)
    : undefined
  return {
    appId: selectSelectedApplicationId(state),
    webhook,
    webhookTemplate,
    fetching: selectWebhookFetching(state),
    error: selectWebhookError(state),
  }
}

const promisifiedUpdateWebhook = attachPromise(updateWebhook)
const mapDispatchToProps = (dispatch, { match }) => {
  const { appId, webhookId } = match.params
  return {
    getWebhook: () => dispatch(getWebhook(appId, webhookId, webhookEntitySelector)),
    navigateToList: () => dispatch(replace(`/applications/${appId}/integrations/webhooks`)),
    updateWebhook: patch => dispatch(promisifiedUpdateWebhook(appId, webhookId, patch)),
  }
}

export default WebhookEdit =>
  connect(
    mapStateToProps,
    mapDispatchToProps,
  )(
    withRequest(
      ({ getWebhook }) => getWebhook(),
      ({ fetching, webhook }) => fetching || !Boolean(webhook),
    )(WebhookEdit),
  )
