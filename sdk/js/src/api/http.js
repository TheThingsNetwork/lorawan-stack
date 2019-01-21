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
import Marshaler from '../util/marshaler'

/**
 * Http Class is a connector for the API that uses the HTTP bridge to connect.
 */
class Http {
  constructor (token, { baseURL, axiosConfig = {}}) {
    const headers = axiosConfig.headers || {}
    let Authorization = null

    if (typeof token === 'string') {
      Authorization = `Bearer ${token}`
    }

    this.axios = axios.create({
      baseURL,
      headers: {
        Authorization,
        ...headers,
      },
      ...axiosConfig,
    })

    // Re-evaluate headers on each request if token is a thunk. This can be
    // useful if the token needs to be refreshed frequently, as the case for
    // access tokens.
    if (typeof token === 'function') {
      this.axios.interceptors.request.use(async function (config) {
        const tkn = (await token()).access_token
        config.headers.Authorization = `Bearer ${tkn}`

        return config
      },
      err => Promise.reject(err))
    }
  }

  async handleRequest (method, endpoint, rawPayload) {
    const payload = rawPayload ? Marshaler.payload(rawPayload) : {}
    try {
      const res = await this.axios[method](endpoint, payload)
      return res.data
    } catch (err) {
      if ('response' in err && err.response && 'data' in err.response) {
        throw err.response.data
      } else {
        throw err
      }
    }
  }

  async get (endpoint) {
    return this.handleRequest('get', endpoint)
  }

  async post (endpoint, rawPayload) {
    return this.handleRequest('post', endpoint, rawPayload)
  }

  async patch (endpoint, rawPayload) {
    return this.handleRequest('patch', endpoint, rawPayload)
  }

  async put (endpoint, rawPayload) {
    return this.handleRequest('put', endpoint, rawPayload)
  }

  async delete (endpoint) {
    return this.handleRequest('delete', endpoint)
  }
}

export default Http
