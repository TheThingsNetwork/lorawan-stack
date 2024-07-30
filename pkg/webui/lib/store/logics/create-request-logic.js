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

/* eslint-disable no-await-in-loop */

import { createLogic } from 'redux-logic'

import toast from '@ttn-lw/components/toast'

import {
  isUnauthenticatedError,
  isNetworkError,
  isTimeoutError,
  ingestError,
  isPermissionDeniedError,
  isConnectionFailureError,
} from '@ttn-lw/lib/errors/utils'
import { clear as clearAccessToken } from '@ttn-lw/lib/access-token'
import {
  setLoginStatus,
  setStatusChecking,
  ATTEMPT_RECONNECT,
} from '@ttn-lw/lib/store/actions/status'
import { promisifyDispatch } from '@ttn-lw/lib/store/middleware/request-promise-middleware'
import attachPromise, { getResultActionFromType } from '@ttn-lw/lib/store/actions/attach-promise'
import { selectIsCheckingStatus } from '@ttn-lw/lib/store/selectors/status'
import { TokenError } from '@ttn-lw/lib/errors/custom-errors'
import errorMessages from '@ttn-lw/lib/errors/error-messages'

let connectionChecking = null
let lastError

/**
 * Logic creator for request logics, it will handle promise resolution, as well
 * as result action dispatch automatically.
 *
 * @param {object} options - The logic options (to be passed to `createLogic()`).
 * @param {boolean} options.noCancelOnRouteChange - Flag to disable the decoration
 * of the `cancelType` option.
 * @param {string} abortType - The type of the abort action.
 * @returns {object} The `redux-logic` (decorated) logic.
 */
const createRequestLogic = (
  { noCancelOnRouteChange, ...options },
  abortType = getResultActionFromType(options.type, 'ABORT'),
) => {
  if (
    options.type === undefined ||
    (options.type instanceof Array && options.type.length === 0) ||
    (options.type instanceof Array &&
      options.type.some(type => typeof type !== 'string' || type.indexOf('_REQUEST') === -1))
  ) {
    throw new Error('Could not derive result actions from provided options')
  }
  return createLogic({
    ...options,
    cancelType: abortType,
    process: async (deps, dispatch, done) => {
      const { action, getState, cancelled$ } = deps
      const promiseAttached = action.meta && action.meta._attachPromise
      const promisifiedDispatch = promisifyDispatch(dispatch)
      let actionSubscription

      // Compose response actions.
      const successAction = payload => ({
        type: getResultActionFromType(action.type, 'SUCCESS'),
        payload,
      })
      const failAction = (error, originalAction) => ({
        type: getResultActionFromType(action.type, 'FAILURE'),
        error: true,
        payload: error,
        meta: { requestPayload: originalAction.payload },
      })
      const abortAction = () => ({ type: getResultActionFromType(action.type, 'ABORT') })

      if (!noCancelOnRouteChange) {
        // Subscribe to action observable to dispatch an abort action on route changes.
        // TODO: Reintroduce the route change detection after removing the `connected-react-router`.
      }

      let success = false
      let connectionCheckResult = null
      while (!success) {
        try {
          // Check if the internet connection is currently (deemed) interrupted.
          if (connectionChecking) {
            const cancellation = new Promise((resolve, reject) => {
              cancelled$.subscribe({ next: reject, complete: resolve })
            })
            try {
              // Wait until the connection has been re-established or the
              // request has been aborted, e.g. due to a route change.
              connectionCheckResult = await Promise.race([connectionChecking, cancellation])
            } catch (error) {
              // The request was cancelled, so we cancel
              // further request execution.
              break
            }

            connectionChecking = null

            // Avoid request loops when connection check detected no connection loss.
            // Rather bubble up the error so it can stop the request action and be
            // handled by consuming logic.
            if (connectionCheckResult?.falseAlert) {
              throw lastError
            }
          }

          // Run the logic process.
          const res = await options.process(deps, promisifiedDispatch)
          success = true

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
          lastError = e

          // If there was an unauthenticated error, the access token is not
          // valid and we can delete it. A "Log back in"-modal will then pop up.
          if (
            isUnauthenticatedError(e) ||
            (e instanceof TokenError && isPermissionDeniedError(e?.cause))
          ) {
            clearAccessToken()
            dispatch(setLoginStatus())
            break
            // If there was a network error, it could mean that the network
            // connection is currently interrupted. Setting the online state to
            // `checking` will trigger respective connection management logics.
          } else if (
            (isNetworkError(e) || isTimeoutError(e)) &&
            !connectionCheckResult?.falseAlert
          ) {
            // We only need to set the status and trigger the connection checks
            // if the online status was `online` previously.
            if (!selectIsCheckingStatus(getState()) && action.type !== ATTEMPT_RECONNECT) {
              // It is necessary to promisify the dispatch function explicitly again
              // even though we already have a middleware to do that since`redux-logic`
              // modifies the dispatch function to return the original input action.
              connectionChecking = promisifiedDispatch(attachPromise(setStatusChecking()))
            }

            // Trigger a retry once the app is back online.
            continue

            // If the connection failed, the backend is likely updating. We can then
            // abort the request and display a toast with info about the status page
            // and ask to try again later.
          } else if (isConnectionFailureError(e)) {
            toast({
              messageGroup: 'connection-failure',
              message: errorMessages.connectionFailure,
              preventConsecutive: true,
              type: toast.types.WARNING,
            })
            dispatch(abortAction())
            break
          }

          // Pass relevant errors to Sentry.
          ingestError(
            e,
            { ingestedBy: 'createReqestLogic', requestAction: action },
            { requestAction: action.type },
          )

          // Dispatch the failure action and reject the promise, if attached.
          dispatch(failAction(e, action))
          if (promiseAttached) {
            const {
              meta: { _reject },
            } = action
            _reject(e)
          }
          // Do not retry if there was a (fatal) error.
          break
        }
      }

      if (actionSubscription) {
        actionSubscription.unsubscribe()
      }

      done()
    },
  })
}

export default createRequestLogic
