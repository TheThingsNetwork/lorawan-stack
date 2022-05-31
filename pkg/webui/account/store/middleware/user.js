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

import axios from 'axios'

import tts from '@account/api/tts'

import api from '@account/api'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'
import {
  isGravatarProfilePicture,
  getSmallestAvailableProfilePicture,
} from '@ttn-lw/lib/selectors/profile-picture'

import * as user from '@account/store/actions/user'

import { selectUserId } from '@account/store/selectors/user'

const fixProfilePicture = async result => {
  if (isGravatarProfilePicture(result.profile_picture)) {
    const src = getSmallestAvailableProfilePicture(result.profile_picture)
    try {
      await axios.get(src)
    } catch (err) {
      result.profile_picture = null
    }
  }
}

const logoutLogic = createRequestLogic({
  type: user.LOGOUT,
  process: async () => {
    await api.account.logout()
    window.location.reload()
  },
})

const getUserLogic = createRequestLogic({
  type: user.GET_USER,
  process: async ({ action, getState }) => {
    const {
      meta: { selector },
    } = action
    const userId =
      'payload' in action && action.payload.id ? action.payload.id : selectUserId(getState())

    const result = await tts.Users.getById(userId, selector)
    await fixProfilePicture(result)

    return result
  },
})

const updateUserLogic = createRequestLogic({
  type: user.UPDATE_USER,
  process: async ({ action, getState }) => {
    const userId =
      'payload' in action && action.payload.id ? action.payload.id : selectUserId(getState())
    const { patch } = action.payload

    const result = await tts.Users.updateById(userId, patch)
    await fixProfilePicture(result)

    return { ...patch, ...result }
  },
})

const deleteUserLogic = createRequestLogic({
  type: user.DELETE_USER,
  process: async ({ action, getState }) => {
    const userId =
      'payload' in action && action.payload.id ? action.payload.id : selectUserId(getState())
    const { options } = action.meta

    if (options.purge) {
      return await tts.Users.purgeById(userId)
    }

    return await tts.Users.deleteById(userId)
  },
})

const getUserRightsLogic = createRequestLogic({
  type: user.GET_USER_RIGHTS,
  process: async ({ action }) => {
    const { userId } = action.payload
    const result = await tts.Users.getRightsById(userId)

    return result.rights.sort()
  },
})

export default [logoutLogic, getUserLogic, updateUserLogic, deleteUserLogic, getUserRightsLogic]
