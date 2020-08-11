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

import { createAction } from 'redux-actions'

export const createRequestActions = function(baseType, requestPayloadCreator, requestMetaCreator) {
  const requestType = `${baseType}_REQUEST`
  const successType = `${baseType}_SUCCESS`
  const failureType = `${baseType}_FAILURE`

  return [
    {
      request: requestType,
      success: successType,
      failure: failureType,
    },
    {
      request: createAction(requestType, requestPayloadCreator, requestMetaCreator),
      success: createAction(successType),
      failure: createAction(failureType),
    },
  ]
}

/**
 * The attachPromise function extends an action creator or action to include a
 * flag which results in a promise being attached to the action by the promise
 * middleware.
 *
 * @param {object|Function} actionOrActionCreator - The original action or
 * action creator.
 * @returns {object|Function} - The modified action or action creator.
 */
export const attachPromise = actionOrActionCreator => {
  const decorateAction = action => ({
    ...action,
    meta: {
      ...action.meta,
      _attachPromise: true,
    },
  })

  if (typeof actionOrActionCreator === 'object') {
    return decorateAction(actionOrActionCreator)
  }

  return function(...args) {
    const action = actionOrActionCreator(...args)
    return decorateAction(action)
  }
}
