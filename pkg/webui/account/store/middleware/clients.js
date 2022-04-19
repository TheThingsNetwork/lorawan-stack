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

import tts from '@account/api/tts'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as clients from '@account/store/actions/clients'

const getClientLogic = createRequestLogic({
  type: clients.GET_CLIENT,
  process: async ({ action }) => {
    const {
      payload: { id },
      meta: { selector },
    } = action
    const client = await tts.Clients.getById(id, selector)
    return client
  },
})

const updateClientLogic = createRequestLogic({
  type: clients.UPDATE_CLIENT,
  process: async ({ action }) => {
    const { id, patch } = action.payload

    const result = await tts.Clients.updateById(id, patch)

    return { ...patch, ...result }
  },
})

const deleteClientLogic = createRequestLogic({
  type: clients.DELETE_CLIENT,
  process: async ({ action }) => {
    const { id } = action.payload
    const { options } = action.meta

    if (options.purge) {
      await tts.Clients.purgeById(id)
    } else {
      await tts.Clients.deleteById(id)
    }

    return { id }
  },
})

const restoreClientLogic = createRequestLogic({
  type: clients.RESTORE_CLIENT,
  process: async ({ action }) => {
    const { id } = action.payload

    await tts.Clients.restoreById(id)

    return { id }
  },
})

const getClientsLogic = createRequestLogic({
  type: clients.GET_CLIENTS_LIST,
  latest: true,
  process: async ({ action }) => {
    const {
      params: { page, limit, query, order, deleted },
    } = action.payload
    const { selectors, options } = action.meta
    const result = options.isSearch
      ? await tts.Clients.search(
          {
            page,
            limit,
            query,
            order,
            deleted,
          },
          selectors,
        )
      : await tts.Clients.getAll({ page, limit, order }, selectors)
    return { entities: result.clients, totalCount: result.totalCount }
  },
})

const getClientRightsLogic = createRequestLogic({
  type: clients.GET_CLIENT_RIGHTS,
  process: async ({ action }) => {
    const { id } = action.payload
    const result = await tts.Clients.getRightsById(id)

    return result.rights.sort()
  },
})

export default [
  getClientLogic,
  updateClientLogic,
  deleteClientLogic,
  restoreClientLogic,
  getClientsLogic,
  getClientRightsLogic,
]
