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
import * as Sentry from '@sentry/browser'

import { error } from '@ttn-lw/lib/log'
import {
  isUnauthenticatedError,
  isInvalidArgumentError,
  getBackendErrorId,
  isUnknown,
} from '@ttn-lw/lib/errors/utils'
import { clear as clearAccessToken } from '@ttn-lw/lib/access-token'

const getResultActionFromType = function(typeString, status) {
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
 * as result action dispatch automatically.
 *
 * @param {object} options - The logic options (to be passed to `createLogic()`).
 * @param {(string|Function)} successType - The success action type or action creator.
 * @param {(string|Function)} failType - The fail action type or action creator.
 * @returns {object} The `redux-logic` (decorated) logic.
 */
const createRequestLogic = function(
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
    async process(deps, dispatch, done) {
      const promiseAttached = deps.action.meta && deps.action.meta._attachPromise

      try {
        const res = await options.process(deps, dispatch)

        // After successful request, dispatch success action.
        dispatch(successAction(res))

        // If we have a promise attached, resolve it.
        if (promiseAttached) {
          const {
            meta: { _resolve },
          } = deps.action
          _resolve(res)
        }
      } catch (e) {
        // Log the error when in development mode
        error(e)

        if (isUnauthenticatedError(e)) {
          // If there was an unauthenticated error, the access token is not
          // valid and we can delete it. Reloading will then initiate the auth
          // flow.
          clearAccessToken()
          window.location.reload()
        } else {
          // Otherwise, dispatch the fail action and report it to Sentry.
          if (isInvalidArgumentError(e)) {
            Sentry.withScope(scope => {
              scope.setExtras(e)
              const fingerprint = getBackendErrorId(e)
              scope.setFingerprint(fingerprint)
              Sentry.captureException(new Error(fingerprint))
            })
          } else if (isUnknown(e)) {
            Sentry.withScope(scope => {
              scope.setExtras(e)
              Sentry.captureException(e)
            })
          }
          dispatch(failAction(e))
        }

        // If we have a promise attached, reject it.
        if (promiseAttached) {
          const {
            meta: { _reject },
          } = deps.action
          _reject(e)
        }
      }

      done()
    },
  })
}

export default createRequestLogic
