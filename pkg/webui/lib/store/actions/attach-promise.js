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

import { createAction } from 'redux-actions'

export const createRequestActions = (baseType, requestPayloadCreator, requestMetaCreator) => {
  const requestType = `${baseType}_REQUEST`
  const successType = `${baseType}_SUCCESS`
  const failureType = `${baseType}_FAILURE`
  const abortType = `${baseType}_ABORT`

  return [
    {
      request: requestType,
      success: successType,
      failure: failureType,
      abort: abortType,
    },
    {
      request: createAction(requestType, requestPayloadCreator, requestMetaCreator),
      success: createAction(successType),
      failure: createAction(failureType),
      abort: createAction(abortType),
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
export default actionOrActionCreator => {
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

  return (...args) => {
    const action = actionOrActionCreator(...args)
    return decorateAction(action)
  }
}

/**
 * Helper function to retrieve the result action types based
 * on the request action type.
 *
 * @param {string} typeString - The request action type.
 * @param {string} status - The result type, either `SUCCESS`, `FAILURE` or `ABORT`.
 * @returns {string} - The result action type.
 */
export const getResultActionFromType = (typeString, status) => typeString.replace('REQUEST', status)
