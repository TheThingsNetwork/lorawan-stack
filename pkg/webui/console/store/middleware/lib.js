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
  successType = options.type.replace('REQUEST', 'SUCCESS'),
  failType = options.type.replace('REQUEST', 'FAILURE'),
) {
  let successAction = successType
  let failAction = failType

  if (typeof successType === 'string') {
    successAction = payload => ({ type: successType, payload })
  }
  if (typeof failType === 'string') {
    failAction = error => ({ type: failType, error })
  }

  return createLogic({
    ...options,
    async process (deps, dispatch, done) {
      const { meta: { _reject, _resolve }} = deps.action
      try {
        const res = await options.process(deps, dispatch)
        dispatch(successAction(res))
        _resolve(res)
      } catch (e) {
        dispatch(failAction(e))
        _reject(e)
      }

      done()
    },
  })
}

export default createRequestLogic
