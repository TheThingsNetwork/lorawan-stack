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

import React, { useState } from 'react'
import urlTemplate from 'url-template'

import WebhookForm from '@console/components/webhook-form'
import WebhookTemplateForm from '@console/components/webhook-template-form'

import { isNotFoundError } from '@ttn-lw/lib/errors/utils'
import PropTypes from '@ttn-lw/lib/prop-types'

const pathExpand = (url, fields) =>
  Boolean(url) && url.path ? { path: urlTemplate.parse(url.path).expand(fields) } : url

const WebhookAdd = props => {
  const {
    appId,
    createApplicationApiKey,
    createWebhook,
    getWebhook,
    navigateToList,
    templateId,
    webhookTemplate,
  } = props

  const convertTemplateToWebhook = React.useCallback(
    async values => {
      const template = webhookTemplate
      const { webhook_id, ...fields } = values

      const headers = Object.keys(template.headers || {}).reduce((acc, key) => {
        const val = template.headers[key]
        const headerValue = val.replace(/{([a-z0-9_-]+)}/i, (_, field) => values[field])
        if (headerValue !== '') {
          acc[key] = headerValue
        }
        return acc
      }, {})

      const webhook = {
        ids: {
          webhook_id: values.webhook_id,
        },
        template_ids: template.ids,
        format: template.format,
        headers,
        template_fields: fields,
        base_url: urlTemplate.parse(template.base_url).expand(fields),
        uplink_message: pathExpand(template.uplink_message, fields),
        uplink_normalized: pathExpand(template.uplink_normalized, fields),
        join_accept: pathExpand(template.join_accept, fields),
        downlink_ack: pathExpand(template.downlink_ack, fields),
        downlink_nack: pathExpand(template.downlink_nack, fields),
        downlink_sent: pathExpand(template.downlink_sent, fields),
        downlink_failed: pathExpand(template.downlink_failed, fields),
        downlink_queued: pathExpand(template.downlink_queued, fields),
        downlink_queue_invalidated: pathExpand(template.downlink_queue_invalidated, fields),
        location_solved: pathExpand(template.location_solved, fields),
        service_data: pathExpand(template.service_data, fields),
        field_mask: template.field_mask,
      }

      if (template.create_downlink_api_key) {
        const key = {
          name: `${webhook_id} downlink API key`,
          rights: ['RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE'],
        }
        const { key: downlink_api_key } = await createApplicationApiKey(appId, key)
        webhook.downlink_api_key = downlink_api_key
      }

      return webhook
    },
    [appId, createApplicationApiKey, webhookTemplate],
  )

  const existCheck = React.useCallback(
    async webhookId => {
      try {
        await getWebhook(appId, webhookId, [])
        return true
      } catch (error) {
        if (isNotFoundError(error)) {
          return false
        }

        throw error
      }
    },
    [appId, getWebhook],
  )

  const handleSubmit = React.useCallback(
    async webhook => {
      await createWebhook(appId, webhook)
    },
    [appId, createWebhook],
  )

  const handleSubmitSuccess = React.useCallback(() => {
    navigateToList(appId)
  }, [appId, navigateToList])

  const [error, setError] = useState()
  const handleWebhookSubmit = React.useCallback(
    async (values, webhook, { setSubmitting, resetForm }) => {
      setError(undefined)
      try {
        const result = await handleSubmit(webhook)

        resetForm({ values })
        handleSubmitSuccess(result)
      } catch (error) {
        setSubmitting(false)
        setError(error)
      }
    },
    [handleSubmit, handleSubmitSuccess],
  )

  if (!Boolean(webhookTemplate)) {
    return (
      <WebhookForm
        update={false}
        onSubmit={handleWebhookSubmit}
        existCheck={existCheck}
        error={error}
      />
    )
  }

  return (
    <WebhookTemplateForm
      appId={appId}
      templateId={templateId}
      onSubmit={handleWebhookSubmit}
      webhookTemplate={webhookTemplate}
      convertTemplateToWebhook={convertTemplateToWebhook}
      existCheck={existCheck}
      error={error}
    />
  )
}

WebhookAdd.propTypes = {
  appId: PropTypes.string.isRequired,
  createApplicationApiKey: PropTypes.func.isRequired,
  createWebhook: PropTypes.func.isRequired,
  getWebhook: PropTypes.func.isRequired,
  navigateToList: PropTypes.func.isRequired,
  templateId: PropTypes.string.isRequired,
  webhookTemplate: PropTypes.webhookTemplate,
}

WebhookAdd.defaultProps = {
  webhookTemplate: undefined,
}

export default WebhookAdd
