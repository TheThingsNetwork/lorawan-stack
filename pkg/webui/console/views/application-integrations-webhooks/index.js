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

import React from 'react'
import { Routes, Route, useParams } from 'react-router-dom'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import ErrorView from '@ttn-lw/lib/components/error-view'
import ValidateRouteParam from '@ttn-lw/lib/components/validate-route-param'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import Require from '@console/lib/components/require'

import ApplicationWebhookChoose from '@console/views/application-integrations-webhook-add-choose'
import ApplicationWebhookEdit from '@console/views/application-integrations-webhook-edit'
import ApplicationWebhookAdd from '@console/views/application-integrations-webhook-add'
import ApplicationWebhooksList from '@console/views/application-integrations-webhooks-list'
import SubViewError from '@console/views/sub-view-error'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { pathId as pathIdRegexp } from '@ttn-lw/lib/regexp'

import { mayViewApplicationEvents } from '@console/lib/feature-checks'

import { listWebhookTemplates } from '@console/store/actions/webhook-templates'

const ApplicationWebhooksInner = () => {
  const { appId } = useParams()

  useBreadcrumbs(
    'apps.single.integrations.webhooks',
    <Breadcrumb
      path={`/applications/${appId}/integrations/webhooks`}
      content={sharedMessages.webhooks}
    />,
  )

  return (
    <ErrorView errorRender={SubViewError}>
      <Routes>
        <Route index Component={ApplicationWebhooksList} />
        <Route path="add" Component={ApplicationWebhookAdd} />
        <Route
          exact
          path=":webhookId"
          component={
            <ValidateRouteParam
              check={{ webhookId: pathIdRegexp }}
              Component={ApplicationWebhookEdit}
            />
          }
        />
        <Route path="add/template/*" Component={ApplicationWebhookChoose} />
      </Routes>
    </ErrorView>
  )
}

const selector = [
  'base_url',
  'create_downlink_api_key',
  'description',
  'documentation_url',
  'downlink_ack',
  'downlink_failed',
  'downlink_nack',
  'downlink_queue_invalidated',
  'downlink_queued',
  'downlink_sent',
  'fields',
  'format',
  'headers',
  'ids',
  'info_url',
  'join_accept',
  'location_solved',
  'logo_url',
  'name',
  'service_data',
  'uplink_message',
  'uplink_normalized',
  'field_mask',
]

const ApplicationWebhooks = () => {
  const { appId } = useParams()

  return (
    <Require
      featureCheck={mayViewApplicationEvents}
      otherwise={{ redirect: `/applications/${appId}` }}
    >
      <RequireRequest requestAction={listWebhookTemplates(selector)}>
        <ApplicationWebhooksInner />
      </RequireRequest>
    </Require>
  )
}

export default ApplicationWebhooks
