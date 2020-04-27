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

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'

export const LIST_WEBHOOK_TEMPLATES_BASE = 'LIST_WEBHOOK_TEMPLATES'
export const [
  {
    request: LIST_WEBHOOK_TEMPLATES,
    success: LIST_WEBHOOK_TEMPLATES_SUCCESS,
    failure: LIST_WEBHOOK_TEMPLATES_FAILURE,
  },
  {
    request: listWebhookTemplates,
    success: listWebhookTemplatesSuccess,
    failure: listWebhookTemplatesFailure,
  },
] = createRequestActions(LIST_WEBHOOK_TEMPLATES_BASE, undefined, selector => ({ selector }))

export const GET_WEBHOOK_TEMPLATE_BASE = 'GET_WEBHOOK_TEMPLATE'
export const [
  {
    request: GET_WEBHOOK_TEMPLATE,
    success: GET_WEBHOOK_TEMPLATE_SUCCESS,
    failure: GET_WEBHOOK_TEMPLATE_FAILURE,
  },
  {
    request: getWebhookTemplate,
    success: getWebhookTemplateSuccess,
    failure: getWebhookTemplateFailure,
  },
] = createRequestActions(
  GET_WEBHOOK_TEMPLATE_BASE,
  id => ({ id }),
  (id, selector) => ({ selector }),
)
