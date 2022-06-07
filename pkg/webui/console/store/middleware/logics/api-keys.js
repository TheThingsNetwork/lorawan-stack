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
import { entitySdkServiceMap } from '@console/constants/entities'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as apiKeys from '@console/store/actions/api-keys'

const validParentTypes = Object.keys(entitySdkServiceMap)

const parentTypeValidator = ({ action }, allow) => {
  if (!validParentTypes.includes(action.payload.parentType)) {
    // Do not reject the action but throw an error, as this is an implementation
    // error.
    throw new Error(`Invalid parent entity type ${action.payload.parentType}`)
  }
  allow(action)
}

const getApiKeysLogic = createRequestLogic({
  type: apiKeys.GET_API_KEYS_LIST,
  validate: parentTypeValidator,
  process: async ({ action }) => {
    const {
      parentType,
      parentId,
      params: { page, limit, order },
    } = action.payload
    const data = await tts[entitySdkServiceMap[parentType]].ApiKeys.getAll(parentId, {
      limit,
      page,
      order,
    })

    return { parentType, entities: data.api_keys, totalCount: data.totalCount }
  },
})

const getApiKeyLogic = createRequestLogic({
  type: apiKeys.GET_API_KEY,
  validate: parentTypeValidator,
  process: async ({ action }) => {
    const { parentType, parentId, keyId } = action.payload

    return tts[entitySdkServiceMap[parentType]].ApiKeys.getById(parentId, keyId)
  },
})

const createApplicationApiKeyLogic = createRequestLogic({
  type: apiKeys.CREATE_APP_API_KEY,
  process: async ({ action }) => {
    const { id, key } = action.payload

    return tts.Applications.ApiKeys.create(id, key)
  },
})

export default [getApiKeyLogic, getApiKeysLogic, createApplicationApiKeyLogic]
