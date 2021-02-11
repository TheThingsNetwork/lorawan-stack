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

import { merge } from 'lodash'

import { getWebhookId } from '@ttn-lw/lib/selectors/id'

import {
  GET_WEBHOOK,
  GET_WEBHOOK_SUCCESS,
  GET_WEBHOOKS_LIST_SUCCESS,
  UPDATE_WEBHOOK_SUCCESS,
} from '@console/store/actions/webhooks'

const defaultState = {
  selectedWebhook: null,
  totalCount: undefined,
  entities: {},
}

const webhooks = function (state = defaultState, { type, payload }) {
  switch (type) {
    case GET_WEBHOOK:
      return {
        ...state,
        selectedWebhook: payload.webhookId,
      }
    case GET_WEBHOOK_SUCCESS:
      return {
        ...state,
        entities: {
          ...state.entities,
          [getWebhookId(payload)]: payload,
        },
      }
    case GET_WEBHOOKS_LIST_SUCCESS:
      return {
        ...state,
        entities: {
          ...payload.entities.reduce((acc, webhook) => {
            acc[getWebhookId(webhook)] = webhook
            return acc
          }, {}),
        },
        totalCount: payload.totalCount,
      }
    case UPDATE_WEBHOOK_SUCCESS:
      const webhookId = getWebhookId(payload)

      return {
        ...state,
        entities: {
          ...state.entities,
          [webhookId]: merge({}, state.entities[webhookId], payload),
        },
      }
    default:
      return state
  }
}

export default webhooks
