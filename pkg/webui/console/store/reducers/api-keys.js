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

import { GET_API_KEYS_LIST_SUCCESS, GET_API_KEY_SUCCESS, GET_API_KEY } from '../actions/api-keys'
import { getApiKeyId } from '../../../lib/selectors/id'

const defaultState = {
  entities: {},
  selectedApiKey: null,
}

const apiKey = (state = {}, apiKey) => ({
  ...state,
  ...apiKey,
})

const apiKeys = function(state = defaultState, { type, payload }) {
  switch (type) {
    case GET_API_KEY:
      return {
        ...state,
        selectedApiKey: payload.keyId,
      }
    case GET_API_KEY_SUCCESS:
      const id = getApiKeyId(payload)
      return {
        ...state,
        entities: {
          ...state.entities,
          [id]: apiKey(state.entities[id], payload),
        },
      }
    case GET_API_KEYS_LIST_SUCCESS:
      return {
        ...state,
        entities: {
          ...payload.entities.reduce(
            (acc, ak) => {
              const id = getApiKeyId(ak)
              acc[id] = apiKey(state.entities[id], ak)
              return acc
            },
            { ...state.entities },
          ),
        },
      }
    default:
      return state
  }
}

export default apiKeys
