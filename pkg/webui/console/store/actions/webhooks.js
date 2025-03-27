// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'
import {
  createPaginationBaseActionType,
  createPaginationByIdRequestActions,
} from '@ttn-lw/lib/store/actions/pagination'

export const GET_WEBHOOK_BASE = 'GET_WEBHOOK'
export const [
  { request: GET_WEBHOOK, success: GET_WEBHOOK_SUCCESS, failure: GET_WEBHOOK_FAILURE },
  { request: getWebhook, success: getWebhookSuccess, failure: getWebhookFailure },
] = createRequestActions(
  GET_WEBHOOK_BASE,
  (appId, webhookId) => ({ appId, webhookId }),
  (appId, webhookId, selector) => ({ selector }),
)

export const GET_WEBHOOKS_LIST_BASE = createPaginationBaseActionType('WEBHOOKS')
export const [
  {
    request: GET_WEBHOOKS_LIST,
    success: GET_WEBHOOKS_LIST_SUCCESS,
    failure: GET_WEBHOOKS_LIST_FAILURE,
  },
  { request: getWebhooksList, success: getWebhooksListSuccess, failure: getWebhooksListFailure },
] = createPaginationByIdRequestActions('WEBHOOKS')

export const UPDATE_WEBHOOK_BASE = 'UPDATE_WEBHOOK'
export const [
  { request: UPDATE_WEBHOOK, success: UPDATE_WEBHOOK_SUCCESS, failure: UPDATE_WEBHOOK_FAILURE },
  { request: updateWebhook, success: updateWebhookSuccess, failure: updateWebhookFailure },
] = createRequestActions(UPDATE_WEBHOOK_BASE, (appId, webhookId, patch) => ({
  appId,
  webhookId,
  patch,
}))

export const CREATE_WEBHOOK_BASE = 'CREATE_WEBHOOK'
export const [
  { request: CREATE_WEBHOOK, success: CREATE_WEBHOOK_SUCCESS, failure: CREATE_WEBHOOK_FAILURE },
  { request: createWebhook, success: createWebhookSuccess, failure: createWebhookFailure },
] = createRequestActions(CREATE_WEBHOOK_BASE, (appId, webhook) => ({ appId, webhook }))
