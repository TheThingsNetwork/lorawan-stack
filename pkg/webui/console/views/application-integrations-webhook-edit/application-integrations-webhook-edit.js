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

import tts from '@console/api/tts'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import toast from '@ttn-lw/components/toast'

import WebhookForm from '@console/components/webhook-form'

import diff from '@ttn-lw/lib/diff'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './application-integration-webhook-edit.styl'

const m = defineMessages({
  editWebhook: 'Edit webhook',
  updateSuccess: 'Webhook updated',
  deleteSuccess: 'Webhook deleted',
  reactivateSuccess: 'Webhook activated',
})

const ApplicationWebhookEdit = props => {
  const { webhook, appId, webhookTemplate, updateWebhook, match, navigateToList } = props
  const { webhookId } = match.params

  useBreadcrumbs(
    'apps.single.integrations.webhooks.edit',
    <Breadcrumb
      path={`/applications/${appId}/integrations/${webhookId}`}
      content={sharedMessages.edit}
    />,
  )

  const handleSubmit = React.useCallback(
    async updatedWebhook => {
      if ('basic_auth' in updatedWebhook && !updatedWebhook.basic_auth?.value) {
        const { Authorization, ...restHeaders } = updatedWebhook.headers
        updatedWebhook.headers = restHeaders
      }
      console.log(updatedWebhook)
      const patch = diff(webhook, updatedWebhook, ['ids'])

      // Ensure that the header prop is always patched fully, otherwise we loose
      // old header entries.
      if ('headers' in patch) {
        patch.headers = updatedWebhook.headers
      }

      await updateWebhook(patch)
    },
    [updateWebhook, webhook],
  )
  const handleSubmitSuccess = React.useCallback(() => {
    toast({
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
  }, [])

  const handleDelete = React.useCallback(async () => {
    await tts.Applications.Webhooks.deleteById(appId, webhookId)
  }, [appId, webhookId])
  const handleDeleteSuccess = React.useCallback(() => {
    toast({
      message: m.deleteSuccess,
      type: toast.types.SUCCESS,
    })

    navigateToList()
  }, [navigateToList])

  const handleReactivateSuccess = React.useCallback(() => {
    toast({
      message: m.reactivateSuccess,
      type: toast.types.SUCCESS,
    })
  }, [])
  const handleReactivate = React.useCallback(
    async updatedHealthStatus => {
      await tts.Applications.Webhooks.updateById(appId, webhookId, updatedHealthStatus, [
        'health_status',
      ])
    },
    [appId, webhookId],
  )

  return (
    <Container>
      <PageTitle title={m.editWebhook} />
      <Row>
        <Col lg={8} md={12}>
          <WebhookForm
            update
            appId={appId}
            initialWebhookValue={webhook}
            webhookTemplate={webhookTemplate}
            onSubmit={handleSubmit}
            onSubmitSuccess={handleSubmitSuccess}
            onDelete={handleDelete}
            onDeleteSuccess={handleDeleteSuccess}
            onReactivate={handleReactivate}
            onReactivateSuccess={handleReactivateSuccess}
            reactivateButtonMessage={m.reactivateButton}
            suspendedWebhookMessage={m.suspendedWebhookMessage}
            buttonStyle={style}
          />
        </Col>
      </Row>
    </Container>
  )
}

ApplicationWebhookEdit.propTypes = {
  appId: PropTypes.string.isRequired,
  match: PropTypes.match.isRequired,
  navigateToList: PropTypes.func.isRequired,
  updateWebhook: PropTypes.func.isRequired,
  webhook: PropTypes.webhook.isRequired,
  webhookTemplate: PropTypes.webhookTemplate,
}

ApplicationWebhookEdit.defaultProps = {
  webhookTemplate: undefined,
}

export default ApplicationWebhookEdit
