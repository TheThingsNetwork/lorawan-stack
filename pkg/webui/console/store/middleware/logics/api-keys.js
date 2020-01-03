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

import * as apiKeys from '../../actions/api-keys'

import api from '../../../api'
import createRequestLogic from './lib'

const validParentTypes = ['application', 'gateway', 'organization']

const parentTypeValidator = function({ action }, allow) {
  if (!validParentTypes.includes(action.payload.parentType)) {
    // Do not reject the action but throw an error, as this is an implementation
    // error
    throw new Error(`Invalid parent entity type ${action.payload.parentType}`)
  }
  allow(action)
}

const getApiKeysLogic = createRequestLogic({
  type: apiKeys.GET_API_KEYS_LIST,
  validate: parentTypeValidator,
  async process({ getState, action }) {
    const {
      parentType,
      parentId,
      params: { page, limit },
    } = action.payload
    const data = await api[parentType].apiKeys.list(parentId, { limit, page })
    return { parentType, entities: data.api_keys, totalCount: data.totalCount }
  },
})

const getApiKeyLogic = createRequestLogic({
  type: apiKeys.GET_API_KEY,
  validate: parentTypeValidator,
  async process({ action }) {
    const { parentType, parentId, keyId } = action.payload
    return api[parentType].apiKeys.get(parentId, keyId)
  },
})

export default [getApiKeyLogic, getApiKeysLogic]
