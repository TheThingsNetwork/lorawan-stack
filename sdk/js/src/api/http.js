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
import { STACK_COMPONENTS } from '../util/constants'

/**
 * Http Class is a connector for the API that uses the HTTP bridge to connect.
 */
class Http {
  constructor (token, stackConfig, axiosConfig = {}) {
    const headers = axiosConfig.headers || {}

    let Authorization = null
    if (typeof token === 'string') {
      Authorization = `Bearer ${token}`
    }

    const stackComponents = Object.keys(stackConfig)
    const instances = stackComponents.reduce(function (acc, curr) {
      acc[curr] = axios.create({
        baseURL: stackConfig[curr],
        headers: {
          Authorization,
          ...headers,
        },
        ...axiosConfig,
      })

      return acc
    }, {})

    for (const instance in instances) {
      this[instance] = instances[instance]

      // Re-evaluate headers on each request if token is a thunk. This can be
      // useful if the token needs to be refreshed frequently, as the case for
      // access tokens.
      if (typeof token === 'function') {
        this[instance].interceptors.request.use(async function (config) {
          const tkn = (await token()).access_token
          config.headers.Authorization = `Bearer ${tkn}`

          return config
        },
        err => Promise.reject(err))
      }
    }
  }

  async handleRequest (method, endpoint, component, payload = {}, config) {
    const parsedComponent = component || this._parseStackComponent(endpoint)
    try {
      return await this[parsedComponent]({
        method,
        url: endpoint,
        data: payload,
        ...config,
      })
    } catch (err) {
      if ('response' in err && err.response && 'data' in err.response) {
        throw err.response.data
      } else {
        throw err
      }
    }
  }

  async get (endpoint, component, params) {
    // Convert payload to query params (should usually be field_mask only)
    const config = {}
    if (params && Object.keys(params).length > 0) {
      if ('field_mask' in params) {
        // Convert field mask prop to a query param friendly format
        params.field_mask = params.field_mask.paths.join(',')
      }
      config.params = params
    }

    return this.handleRequest('get', endpoint, component, undefined, config)
  }

  async post (endpoint, component, payload) {
    return this.handleRequest('post', endpoint, component, payload)
  }

  async patch (endpoint, component, payload) {
    return this.handleRequest('patch', endpoint, component, payload)
  }

  async put (endpoint, component, payload) {
    return this.handleRequest('put', endpoint, component, payload)
  }

  async delete (endpoint, component) {
    return this.handleRequest('delete', component, endpoint)
  }

  /**
   *  Extracts the stack component abbreviation from the endpoint.
   * @param {string} endpoint - The endpoint got for a request method.
   * @returns {string} One of {is|as|gs|js|ns}.
   */
  _parseStackComponent (endpoint) {
    try {
      const component = endpoint.split('/')[1]
      return STACK_COMPONENTS.includes(component) ? component : 'is'
    } catch (err) {
      throw new Error('Unable to extract the stack component:', endpoint)
    }
  }
}

export default Http
