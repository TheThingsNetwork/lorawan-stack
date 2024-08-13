// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import tts from '@console/api/tts'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'
import {
  getSmallestAvailableProfilePicture,
  isGravatarProfilePicture,
} from '@ttn-lw/lib/selectors/profile-picture'

import * as user from '@console/store/actions/user'

import { validateRecencyFrequencyEntities } from '@console/store/reducers/recency-frequency-items'

import { selectUserId } from '@console/store/selectors/user'

import { loadStateFromLocalStorage } from '../local-storage'

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

const applyPersistedStateLogic = createRequestLogic({
  type: user.APPLY_PERSISTED_STATE,
  validate: ({ action }, allow, reject) => {
    const userId = action.payload
    if (!userId) {
      reject()
    }

    allow()
  },
  process: async ({ action }) => {
    const userId = action.payload
    const persistedState = loadStateFromLocalStorage(userId, {
      recencyFrequencyItems: validateRecencyFrequencyEntities,
    })

    return persistedState
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
  process: async ({ getState }) => {
    const userId = selectUserId(getState())
    if (!userId) {
      return []
    }
    const result = await tts.Users.getRightsById(userId)

    return result.rights.sort()
  },
})

export default [applyPersistedStateLogic, updateUserLogic, deleteUserLogic, getUserRightsLogic]
