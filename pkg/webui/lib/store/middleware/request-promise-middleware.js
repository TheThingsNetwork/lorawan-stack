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

import { CancelablePromise } from 'cancelable-promise'

/**
 * `promisifiedDispatch` is a decorator for the dispatch function that attaches
 * a cancelable promise to the action that it will use as return value.
 * You should usually use the middleware instead.
 *
 * @param {object} dispatch - The to be decorated dispatch function.
 * @returns {Function} - The decorated dispatch function.
 */
export const promisifiedDispatch = dispatch => action => {
  if (action.meta && action.meta._attachPromise && !action.meta._resolve && !action.meta._reject) {
    return new CancelablePromise((resolve, reject) => {
      action.meta = {
        ...action.meta,
        _resolve: resolve,
        _reject: reject,
      }
      dispatch(action)
    })
  }
  return dispatch(action)
}

/**
 * This middleware will check for request actions and attach a cancelable
 * promise to the action.
 *
 * @param {object} store - The store to apply the middleware to.
 * @returns {object} The middleware.
 */
const requestPromiseMiddleware = store => promisifiedDispatch

export default requestPromiseMiddleware
