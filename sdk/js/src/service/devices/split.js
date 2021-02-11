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

import Marshaler from '../../util/marshaler'
import deviceEntityMap from '../../../generated/device-entity-map.json'
import { STACK_COMPONENTS_MAP } from '../../util/constants'

const { is: IS, ns: NS, as: AS, js: JS } = STACK_COMPONENTS_MAP

/**
 * Takes the requested paths of the device and returns a request tree. The
 * splitting is achieved by looking up path responsibilities as defined in the
 * generated device entity map json.
 *
 * @param {object} paths - The requested paths (from the field mask) of the
 * device.
 * @param {string} direction - The direction, either 'set' or 'get'.
 * @param {object} base - An optional base value for the returned request tree.
 * @param {object} components - A component whitelist, unincluded components
 * will be excluded from the request tree.
 * @returns {object} A request tree object, consisting of resulting paths for
 * each component eg: `{ is: ['ids'], as: ['session'], js: ['root_keys'] }`.
 */
function splitPaths(paths = [], direction, base = {}, components = [IS, NS, AS, JS]) {
  const result = base
  const retrieveIndex = direction === 'get' ? 0 : 1

  for (const path of paths) {
    // Look up the current path in the device entity map.
    const subtree = traverse(deviceEntityMap).get(path) || traverse(deviceEntityMap).get([path[0]])

    if (!subtree) {
      throw new Error(`Invalid or unknown field mask path used: ${path}`)
    }

    const definition = '_root' in subtree ? subtree._root[retrieveIndex] : subtree[retrieveIndex]

    const map = function (requestTree, component, path) {
      if (components.includes(component)) {
        result[component] = !result[component] ? [path] : [...result[component], path]
      }
    }

    if (definition) {
      if (definition instanceof Array) {
        for (const component of definition) {
          map(result, component, path)
        }
      } else {
        map(result, definition, path)
      }
    }
  }
  return result
}

/**
 * A wrapper function to obtain a request tree for writing values to a device.
 *
 * @param {object} paths - The requested paths (from the field mask) of the
 * device.
 * @param {object} base - An optional base value for the returned request tree.
 * @param {object} components - A component whitelist, unincluded components
 * will be excluded from the request tree.
 * @returns {object} A request tree object, consisting of resulting paths for
 * each component eg: `{ is: ['ids'], as: ['session'], js: ['root_keys'] }`.
 */
export function splitSetPaths(paths, base, components) {
  return splitPaths(paths, 'set', base, components)
}

/**
 * A wrapper function to obtain a request tree for reading values to a device.
 *
 * @param {object} paths - The requested paths (from the field mask) of the
 * device.
 * @param {object} base - An optional base value for the returned request tree.
 * @param {object} components - A component whitelist, unincluded components
 * will be excluded from the request tree.
 * @returns {object} A request tree object, consisting of resulting paths for
 * each component eg: `{ is: ['ids'], as: ['session'], js: ['root_keys'] }`.
 */
export function splitGetPaths(paths, base, components) {
  return splitPaths(paths, 'get', base, components)
}

/**
 * `makeRequests` will make the necessary api calls based on the request tree
 * and other options.
 *
 * @param {object} api - The Api object as passed to the service.
 * @param {object} stackConfig - The Things Stack config object.
 * @param {string} operation - The operation, an enum of 'create', 'set', 'get'
 * and 'delete'.
 * @param {string} requestTree - The request tree, as returned by the
 * `splitPaths` function.
 * @param {object} params - The parameters object to be passed to the requests.
 * @param {object} payload - The payload to be passed to the requests.
 * @param {boolean} ignoreNotFound - Flag indicating whether not found errors
 * should be translated to an empty device instead of throwing.
 * @returns {object} An array of device registry responses together with the
 * paths (field_mask) that they were requested with.
 */
