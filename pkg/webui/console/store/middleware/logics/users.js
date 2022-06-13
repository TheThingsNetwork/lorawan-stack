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

import tts from '@console/api/tts'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as users from '@console/store/actions/users'

const getUserLogic = createRequestLogic({
  type: users.GET_USER,
  process: async ({ action }) => {
    const {
      payload: { id },
      meta: { selector },
    } = action

    return await tts.Users.getById(id, selector)
  },
})

const updateUserLogic = createRequestLogic({
  type: users.UPDATE_USER,
  process: ({ action }) => {
    const {
      payload: { id, patch },
    } = action

    return tts.Users.updateById(id, patch)
  },
})

const deleteUserLogic = createRequestLogic({
  type: users.DELETE_USER,
  process: async ({ action }) => {
    const {
      payload: { id },
      meta: { options },
    } = action

    if (options.purge) {
      await tts.Users.purgeById(id)
    } else {
      await tts.Users.deleteById(id)
    }

    return { id }
  },
})

const getUsersLogic = createRequestLogic({
  type: users.GET_USERS_LIST,
  process: async ({ action }) => {
    const {
      params: { page, limit, query, order },
    } = action.payload
    const { selectors } = action.meta

    const data = query
      ? await tts.Users.search(
          {
            page,
            limit,
            query,
            order,
          },
          selectors,
        )
      : await tts.Users.getAll({ page, limit, order }, selectors)

    return { entities: data.users, totalCount: data.totalCount }
  },
})

const createUserLogic = createRequestLogic({
  type: users.CREATE_USER,
  process: async ({ action }) => {
    const {
      payload: { user },
    } = action

    return await tts.Users.create(user)
  },
})

const getUsersRightsLogic = createRequestLogic({
  type: users.GET_USER_RIGHTS_LIST,
  process: async ({ action }) => {
    const { id } = action.payload
    const result = await tts.Users.getRightsById(id)

    return result.rights.sort()
  },
})

export default [
  getUserLogic,
  getUsersLogic,
  updateUserLogic,
  deleteUserLogic,
  createUserLogic,
  getUsersRightsLogic,
]
