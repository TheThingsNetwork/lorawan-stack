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

import apiDefinition from '../../generated/api-definition.json'
import { STACK_COMPONENTS } from '../util/constants'
import Http from './http'

/**
 * Api Class is an abstraction on the API connection which can use either the
 * HTTP or gRPC connector to communicate with TTN Stack for LoraWAN API in order
 * to expose the same class API for both
 */
class Api {
  constructor (connectionType = 'http', stackConfig, axiosConfig, token) {
    this.connectionType = connectionType

    if (this.connectionType !== 'http') {
      throw new Error('Only http connection type is supported')
    }

    this._connector = new Http(token, stackConfig, axiosConfig)
    const connector = this._connector

    for (const serviceName of Object.keys(apiDefinition)) {
      const service = apiDefinition[serviceName]

      this[serviceName] = {}

      for (const rpcName of Object.keys(service)) {
        const rpc = service[rpcName]

        this[serviceName][rpcName] = function ({ routeParams = {}, component } = {}, payload) {

          const componentType = typeof component
          if (componentType === 'string' && !STACK_COMPONENTS.includes(component)) {
            throw new Error(`Unknown stack component: ${component}`)
          }
          if (component && componentType !== 'string') {
            throw new Error(`Invalid component argument type: ${typeof componentType}`)
          }

          const paramSignature = Object.keys(routeParams).sort().join()
          const endpoint = rpc.http.find(function (prospect) {
            return prospect.parameters.sort().join() === paramSignature
          })

          if (!endpoint) {
            throw new Error(`The parameter signature did not match the one of the rpc.
Rpc: ${serviceName}.${rpcName}()
Signature tried: ${paramSignature}`)
          }

          let route = endpoint.pattern
          const isStream = Boolean(endpoint.stream)

          for (const parameter of endpoint.parameters) {
            route = route.replace(`{${parameter}}`, routeParams[parameter])
          }

          return connector.handleRequest(endpoint.method, route, component, payload, isStream)
        }

        this[serviceName][`${rpcName}AllowedFieldMaskPaths`] = rpc.allowedFieldMaskPaths
      }
    }
  }
}

export default Api
