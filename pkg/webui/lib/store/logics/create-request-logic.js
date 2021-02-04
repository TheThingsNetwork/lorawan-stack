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
import { defineMessages } from 'react-intl'

import {
  isUnauthenticatedError,
  isNetworkError,
  isTimeoutError,
  createFrontendError,
  ingestError,
} from '@ttn-lw/lib/errors/utils'
import { clear as clearAccessToken } from '@ttn-lw/lib/access-token'
import {
  setStatusChecking,
  ATTEMPT_RECONNECT,
  attemptReconnect,
} from '@ttn-lw/lib/store/actions/status'
import { getResultActionFromType } from '@ttn-lw/lib/store/actions/attach-promise'
import { selectIsCheckingStatus, selectIsOnlineStatus } from '@ttn-lw/lib/store/selectors/status'
import { selectApplicationSiteTitle } from '@ttn-lw/lib/selectors/env'

const siteTitle = selectApplicationSiteTitle()

const m = defineMessages({
  applicationIsOfflineTitle: '{applicationName} is offline',
  applicationIsOfflineMessage:
    'The action cannot be performed because your host machine is currently offline or has connection issues. Please check your internet connection and try again.',
})

const offlineError = createFrontendError(
  { ...m.applicationIsOfflineTitle, values: { applicationName: siteTitle } },
  m.applicationIsOfflineMessage,
  'request_denied_application_is_offline',
)

/**
 * Logic creator for request logics, it will handle promise resolution, as well
 * as result action dispatch automatically.
 *
 * @param {object} options - The logic options (to be passed to `createLogic()`).
 * @param {(string|Function)} successType - The success action type or action creator.
 * @param {(string|Function)} failType - The fail action type or action creator.
 * @returns {object} The `redux-logic` (decorated) logic.
 */
const createRequestLogic = (
  options,
  successType = getResultActionFromType(options.type, 'SUCCESS'),
  failType = getResultActionFromType(options.type, 'FAILURE'),
) => {
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
    process: async (deps, dispatch, done) => {
      const { action, getState } = deps
      const promiseAttached = action.meta && action.meta._attachPromise

      if (!selectIsOnlineStatus(getState())) {
        // If the application is currently (deemed) offline, skip making the
        // (ill-fated) request and handle the request as failed.
        if (promiseAttached) {
          const {
            meta: { _reject },
          } = action
          _reject(offlineError)
        }

        dispatch(failAction(offlineError))

        // Additionally issue a reconnect attempt immediately.
        dispatch(attemptReconnect())

        return done()
      }

      try {
        const res = await options.process(deps, dispatch)

        // After successful request, dispatch success action.
        dispatch(successAction(res))

        // If we have a promise attached, resolve it.
        if (promiseAttached) {
          const {
            meta: { _resolve },
          } = action
          _resolve(res)
        }
      } catch (e) {
        ingestError(
          e,
          { ingestedBy: 'createReqestLogic', requestAction: action },
          { requestAction: action.type },
        )

        if (isUnauthenticatedError(e)) {
          // If there was an unauthenticated error, the access token is not
          // valid and we can delete it. Reloading will then initiate the auth
          // flow.
          clearAccessToken()
          window.location.reload()
        } else if (isNetworkError(e) || isTimeoutError(e)) {
          // If there was a network error, it could mean that the network
          // connection is currently interrupted. Setting the online state to
          // `checking` will trigger respective connection management logics.
          if (!selectIsCheckingStatus(getState()) && action.type !== ATTEMPT_RECONNECT) {
            // We only need to set the status and trigger the connection checks
            // if the online status was `online` previously.
            dispatch(setStatusChecking())
          }
        }

        // Dispatch the failure action and reject the promise, if attached.
        dispatch(failAction(e))
        if (promiseAttached) {
          const {
            meta: { _reject },
          } = action
          _reject(e)
        }
      }

      done()
    },
  })
}

export default createRequestLogic
