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

import tts from '@console/api/tts'

import { clear as clearAccessToken } from '@ttn-lw/lib/access-token'
import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'
import * as init from '@ttn-lw/lib/store/actions/init'
import { TokenError } from '@ttn-lw/lib/errors/custom-errors'
import { isPermissionDeniedError, isUnauthenticatedError } from '@ttn-lw/lib/errors/utils'
import { selectPageStatusBaseUrlConfig } from '@ttn-lw/lib/selectors/env'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { getNetworkStatusSummary } from '@ttn-lw/lib/store/actions/status'

import { getActiveUserSessionIdSuccess } from '@console/store/actions/sessions'
import * as user from '@console/store/actions/user'
import { getInboxNotifications } from '@console/store/actions/notifications'
import { getAllBookmarks } from '@console/store/actions/user-preferences'

const consoleAppLogic = createRequestLogic({
  type: init.INITIALIZE,
  noCancelOnRouteChange: true,
  process: async (_, dispatch) => {
    dispatch(user.getUserRights())

    let info, rights, sessionId

    try {
      // There is no way to retrieve the current user directly within the
      // console app, so first get the authentication info and only afterwards
      // fetch the user.
      info = await tts.Auth.getAuthInfo()
      sessionId = info.oauth_access_token.user_session_id
      rights = info.oauth_access_token.rights
      dispatch(getActiveUserSessionIdSuccess(sessionId))
      dispatch(user.getUserRightsSuccess(rights))
    } catch (error) {
      if (
        error instanceof TokenError
          ? !isUnauthenticatedError(error?.cause) && !isPermissionDeniedError(error?.cause)
          : !isUnauthenticatedError(error)
      ) {
        throw error
      }

      // Clear existing access token since it does
      // not appear to be valid anymore.
      clearAccessToken()
      dispatch(user.getUserRightsFailure())
      info = undefined
    }

    if (info) {
      try {
        const userId = info.oauth_access_token.user_ids.user_id
        const statusPageUrl = selectPageStatusBaseUrlConfig()
        dispatch(user.getUserMe())
        dispatch(user.applyPersistedState(userId))
        const userResult = await tts.Users.getById(userId, [
          'state',
          'name',
          'primary_email_address',
          'primary_email_address_validated_at',
          'profile_picture',
          'console_preferences',
          'email_notification_preferences',
        ])
        userResult.isAdmin = info.is_admin || false
        dispatch(user.getUserMeSuccess(userResult))

        // Gather the initial actions to be dispatched, so they can be run in parallel.
        const initActions = []

        initActions.push(
          await dispatch(attachPromise(getInboxNotifications({ page: 1, limit: 3 }))),
          await dispatch(attachPromise(getAllBookmarks(userId))),
          statusPageUrl ? await dispatch(attachPromise(getNetworkStatusSummary())) : undefined,
        )

        await Promise.all(initActions)
      } catch (error) {
        dispatch(user.getUserMeFailure(error))
      }
    }

    // eslint-disable-next-line no-console
    console.log('Console initialization successful!')

    return
  },
})

export default [consoleAppLogic]
