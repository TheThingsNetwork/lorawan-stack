// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import axios from 'axios'

import { selectApplicationRootPath, selectCSRFToken } from '@ttn-lw/lib/selectors/env'
import tokenCreator from '@ttn-lw/lib/access-token'

const appRoot = selectApplicationRootPath()

const csrf = selectCSRFToken()

const token = tokenCreator(() => axios.get(`${appRoot}/api/auth/token`))

export default {
  console: {
    token: () => axios.get(`${appRoot}/api/auth/token`),
    logout: async () => {
      const headers = token => ({
        headers: { 'X-CSRF-Token': token },
      })
      try {
        return await axios.post(`${appRoot}/api/auth/logout`, undefined, headers(csrf))
      } catch (error) {
        if (
          error.response &&
          error.response.status === 403 &&
          typeof error.response.data === 'string' &&
          error.response.data.includes('CSRF')
        ) {
          // If the CSRF token is invalid, it likely means that the CSRF cookie
          // has been deleted or became outdated. Making a new request to the
          // current path can then retrieve a fresh CSRF cookie, with which
          // the logout can be retried.
          const csrfResult = await axios.get(window.location)
          const freshCsrf = csrfResult.headers['x-csrf-token']
          if (freshCsrf) {
            return axios.post(`${appRoot}/api/auth/logout`, undefined, headers(freshCsrf))
          }
        }

        throw error
      }
    },
  },
}

export { token }