export async function makeRequests(
  api,
  stackConfig,
  operation,
  requestTree,
  params,
  payload = {},
  ignoreNotFound = false,
) {
  const isCreate = operation === 'create'
  const isSet = operation === 'set'
  const isDelete = operation === 'delete'
  const rpcFunction = isSet || isCreate ? 'Set' : isDelete ? 'Delete' : 'Get'

  // Use a wrapper for the api calls to control the result object and allow
  // ignoring not found errors per component, if wished.
  const requestWrapper = async function (
    call,
    params,
    component,
    payload,
    ignoreRequestNotFound = ignoreNotFound,
  ) {
    const res = { hasAttempted: true, component, paths: requestTree[component], hasErrored: false }
    try {
      const result = await call(params, !isDelete ? payload : undefined)
      return { ...res, device: Marshaler.payloadSingleResponse(result) }
    } catch (error) {
      if (error.code === 5 && ignoreRequestNotFound) {
        return { ...res, device: {} }
      }

      return { ...res, hasErrored: true, error }
    }
  }

  // Split end device payload per stack component.
  function splitPayload(payload = {}, paths, base = {}) {
    if (!Boolean(payload.end_device)) {
      return payload
    }

    const { end_device } = payload

    const result = traverse(base)
    const endDevice = traverse(end_device)

    for (const path of paths) {
      result.set(path, endDevice.get(path))
    }

    return Marshaler.payload(result.value, 'end_device')
  }

  const requests = new Array(3)

  if (isSet && !('end_device.ids.device_id' in params.routeParams)) {
    // Ensure using the PUT method by setting the device id route param. This
    // ensures upserting without issues.
    const { end_device } = payload
    const { ids: { device_id } = {} } = end_device
    if (device_id) {
      params.routeParams['end_device.ids.device_id'] = device_id
    }
  }

  const result = [
    { component: NS, hasAttempted: false, hasErrored: false },
    { component: AS, hasAttempted: false, hasErrored: false },
    { component: JS, hasAttempted: false, hasErrored: false },
    { component: IS, hasAttempted: false, hasErrored: false },
  ]

  const { end_device = {} } = payload

  // Do a possible IS request first.
  if (stackConfig.isComponentAvailable(IS) && IS in requestTree) {
    let func
    if (isSet) {
      func = 'Update'
    } else if (isCreate) {
      func = 'Create'
    } else if (isDelete) {
      func = 'Delete'
    } else {
      func = 'Get'
    }

    result[3] = await requestWrapper(
      api.EndDeviceRegistry[func],
      params,
      IS,
      {
        ...splitPayload(payload, requestTree.is, { ids: end_device.ids }),
        ...Marshaler.pathsToFieldMask(requestTree.is),
      },
      false,
    )

    if (isCreate) {
      // Abort and return the result object when the IS create request has
      // failed.
      if (result[3].hasErrored) {
        return result
      }
      // Set the device id param based on the id of the newly created device.
      params.routeParams['end_device.ids.device_id'] = result[3].device.ids.device_id
    }
  }

  // Compose an array of possible api calls to NS, AS, JS.
  if (stackConfig.isComponentAvailable(NS) && NS in requestTree) {
    requests[0] = requestWrapper(api.NsEndDeviceRegistry[rpcFunction], params, NS, {
      ...splitPayload(payload, requestTree[NS]),
      ...Marshaler.pathsToFieldMask(requestTree[NS]),
    })
  }
  if (stackConfig.isComponentAvailable(AS) && AS in requestTree) {
    requests[1] = requestWrapper(api.AsEndDeviceRegistry[rpcFunction], params, AS, {
      ...splitPayload(payload, requestTree[AS]),
      ...Marshaler.pathsToFieldMask(requestTree[AS]),
    })
  }
  if (stackConfig.isComponentAvailable(JS) && JS in requestTree) {
    requests[2] = requestWrapper(api.JsEndDeviceRegistry[rpcFunction], params, JS, {
      ...splitPayload(payload, requestTree[JS], { ids: end_device.ids }),
      ...Marshaler.pathsToFieldMask(requestTree[JS]),
    })
  }

  // Run the requests in parallel.
  const responses = await Promise.all(requests)

  // Attach the results to the result array.
  for (const [i, response] of responses.entries()) {
    if (response) {
      result[i] = response
    }
  }

  return result
}
