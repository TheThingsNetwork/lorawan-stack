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
import { cloneDeep, isObject } from 'lodash'

import Token from '../util/token'
import EventHandler from '../util/events'
import {
  URI_PREFIX_STACK_COMPONENT_MAP,
  STACK_COMPONENTS_MAP,
  AUTHORIZATION_MODES,
  RATE_LIMIT_RETRIES,
} from '../util/constants'

import subscribeToHttpStream from './stream/subscribeToHttpStream'

/**
 * Http Class is a connector for the API that uses the HTTP bridge to connect.
 */
class Http {
  constructor(authorization, stackConfig, axiosConfig = {}) {
    if (typeof authorization !== 'object' || authorization === null) {
      throw new Error('No authorization settings provided')
    }
    let authToken
    let csrfToken
    if (authorization.mode === AUTHORIZATION_MODES.KEY) {
      if (typeof authorization.key !== 'string' && typeof authorization.key !== 'function') {
        throw new Error('No valid key provided for key authorization')
      }
      authToken = new Token(authorization.key).get()
    } else if (
      authorization.mode === AUTHORIZATION_MODES.SESSION &&
      typeof authorization.csrfToken === 'string'
    ) {
      csrfToken = authorization.csrfToken
    }

    this._stackConfig = stackConfig
    const stackComponents = stackConfig.availableComponents
    const instances = stackComponents.reduce((acc, curr) => {
      const componentUrl = stackConfig.getComponentUrlByName(curr)
      if (componentUrl) {
        acc[curr] = axios.create({
          ...axiosConfig,
          baseURL: componentUrl,
          headers: {
            ...(typeof authToken === 'string' ? { Authorization: `Bearer ${authToken}` } : {}),
            ...(Boolean(csrfToken) ? { 'X-CSRF-Token': csrfToken } : {}),
            ...(axiosConfig.headers || {}),
          },
        })
      }

      return acc
    }, {})

    for (const instance in instances) {
      this[instance] = instances[instance]

      // Re-evaluate headers on each request if token is a thunk. This can be
      // useful if the token needs to be refreshed frequently, as the case for
      // access tokens.
      if (typeof authToken === 'function') {
        this[instance].interceptors.request.use(
          async config => {
            const tkn = (await authToken()).access_token
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
      // request and throw an error instead.
      throw new Error(
        `Cannot run "${method.toUpperCase()} ${endpoint}" API call on disabled component: "${parsedComponent}"`,
      )
    }

    try {
      if (isStream) {
        const url = this._stackConfig.getComponentUrlByName(parsedComponent) + endpoint
        return subscribeToHttpStream(payload, url)
      }

      const config = {
        method,
        url: endpoint,
      }

      if (method === 'get' || method === 'delete') {
        // For GETs convert payload to query params (should usually
        // be field_mask only).
        config.params = this._payloadToQueryParams(payload)
      } else {
        // Otherwise pass data as request body.
        config.data = payload
      }

      let statusCode, response, retryAfter, limit
      let retries = 0

      while (statusCode === undefined || statusCode === 429) {
        if (statusCode === 429 && retryAfter !== undefined) {
          // Dispatch a warning event to note the user about the waiting time
          // resulting from the rate limitation.
          EventHandler.dispatchEvent(
            EventHandler.EVENTS.WARNING,
            `The rate limitation of ${limit} requests per minute was exceeded while making a request. It will be automatically retried when the rate limiter resets.`,
          )

          // Sleep until the cool down elapsed before retrying.
          // eslint-disable-next-line no-await-in-loop
          await new Promise(resolve => setTimeout(resolve, retryAfter * 1000))
        }

        try {
          // eslint-disable-next-line no-await-in-loop
          response = await this[parsedComponent](config)
          statusCode = response.status
        } catch (err) {
          if (
            isObject(err) &&
            'response' in err &&
            isObject(err.response) &&
            'status' in err.response &&
            err.response.status === 429 &&
            retries <= RATE_LIMIT_RETRIES
          ) {
            statusCode = 429
            // Always wait at least one second to avoid retries in quick succession.
            retryAfter = Math.max(1, parseInt(err.response.headers['x-rate-limit-retry']))
            limit = err.response.headers['x-rate-limit-limit']
          } else {
            throw err
          }
        }

        retries++
      }

      for (const key in response.headers) {
        if (!(key.toLowerCase() in response.headers)) {
          // Normalize capitalized HTTP/1 headers to lowercase HTTP/2 headers.
          response.headers[key.toLowerCase()] = response.headers[key]
        }
      }

      if ('x-warning' in response.headers) {
        // Dispatch a warning event when the server has set a warning header.
        EventHandler.dispatchEvent(EventHandler.EVENTS.WARNING, response.headers['x-warning'])
      }

      return response
    } catch (err) {
      if (isObject(err) && 'response' in err && err.response && 'data' in err.response) {
        const error = cloneDeep(err.response.data)

        throw error
      } else {
        throw err
      }
    }
  }

  /**
   * Converts a payload object to a query parameter object, making sure that the
   * field mask parameter is converted correctly.
   *
   * @param {object} payload - The payload object.
   * @returns {object} The params object, to be passed to axios config.
   */
  _payloadToQueryParams(payload) {
    const res = { ...payload }
    if (payload && Object.keys(payload).length > 0) {
      const { field_mask } = payload
      if (!field_mask) {
        return res
      }
      const { paths } = field_mask
      delete res.field_mask
      if (!Array.isArray(paths) || paths.length === 0) {
        return res
      }
      // Convert field mask prop to a query param friendly format
      res.field_mask = paths.join(',')
      return res
    }
    return {}
  }

  /**
   * Extracts The Things Stack component abbreviation from the endpoint.
   *
   * @param {string} endpoint - The endpoint got for a request method.
   * @returns {string} The stack component abbreviation.
   */
  _parseStackComponent(endpoint) {
    try {
      const component = endpoint.split('/')[1]
      return Boolean(URI_PREFIX_STACK_COMPONENT_MAP[component])
        ? URI_PREFIX_STACK_COMPONENT_MAP[component]
        : STACK_COMPONENTS_MAP.is
    } catch (err) {
      throw new Error('Unable to extract The Things Stack component:', endpoint)
    }
  }
}

export default Http
