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

/* eslint-disable no-invalid-this */
/* eslint-disable array-callback-return */

import traverse from 'traverse'
import queryString from 'query-string'

/** Class used to marshal data shapes. Currently a stub. */
class Marshaler {
  static options(options) {
    if (Object.keys(options).length === 0) {
      return null
    }

    const query = {}

    if ('select' in options) {
      query.field_mask = {}
      query.field_mask.paths = options.select
    }

    return query
  }

  static query(params) {
    return queryString.stringify(params)
  }

  static payload(payload, wrap) {
    let res = payload

    if (wrap) {
      res = { [wrap]: res }
    }

    return res
  }

  static payloadListResponse(entity, { data = {}, headers = {} }, transform) {
    const list = data[entity]

    if (!list) {
      return { [entity]: [], totalCount: 0 }
    }

    const totalCount = parseInt(headers['x-total-count'])

    if (isNaN(totalCount)) {
      return { [entity]: list, totalCount: list.length }
    }

    const transformedList = transform ? list.map(transform) : list

    return { [entity]: transformedList, totalCount }
  }

  static payloadSingleResponse(response, transform) {
    if (typeof response !== 'object') {
      throw new Error(`Invalid response type: ${typeof response}`)
    }
    if ('status' in response && response.status > 400) {
      throw new Error(`Response status ${response.status}`)
    }

    const entity = response.data || response
    return transform ? transform(entity) : entity
  }

  static unwrapRights(result, transform) {
    return this.payloadListResponse('rights', result, transform)
  }

  static unwrapApplications(result, transform) {
    return this.payloadListResponse('applications', result, transform)
  }

  static unwrapApplication(result, transform) {
    return this.payloadSingleResponse(result, transform)
  }

  static unwrapDevices(result, transform) {
    return this.payloadListResponse('end_devices', result, transform)
  }

  static unwrapDevice(result, transform) {
    return this.payloadSingleResponse(result, transform)
  }

  static unwrapGateways(result, transform) {
    return this.payloadListResponse('gateways', result, transform)
  }

  static unwrapGateway(result, transform) {
    return this.payloadSingleResponse(result, transform)
  }

  static unwrapUser(result, transform) {
    return this.payloadSingleResponse(result, transform)
  }

  static fieldMaskFromPatch(patch, whitelist, remaps) {
    let paths = []

    traverse(patch).map(function(x) {
      if (this.node instanceof Array) {
        // Add only the top level array path and do not recurse into arrays.
        paths.push(this.path.join('.'))
        this.update(undefined, true)
      } else if (this.isLeaf) {
        paths.push(this.path.join('.'))
      }
    })

    // Field masks can sometimes be arbitrarily mapped to the actual message
    // structure (e.g. for oneoffs). Through the remap argument, it can be
    // accounted for that by remapping these paths.
    if (remaps) {
      paths = paths.map(function(path) {
        for (const remap of remaps) {
          if (path.startsWith(remap[0])) {
            return path.replace(remap[0], remap[1])
          }
        }
        return path
      })
    }

    // If we have a whitelist provided, add paths only in the depth that the
    // whitelist allows and strip all other paths.
    if (whitelist) {
      paths = whitelist.reduce((acc, e) => {
        if (paths.some(path => path.startsWith(e))) {
          acc.push(e)
        }
        return acc
      }, [])
    }

    return paths
  }

  /** This function will convert a paths object to a proper field mask.
   * @param {Object} paths - The raw field mask as array and/or string.
   * @returns {Object} The field mask object ready to be attached to a request.
   */
  static pathsToFieldMask(paths) {
    if (!paths) {
      return
    }
    return { field_mask: { paths: paths.map(e => e.join('.')) } }
  }

  /** This function will convert a selector parameter and convert it to a
   * streamlined array of paths.
   * @param {Object} selector - The raw selector passed by the user
   * @returns {Object} The field mask object ready to be attached to a request.
   */
  static selectorToPaths(selector) {
    if (typeof selector === 'string') {
      return selector.split(',').map(e => e.split('.'))
    }
    if (selector instanceof Array) {
      return selector.map(e => (typeof e === 'string' ? e.split('.') : e))
    }
    return selector
  }

  /** This function will convert a selector parameter and convert it to a
   * proper field mask object, ready to be passed to the API.
   * @param {Object} selector - The raw selector passed by the user
   * @returns {Object} The field mask object ready to be attached to a request.
   */
  static selectorToFieldMask(selector) {
    return this.pathsToFieldMask(this.selectorToPaths(selector))
  }

  static fieldMask(fieldMask) {
    return { paths: fieldMask }
  }

  static queryFieldMask(fields = []) {
    return { 'field_mask.paths': fields }
  }
}

export default Marshaler
