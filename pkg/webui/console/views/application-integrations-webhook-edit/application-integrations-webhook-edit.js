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
import { Container, Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import WebhookEdit from '@console/containers/webhook-edit'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const m = defineMessages({
  editWebhook: 'Edit webhook',
  updateSuccess: 'Webhook updated',
  deleteSuccess: 'Webhook deleted',
  reactivateSuccess: 'Webhook activated',
})

const ApplicationWebhookEdit = props => {
  const {
    webhook,
    appId,
    webhookTemplate,
    match,
    healthStatusEnabled,
    webhookRetryInterval,
    hasUnhealthyWebhookConfig,
  } = props
  const { webhookId } = match.params
  useBreadcrumbs(
    'apps.single.integrations.webhooks.edit',
    <Breadcrumb
      path={`/applications/${appId}/integrations/${webhookId}`}
      content={sharedMessages.edit}
    />,
  )

  return (
    <Container>
      <PageTitle title={m.editWebhook} className="mb-0" hideHeading={Boolean(webhookTemplate)} />
      <Row>
        <Col lg={8} md={12}>
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
        </Col>
      </Row>
    </Container>
  )
}

ApplicationWebhookEdit.propTypes = {
  appId: PropTypes.string.isRequired,
  hasUnhealthyWebhookConfig: PropTypes.bool.isRequired,
  healthStatusEnabled: PropTypes.bool.isRequired,
  match: PropTypes.match.isRequired,
  webhook: PropTypes.webhook.isRequired,
  webhookRetryInterval: PropTypes.string,
  webhookTemplate: PropTypes.webhookTemplate,
}

ApplicationWebhookEdit.defaultProps = {
  webhookTemplate: undefined,
  webhookRetryInterval: null,
}

export default ApplicationWebhookEdit
