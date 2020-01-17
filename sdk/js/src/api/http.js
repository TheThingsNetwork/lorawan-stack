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

import { URI_PREFIX_STACK_COMPONENT_MAP } from '../util/constants'
import EventHandler from '../util/events'
import stream from './stream/stream-node'

/**
 * Http Class is a connector for the API that uses the HTTP bridge to connect.
 */
class Http {
  constructor(token, stackConfig, axiosConfig = {}) {
    this._stackConfig = stackConfig
    const headers = axiosConfig.headers || {}

    let Authorization = null
    if (typeof token === 'string') {
      Authorization = `Bearer ${token}`
    }

    const stackComponents = stackConfig.availableComponents
    const instances = stackComponents.reduce(function(acc, curr) {
      const componentUrl = stackConfig.getComponentUrlByName(curr)
      if (componentUrl) {
        acc[curr] = axios.create({
          baseURL: componentUrl,
          headers: {
            Authorization,
            ...headers,
          },
          ...axiosConfig,
        })
      }

      return acc
    }, {})

    for (const instance in instances) {
      this[instance] = instances[instance]

      // Re-evaluate headers on each request if token is a thunk. This can be
      // useful if the token needs to be refreshed frequently, as the case for
      // access tokens.
      if (typeof token === 'function') {
        this[instance].interceptors.request.use(
          async function(config) {
            const tkn = (await token()).access_token
            config.headers.Authorization = `Bearer ${tkn}`

            return config
          },
          err => Promise.reject(err),
        )
      }
    }
  }

  async handleRequest(method, endpoint, component, payload = {}, isStream) {
    const parsedComponent = component || this._parseStackComponent(endpoint)
    if (!this._stackConfig.isComponentAvailable(parsedComponent)) {
      // If the component has not been defined in The Things Stack config, make no
      // request and throw an error instead
      throw new Error(
        `Cannot run "${method.toUpperCase()} ${endpoint}" API call on disabled component: "${parsedComponent}"`,
      )
    }

    try {
      if (isStream) {
        const url = this._stackConfig.getComponentUrlByName(parsedComponent) + endpoint
        return stream(payload, url)
      }

      const config = {
        method,
        url: endpoint,
      }

      if (method === 'get' || method === 'delete') {
        // For GETs and DELETEs, convert payload to query params (should usually
        // be field_mask only)
        config.params = this._payloadToQueryParams(payload)
      } else {
        // Otherwise pass data as request body
        config.data = payload
      }

      const response = await this[parsedComponent](config)

      if ('X-Warning' in response.headers || 'x-warning' in response.headers) {
        // Dispatch a warning event when the server has set a warning header
        EventHandler.dispatchEvent(
          EventHandler.EVENTS.WARNING,
          response.headers['X-Warning'] || response.headers['x-warning'],
        )
      }

      return response
    } catch (err) {
      if ('response' in err && err.response && 'data' in err.response) {
        throw err.response.data
      } else {
        throw err
      }
    }
  }

  /**
   * Converts a payload object to a query parameter object, making sure that the
   * field mask parameter is converted correctly.
   * @param {Object} payload - The payload object.
   * @returns {Object} The params object, to be passed to axios config.
   */
  _payloadToQueryParams(payload) {
    const res = { ...payload }
    if (payload && Object.keys(payload).length > 0) {
      if ('field_mask' in payload) {
        // Convert field mask prop to a query param friendly format
        res.field_mask = payload.field_mask.paths.join(',')
      }
      return res
    }
  }

  /**
   * Extracts The Things Stack component abbreviation from the endpoint.
   * @param {string} endpoint - The endpoint got for a request method.
   * @returns {string} One of {is|as|gs|js|ns}.
   */
  _parseStackComponent(endpoint) {
    try {
      const component = endpoint.split('/')[1]
      return Boolean(URI_PREFIX_STACK_COMPONENT_MAP[component])
        ? URI_PREFIX_STACK_COMPONENT_MAP[component]
        : 'is'
    } catch (err) {
      throw new Error('Unable to extract The Things Stack component:', endpoint)
    }
  }
}

export default Http
