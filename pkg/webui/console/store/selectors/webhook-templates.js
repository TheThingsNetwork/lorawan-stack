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

import { createFetchingSelector } from '@ttn-lw/lib/store/selectors/fetching'
import { createErrorSelector } from '@ttn-lw/lib/store/selectors/error'

import {
  LIST_WEBHOOK_TEMPLATES_BASE,
  GET_WEBHOOK_TEMPLATE_BASE,
} from '@console/store/actions/webhook-templates'

const selectWebhookTemplatesStore = state => state.webhookTemplates
const selectWebhookTemplatesEntitiesStore = state => selectWebhookTemplatesStore(state).entities

export const selectWebhookTemplateById = function (state, id) {
  const entities = selectWebhookTemplatesEntitiesStore(state)
  if (!Boolean(entities)) return undefined

  return entities[id]
}
export const selectWebhookTemplateError = createErrorSelector(GET_WEBHOOK_TEMPLATE_BASE)
export const selectWebhookTemplateFetching = createFetchingSelector(GET_WEBHOOK_TEMPLATE_BASE)

export const selectWebhookTemplates = function (state) {
  const { entities } = selectWebhookTemplatesStore(state)

  if (!Boolean(entities)) return undefined

  return Object.keys(entities).map(key => entities[key])
}
export const selectWebhookTemplatesError = createErrorSelector(LIST_WEBHOOK_TEMPLATES_BASE)
export const selectWebhookTemplatesFetching = createFetchingSelector(LIST_WEBHOOK_TEMPLATES_BASE)
