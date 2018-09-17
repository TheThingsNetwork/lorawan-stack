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

import apiDefinition from '../../generated/api-definition.json'
import Http from './http'

/**
 * Api Class is an abstraction on the API connection which can use either the
 * HTTP or gRPC connector to communicate with the TTN LoraWAN API in order to
 * expose the same class API for both
 */
class Api {
  constructor (connectionType, connectionConfig, token) {
    this.connectionType = connectionType
    this.connectionConfig = connectionConfig
    this.token = token

    this.connector = this.connectionType === 'http' ? new Http(token, connectionConfig) : null
    for (const rpcName of Object.keys(apiDefinition)) {
      const rpc = apiDefinition[rpcName]
      this[rpcName] = function (params = {}, body) {
        if (this.connectionType === 'http') {
          if (!rpc.http) {
            throw new Error('gRPC connector is not implemented yet.')
          }

          const paramSignature = Object.keys(params).sort().join()

          const endpoint = rpc.http.find(function (prospect) {
            return prospect.parameters.sort().join() === paramSignature
          })

          if (!endpoint) {
            throw new Error('The parameter signature did not match the one of the rpc.')
          }

          let route = endpoint.pattern

          for (const parameter of endpoint.parameters) {
            route = route.replace(`{${parameter}}`, params[parameter])
          }

          return this.connector[endpoint.method](route, body)
        }
      }
    }
  }
}

export default Api
