// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
import token from '../lib/access-token'

export default {
  v3: {
    is: {
      clients: {
        get (client_id) {
          return axios.get(`/api/v3/is/clients/${client_id}`)
        },
      },
      users: {
        async me () {
          return axios.get('/api/v3/is/users/me', {
            headers: {
              Authorization: `Bearer ${(await token()).access_token}`,
            },
          })
        },
      },
    },
  },
  console: {
    auth: {
      token () {
        return axios.get('/console/api/auth/token')
      },
      refresh () {
        return axios.put('/console/api/auth/refresh')
      },
      logout () {
        return axios.post('/console/api/auth/logout')
      },
    },
  },
  oauth: {
    login (credentials) {
      return axios.post('/oauth/api/auth/login', credentials)
    },
  },
}
