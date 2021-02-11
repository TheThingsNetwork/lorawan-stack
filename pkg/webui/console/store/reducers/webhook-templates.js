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

import { getWebhookTemplateId } from '@ttn-lw/lib/selectors/id'

import {
  LIST_WEBHOOK_TEMPLATES_SUCCESS,
  GET_WEBHOOK_TEMPLATE_SUCCESS,
} from '@console/store/actions/webhook-templates'

const defaultState = {
  entities: undefined,
}

const webhookTemplates = function (state = defaultState, { type, payload }) {
  switch (type) {
    case LIST_WEBHOOK_TEMPLATES_SUCCESS:
      const entities = payload.reduce(
        (acc, template) => {
          acc[getWebhookTemplateId(template)] = template

          return acc
        },
        { ...state.entities },
      )
      return {
        ...state,
        entities,
      }
    case GET_WEBHOOK_TEMPLATE_SUCCESS:
      return {
        ...state,
        entities: {
          ...state.entities,
          [getWebhookTemplateId(payload)]: payload,
        },
      }
    default:
      return state
  }
}

export default webhookTemplates
