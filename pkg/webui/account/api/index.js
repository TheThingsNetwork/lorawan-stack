// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import { isPermissionDeniedError } from '@ttn-lw/lib/errors/utils'

const appRoot = selectApplicationRootPath()

const csrf = selectCSRFToken()
const instance = axios.create({
  headers: { 'X-CSRF-Token': csrf },
})

instance.interceptors.response.use(
  response => response,
  async error => {
    if (isPermissionDeniedError(error) && error.response.data.includes('CSRF')) {
      // If the CSRF token is invalid, it likely means that the CSRF cookie
      // has been deleted or became outdated. Making a new request to the
      // current path can then retrieve a fresh CSRF cookie, with which
      // the request can be retried.
      const csrfResult = await axios.get(window.location)
      const freshCsrf = csrfResult.headers['x-csrf-token']
      const freshRequest = {
        ...error.config,
        headers: { ...error.config.headers, 'X-CSRF-Token': freshCsrf },
      }
      return axios(freshRequest)
    }
    if ('response' in error && error.response && 'data' in error.response) {
      throw error.response.data
    }
    throw error
  },
)

export default {
  account: {
    login: credentials => instance.post(`${appRoot}/api/auth/login`, credentials),
    logout: () => instance.post(`${appRoot}/api/auth/logout`),
    me: () => instance.get(`${appRoot}/api/me`),
  },
}
