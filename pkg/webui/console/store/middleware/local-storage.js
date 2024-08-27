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

import { pick } from 'lodash'

import log from '@ttn-lw/lib/log'

export const localStorageMiddleware = (paths, userIdSelector) => store => next => action => {
  const result = next(action)
  const state = store.getState()
  const userId = userIdSelector(state)
  if (userId === undefined) {
    // Since the persisted state is user-specific, we don't persist the state
    // if the user is not authenticated.
    return result
  }
  const stateToPersist = pick(state, paths)
  localStorage.setItem(`${userId}/console-state`, JSON.stringify(stateToPersist))
  return result
}

export const loadStateFromLocalStorage = (userId, validators) => {
  try {
    const serializedState = localStorage.getItem(`${userId}/console-state`)
    if (serializedState === null) {
      return {}
    }
    const parsedData = JSON.parse(serializedState)
    for (const [key, validator] of Object.entries(validators)) {
      if (key in parsedData) {
        parsedData[key] = validator(parsedData[key])
      }
    }
    return parsedData
  } catch (err) {
    log('Failed to load state from local storage', err)
    return {}
  }
}
