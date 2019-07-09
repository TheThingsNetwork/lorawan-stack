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


import { createLogic } from 'redux-logic'

const getResultActionFromType = function (typeString, status) {
  if (typeString instanceof Array) {
    if (typeString.length === 1) {
      return typeString[0].replace('REQUEST', status)
    }

    return undefined
  }
  return typeString.replace('REQUEST', status)
}

/**
 * Logic creator for request logics, it will handle promise resolution, as well
 * as result action dispatch automatically
 * @param {Object} options - The logic options (to be passed to `createLogic()`)
 * @param {(string\|function)} successType - The success action type or action creator
 * @param {(string\|function)} failType - The fail action type or action creator
 * @returns {Object} The `redux-logic` (decorated) logic
 */
const createRequestLogic = function (
  options,
  successType = getResultActionFromType(options.type, 'SUCCESS'),
  failType = getResultActionFromType(options.type, 'FAILURE'),
) {

  if (!successType || !failType) {
    throw new Error('Could not derive result actions from provided options')
  }

  let successAction = successType
  let failAction = failType

  if (typeof successType === 'string') {
    successAction = payload => ({ type: successType, payload })
  }
  if (typeof failType === 'string') {
    failAction = error => ({ type: failType, error: true, payload: error })
  }

  return createLogic({
    ...options,
    async process (deps, dispatch, done) {
      const { meta: { _resolve }} = deps.action
      let res, resultAction

      try {
        res = await options.process(deps, dispatch)
        resultAction = successAction(res)
      } catch (error) {
        resultAction = failAction(error)
      }

      dispatch(resultAction)

      // Resolve the promise also on failure actions, as the dispatch and logic
      // operated correctly. We never want to reject the promise, as doing so
      // could lead to unwanted unresolved promise rejections when dispatching.
      _resolve(resultAction)

      done()
    },
  })
}

export default createRequestLogic
