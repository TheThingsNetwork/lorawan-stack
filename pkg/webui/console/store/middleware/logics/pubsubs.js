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

import * as pubsubs from '@console/store/actions/pubsubs'
import * as pubsubFormats from '@console/store/actions/pubsub-formats'

const createPubsubLogic = createRequestLogic({
  type: pubsubs.CREATE_PUBSUB,
  process: async ({ action }) => {
    const { appId, pubsub } = action.payload

    return await tts.Applications.PubSubs.create(appId, pubsub)
  },
})

const getPubsubLogic = createRequestLogic({
  type: pubsubs.GET_PUBSUB,
  process: async ({ action }) => {
    const {
      payload: { appId, pubsubId },
      meta: { selector },
    } = action

    return await tts.Applications.PubSubs.getById(appId, pubsubId, selector)
  },
})

const getPubsubsLogic = createRequestLogic({
  type: pubsubs.GET_PUBSUBS_LIST,
  process: async ({ action }) => {
    const {
      id: appId,
      params: { page, limit, order },
    } = action.payload
    const { selectors } = action.meta
    const res = await tts.Applications.PubSubs.getAll(appId, { page, limit, order }, selectors)

    return { entities: res.pubsubs, totalCount: res.totalCount }
  },
})

const updatePubsubsLogic = createRequestLogic({
  type: pubsubs.UPDATE_PUBSUB,
  process: async ({ action }) => {
    const { appId, pubsubId, patch } = action.payload

    return await tts.Applications.PubSubs.updateById(appId, pubsubId, patch)
  },
})

const getPubsubFormatsLogic = createRequestLogic({
  type: pubsubFormats.GET_PUBSUB_FORMATS,
  process: async () => {
    const { formats } = await tts.Applications.PubSubs.getFormats()

    return formats
  },
})

const deletePubsub = createRequestLogic({
  type: pubsubs.DELETE_PUBSUB,
  process: async ({ action }) => {
    const { appId, pubsubId } = action.payload

    return await tts.Applications.PubSubs.deleteById(appId, pubsubId)
  },
})

export default [
  createPubsubLogic,
  getPubsubLogic,
  getPubsubsLogic,
  updatePubsubsLogic,
  getPubsubFormatsLogic,
  deletePubsub,
]
