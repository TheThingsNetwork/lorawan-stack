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

import api from '../api'
import * as cache from './cache'

export default async function() {
  const storedToken = cache.get('accessToken')
  let token

  if (!storedToken || Date.parse(storedToken.expiry) < Date.now()) {
    // If we don't have a token stored or it's expired, we want to retrieve it
    const response = await api.console.token()
    token = response.data
  } else {
    // If we have a stored token and its valid, we want to use it
    return storedToken
  }

  // We want to make sure the stored token is the correct one
  if (!storedToken || storedToken.access_token !== token.access_token) {
    cache.set('accessToken', token)
  }

  return token
}

export function clear() {
  cache.remove('accessToken')
}
