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

import { isPlainObject } from 'lodash'

import { TokenError } from './errors/custom-errors'
import * as cache from './cache'

export default fetchToken => {
  let tokenPromise
  let finishedRetrieval = true

  const retrieveToken = async () => {
    try {
      const response = await fetchToken()
      const token = response.data
      if (!isPlainObject(token) || !('access_token' in token)) {
        throw new TokenError('Received invalid token')
      }

      cache.set('accessToken', token)

      return token
    } catch (error) {
      throw new TokenError('Could not fetch token', error)
    } finally {
      finishedRetrieval = true
    }
  }

  return () => {
    const token = cache.get('accessToken')

    if (!token || Date.parse(token.expiry) < Date.now()) {
      // If we don't have a token stored or it's expired, we want to retrieve it.

      // Prevent issuing more than one request at a time.
      if (finishedRetrieval) {
        finishedRetrieval = false

        // Remove stored, invalid token.
        clear()

        // Retrieve new token and store it.
        tokenPromise = retrieveToken()
      }
      return tokenPromise
    }

    // If we have a stored token and its valid, we want to use it.
    return token
  }
}

export const clear = () => {
  cache.remove('accessToken')
}
