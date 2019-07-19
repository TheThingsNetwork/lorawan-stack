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

import axios from 'axios'

import getCookieValue from '../../lib/cookie'

const csrf = getCookieValue('_csrf')
const instance = axios.create({
  headers: { 'X-CSRF-Token': csrf },
})

export default {
  users: {
    async register (userData) {
      return axios.post(`/api/v3/users`, userData)
    },
  },
  oauth: {
    login (credentials) {
      return instance.post('/oauth/api/auth/login', credentials)
    },
    logout () {
      return instance.post('/oauth/api/auth/logout')
    },
    me () {
      return instance.get('/oauth/api/me')
    },
    reset (user_id) {
      return instance.post('/oauth/api/auth/password-reset', user_id)
    },
  },
}
