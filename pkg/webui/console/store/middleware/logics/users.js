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

import api from '@console/api'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as users from '@console/store/actions/users'

const getUserLogic = createRequestLogic({
  type: users.GET_USER,
  process({ action }, dispatch) {
    const {
      payload: { id },
      meta: { selector },
    } = action

    return api.users.get(id, selector)
  },
})

const updateUserLogic = createRequestLogic({
  type: users.UPDATE_USER,
  process({ action }, dispatch) {
    const {
      payload: { id, patch },
    } = action

    return api.users.update(id, patch)
  },
})

const deleteUserLogic = createRequestLogic({
  type: users.DELETE_USER,
  async process({ action }, dispatch) {
    const {
      payload: { id },
    } = action

    await api.users.delete(id)

    return { id }
  },
})

const getUsersLogic = createRequestLogic({
  type: users.GET_USERS_LIST,
  async process({ action }) {
    const {
      params: { page, limit, query, order },
    } = action.payload
    const { selectors } = action.meta

    const data = query
      ? await api.users.search(
          {
            page,
            limit,
            id_contains: query,
            order,
          },
          selectors,
        )
      : await api.users.list({ page, limit, order }, selectors)

    return { entities: data.users, totalCount: data.totalCount }
  },
})

const createUserLogic = createRequestLogic({
  type: users.CREATE_USER,
  process({ action }, dispatch) {
    const {
      payload: { user },
    } = action

    return api.users.create(user)
  },
})

export default [getUserLogic, getUsersLogic, updateUserLogic, deleteUserLogic, createUserLogic]
