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

import * as user from '../../actions/user'
import api from '../../../api'
import * as accessToken from '../../../lib/access-token'
import { isUnauthenticatedError } from '../../../../lib/errors/utils'
import createRequestLogic from './lib'

export default [
  createRequestLogic({
    type: user.LOGOUT,
    async process() {
      try {
        // Logout of the Console first
        await api.console.logout()

        try {
          // Then logout of the OAuth Provider
          await api.oauth.logout()
        } catch (err) {
          if (!isUnauthenticatedError(err)) {
            // Unauthenticated errors indicate that the user is already logged
            // out from the OAuth provider. We can hence ignore this error.
            throw err
          }
        }
      } finally {
        accessToken.clear()
      }
      return true
    },
  }),
]
