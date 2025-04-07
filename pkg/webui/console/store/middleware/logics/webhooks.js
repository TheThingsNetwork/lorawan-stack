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

import tts from '@console/api/tts'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as webhooks from '@console/store/actions/webhooks'
import * as webhookFormats from '@console/store/actions/webhook-formats'
import * as webhookTemplates from '@console/store/actions/webhook-templates'

const getWebhookLogic = createRequestLogic({
  type: webhooks.GET_WEBHOOK,
  process: async ({ action }) => {
    const {
      payload: { appId, webhookId },
      meta: { selector },
    } = action

    return await tts.Applications.Webhooks.getById(appId, webhookId, selector)
  },
})

const getWebhooksLogic = createRequestLogic({
  type: webhooks.GET_WEBHOOKS_LIST,
  process: async ({ action }) => {
    const {
      id: appId,
      params: { page, limit },
    } = action.payload
    const { selectors } = action.meta

    const res = await tts.Applications.Webhooks.getAll(appId, { page, limit }, selectors)

    return { entities: res.webhooks, totalCount: res.totalCount }
  },
})

const updateWebhookLogic = createRequestLogic({
  type: webhooks.UPDATE_WEBHOOK,
  process: async ({ action }) => {
    const { appId, webhookId, patch } = action.payload

    return await tts.Applications.Webhooks.updateById(appId, webhookId, patch)
  },
})

const getWebhookFormatsLogic = createRequestLogic({
  type: webhookFormats.GET_WEBHOOK_FORMATS,
  process: async () => {
    const { formats } = await tts.Applications.Webhooks.getFormats()

    return formats
  },
})

const getWebhookTemplateLogic = createRequestLogic({
  type: webhookTemplates.GET_WEBHOOK_TEMPLATE,
  process: async ({ action }) => {
    const { id } = action.payload
    const { selector } = action.meta

    return await tts.Applications.Webhooks.getTemplate(id, selector)
  },
})

const getWebhookTemplatesLogic = createRequestLogic({
  type: webhookTemplates.LIST_WEBHOOK_TEMPLATES,
  process: async ({ action }) => {
    const { selector } = action.meta
    const { templates } = await tts.Applications.Webhooks.listTemplates(selector)

    return templates
  },
})

const createWebhookLogic = createRequestLogic({
  type: webhooks.CREATE_WEBHOOK,
  process: async ({ action }) => {
    const { appId, webhook } = action.payload

    return await tts.Applications.Webhooks.create(appId, webhook)
  },
})

export default [
  getWebhookLogic,
  getWebhooksLogic,
  updateWebhookLogic,
  getWebhookFormatsLogic,
  getWebhookTemplateLogic,
  getWebhookTemplatesLogic,
  createWebhookLogic,
]
