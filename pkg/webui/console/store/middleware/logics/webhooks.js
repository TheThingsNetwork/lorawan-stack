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

import api from '@console/api'

import * as webhooks from '@console/store/actions/webhooks'
import * as webhookFormats from '@console/store/actions/webhook-formats'
import * as webhookTemplates from '@console/store/actions/webhook-templates'

import createRequestLogic from './lib'

const getWebhookLogic = createRequestLogic({
  type: webhooks.GET_WEBHOOK,
  async process({ action }) {
    const {
      payload: { appId, webhookId },
      meta: { selector },
    } = action
    return api.application.webhooks.get(appId, webhookId, selector)
  },
})

const getWebhooksLogic = createRequestLogic({
  type: webhooks.GET_WEBHOOKS_LIST,
  async process({ action }) {
    const {
      payload: { appId },
      meta: { selector },
    } = action
    const res = await api.application.webhooks.list(appId, selector)
    return { entities: res.webhooks, totalCount: res.totalCount }
  },
})

const updateWebhookLogic = createRequestLogic({
  type: webhooks.UPDATE_WEBHOOK,
  async process({ action }) {
    const { appId, webhookId, patch } = action.payload

    return api.application.webhooks.update(appId, webhookId, patch)
  },
})

const getWebhookFormatsLogic = createRequestLogic({
  type: webhookFormats.GET_WEBHOOK_FORMATS,
  async process() {
    const { formats } = await api.application.webhooks.getFormats()
    return formats
  },
})

const getWebhookTemplateLogic = createRequestLogic({
  type: webhookTemplates.GET_WEBHOOK_TEMPLATE,
  async process({ action }) {
    const { id } = action.payload
    const { selector } = action.meta
    const template = await api.application.webhooks.getTemplate(id, selector)

    return template
  },
})

const getWebhookTemplatesLogic = createRequestLogic({
  type: webhookTemplates.LIST_WEBHOOK_TEMPLATES,
  async process({ action }) {
    const { selector } = action.meta
    const { templates } = await api.application.webhooks.listTemplates(selector)

    return templates
  },
})

export default [
  getWebhookLogic,
  getWebhooksLogic,
  updateWebhookLogic,
  getWebhookFormatsLogic,
  getWebhookTemplateLogic,
  getWebhookTemplatesLogic,
]
