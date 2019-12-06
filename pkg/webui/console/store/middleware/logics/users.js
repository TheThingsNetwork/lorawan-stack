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

import * as users from '../../actions/users'

import api from '../../../api'
import createRequestLogic from './lib'

const getUserLogic = createRequestLogic({
  type: users.GET_USER,
  async process({ action }, dispatch) {
    const {
      payload: { id },
      meta: { selector },
    } = action

    const res = await api.users.get(id, selector)

    // Ensure to set STATE_REQUESTED, which gets stripped as null value from
    // the backend response
    if (selector.includes('state') && !('state' in res)) {
      res.state = 'STATE_REQUESTED'
    }

    return res
  },
})

const updateUserLogic = createRequestLogic({
  type: users.UPDATE_USER,
  async process({ action }, dispatch) {
    const {
      payload: { id, patch },
    } = action

    const res = await api.users.update(id, patch)

    // Ensure to set STATE_REQUESTED, which gets stripped as null value from
    // the backend response
    if ('state' in patch && !('state' in res)) {
      res.state = 'STATE_REQUESTED'
    }

    return res
  },
})

const getUsersLogic = createRequestLogic({
  type: users.GET_USERS_LIST,
  async process({ action }) {
    const {
      params: { page, limit, query },
    } = action.payload
    const { selectors } = action.meta

    const data = query
      ? await api.users.search(
          {
            page,
            limit,
            id_contains: query,
          },
          selectors,
        )
      : await api.users.list({ page, limit }, selectors)

    // Ensure to set STATE_REQUESTED, which gets stripped as null value from
    // the backend response
    for (const user of data.users) {
      if (selectors.includes('state') && !('state' in user)) {
        user.state = 'STATE_REQUESTED'
      }
    }

    return { entities: data.users, totalCount: data.totalCount }
  },
})

export default [getUserLogic, getUsersLogic, updateUserLogic]
