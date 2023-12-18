// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback } from 'react'
import { Container, Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'
import { useNavigate, useParams } from 'react-router-dom'
import { useSelector } from 'react-redux'

import PageTitle from '@ttn-lw/components/page-title'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'

import WebhookAdd from '@console/containers/webhook-add'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { selectWebhookTemplateById } from '@console/store/selectors/webhook-templates'

const m = defineMessages({
  addCustomWebhook: 'Add custom webhook',
  addWebhookViaTemplate: 'Add webhook for {templateName}',
  customWebhook: 'Custom webhook',
})

const ApplicationWebhookAddForm = () => {
  const { appId, templateId } = useParams()
  const isCustom = !Boolean(templateId) || templateId === 'custom'
  const webhookTemplate = useSelector(state => selectWebhookTemplateById(state, templateId))
  const navigate = useNavigate()

  let breadcrumbContent = m.customWebhook
  if (!templateId) {
    breadcrumbContent = sharedMessages.add
  } else if (!isCustom && webhookTemplate.name) {
    breadcrumbContent = webhookTemplate.name
  }

  useBreadcrumbs('apps.single.integrations.webhooks.various.add', [
    {
      path: `/applications/${appId}/integrations/webhooks/add/template/${templateId}`,
      content: breadcrumbContent,
    },
  ])

  const navigateToList = useCallback(
    () => navigate(`/applications/${appId}/integrations/webhooks`),
    [appId, navigate],
  )

  let pageTitle = m.addCustomWebhook
  if (!webhookTemplate) {
    pageTitle = sharedMessages.addWebhook
  } else if (isCustom) {
    pageTitle = {
      ...m.addWebhookViaTemplate,
      values: {
        templateName: webhookTemplate.name,
      },
    }
  }

  // Render Not Found error when the template was not found.
  if (!isCustom && templateId && !webhookTemplate) {
    return <GenericNotFound />
  }

  return (
    <Container>
      <PageTitle title={pageTitle} className="mb-0" hideHeading={Boolean(webhookTemplate)} />
      <Row>
        <Col lg={8} md={12}>
          <WebhookAdd
            appId={appId}
            templateId={templateId}
            webhookTemplate={webhookTemplate}
            navigateToList={navigateToList}
          />
        </Col>
      </Row>
    </Container>
  )
}

export default ApplicationWebhookAddForm
