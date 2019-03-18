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

import traverse from 'traverse'
import deviceEntityMap from '../../../generated/device-entity-map.json'


/** Takes the requested paths of the device and returns a request tree. The
* splitting is achieved by looking up path responsibilities as defined in the
* generated device entity map json.
* @param {Object} paths - The requested paths (from the field mask) of the device
* @param {string} direction - The direction, either 'set' or 'get'
* @param {Object} base - An optional base value for the returned request tree
* @returns {Object} A request tree object, consisting of resulting paths for each
* component eg: { is: ['ids'], as: ['session'], js: ['root_keys'] }
*/
function splitPaths (paths = [], direction, base = {}) {
  const result = base
  const retrieveIndex = direction === 'get' ? 0 : 1

  for (const path of paths) {
    // Look up the current path in the device entity map
    const subtree =
      traverse(deviceEntityMap).get(path)
      || traverse(deviceEntityMap).get([ path[0] ])

    if (!subtree) {
      throw new Error(`Invalid or unknown field mask path used: ${path}`)
    }

    const definition = '_root' in subtree ? subtree._root[retrieveIndex] : subtree[retrieveIndex]

    if (definition) {
      if (definition instanceof Array) {
        for (const component of definition) {
          result[component] = !result[component] ? [ path ] : [ ...result[component], path ]
        }
      } else {
        result[definition] = !result[definition] ? [ path ] : [ ...result[definition], path ]
      }
    }
  }
  return result
}

/** A wrapper function to obtain a request tree for writing values to a device
* @param {Object} paths - The requested paths (from the field mask) of the device
* @param {Object} base - An optional base value for the returned request tree
* @returns {Object} A request tree object, consisting of resulting paths for each
* component eg: { is: ['ids'], as: ['session'], js: ['root_keys'] }
*/
export function splitSetPaths (paths, base) {
  return splitPaths(paths, 'set', base)
}

/** A wrapper function to obtain a request tree for reading values to a device
* @param {Object} paths - The requested paths (from the field mask) of the device
* @param {Object} base - An optional base value for the returned request tree
* @returns {Object} A request tree object, consisting of resulting paths for each
* component eg: { is: ['ids'], as: ['session'], js: ['root_keys'] }
*/
export function splitGetPaths (paths, base) {
  return splitPaths(paths, 'get', base)
}

