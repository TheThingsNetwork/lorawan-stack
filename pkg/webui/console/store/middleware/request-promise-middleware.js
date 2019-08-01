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

/**
 * This middleware will check for request actions and attach a promise to the
 * action.
 * @param {Object} store - The store to apply the middleware to
 * @returns {Object} The middleware
 */
const requestPromiseMiddleware = store => next => function (action) {
  if (action.meta && action.meta._attachPromise) {
    return new Promise(function (resolve, reject) {
      action.meta = {
        ...action.meta,
        _resolve: resolve,
        _reject: reject,
      }
      next(action)
    })
  }

  return next(action)
}

export default requestPromiseMiddleware
