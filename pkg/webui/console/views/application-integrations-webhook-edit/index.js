// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
import { useSelector } from 'react-redux'
import { useParams } from 'react-router-dom'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import WebhookEdit from '@console/containers/webhook-edit'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { getWebhookTemplateId } from '@ttn-lw/lib/selectors/id'

import { getWebhook } from '@console/store/actions/webhooks'

import { selectSelectedWebhook } from '@console/store/selectors/webhooks'
import { selectWebhookTemplateById } from '@console/store/selectors/webhook-templates'
import {
  selectWebhooksHealthStatusEnabled,
  selectWebhookRetryInterval,
  selectWebhookHasUnhealthyConfig,
} from '@console/store/selectors/application-server'

const ApplicationWebhookEditInner = () => {
  const { appId, webhookId } = useParams()
  const healthStatusEnabled = useSelector(selectWebhooksHealthStatusEnabled)
  const webhookRetryInterval = useSelector(selectWebhookRetryInterval)
  const hasUnhealthyWebhookConfig = useSelector(selectWebhookHasUnhealthyConfig)
  const webhook = useSelector(selectSelectedWebhook)
  const webhookTemplateId = getWebhookTemplateId(webhook)
  const webhookTemplate = useSelector(state => selectWebhookTemplateById(state, webhookTemplateId))
  useBreadcrumbs(
    'apps.single.integrations.webhooks.edit',
    <Breadcrumb
      path={`/applications/${appId}/integrations/${webhookId}`}
      content={sharedMessages.edit}
    />,
  )

  return (
    <div className="container container--xl grid">
      <PageTitle
        title={sharedMessages.editWebhook}
        className="mb-0"
        hideHeading={Boolean(webhookTemplate)}
      />
      <div className="item-12">
        <WebhookEdit
          update
          appId={appId}
          selectedWebhook={webhook}
          webhookId={webhookId}
          webhookTemplate={webhookTemplate}
          healthStatusEnabled={healthStatusEnabled}
          webhookRetryInterval={webhookRetryInterval}
          hasUnhealthyWebhookConfig={hasUnhealthyWebhookConfig}
        />
      </div>
    </div>
  )
}

const webhookEntitySelector = [
  'base_url',
  'format',
  'headers',
  'downlink_api_key',
  'uplink_message',
  'uplink_normalized',
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
  'field_mask',
  'paused',
]

const ApplicationWebhookEdit = () => {
  const { appId, webhookId } = useParams()

  return (
    <RequireRequest requestAction={getWebhook(appId, webhookId, webhookEntitySelector)}>
      <ApplicationWebhookEditInner />
    </RequireRequest>
  )
}

export default ApplicationWebhookEdit
