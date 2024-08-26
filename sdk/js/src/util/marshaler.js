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

/** Class used to marshal data shapes. */
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

  static payloadListResponse(entity, { data = {}, headers = {} }) {
    const list = data[entity]

    if (!list) {
      return { [entity]: [], totalCount: 0 }
    }

    const totalCount = parseInt(headers['x-total-count'])

    if (isNaN(totalCount)) {
      return { [entity]: list, totalCount: list.length }
    }

    return { [entity]: list, totalCount }
  }

  static payloadSingleResponse(response) {
    if (typeof response !== 'object') {
      throw new Error(`Invalid response type: ${typeof response}`)
    }
    if ('status' in response && response.status >= 400) {
      throw new Error(`Response status ${response.status}`)
    }

    const entity = response.data || response

    return entity
  }

  static unwrapRights(result) {
    return this.payloadListResponse('rights', result)
  }

  static unwrapApplications(result) {
    return this.payloadListResponse('applications', result)
  }

  static unwrapApplication(result) {
    return this.payloadSingleResponse(result)
  }

  static unwrapDevices(result) {
    return this.payloadListResponse('end_devices', result)
  }

  static unwrapDevice(result) {
    return this.payloadSingleResponse(result)
  }

  static unwrapGateways(result) {
    return this.payloadListResponse('gateways', result)
  }

  static unwrapGateway(result) {
    return this.payloadSingleResponse(result)
  }

  static unwrapUser(result) {
    return this.payloadSingleResponse(result)
  }

  static unwrapClients(result) {
    return this.payloadSingleResponse(result)
  }

  static unwrapPacketBrokerNetworks(result) {
    return this.payloadListResponse('networks', result)
  }

  static unwrapPacketBrokerPolicies(result) {
    return this.payloadListResponse('policies', result)
  }

  static unwrapInvitations(result) {
    return this.payloadListResponse('invitations', result)
  }

  static unwrapInvitation(result) {
    return this.payloadSingleResponse(result)
  }

  static unwrapBookmarks(result) {
    return this.payloadListResponse('bookmarks', result)
  }

  static unwrapBookmark(result) {
    return this.payloadSingleResponse(result)
  }

  static fieldMaskFromPatch(patch, whitelist, remaps = []) {
    const paths = []

    traverse(patch).map(function () {
      const isArray = this.node instanceof Array
      if (isArray) {
        // Add only the top level array path and do not recurse into arrays.
        this.update(undefined, true)
      }
      if (this.isLeaf || isArray) {
        let pathArray = this.path
        const pathString = pathArray.join('.')

        // Field masks can sometimes be arbitrarily mapped to the actual message
        // structure (e.g. for oneoffs). Through the remap argument, it can be
        // accounted for that by remapping these paths.
        for (const remap of remaps) {
          if (pathString.startsWith(remap[0])) {
            pathArray = pathString.replace(remap[0], remap[1]).split('.')
          }
        }

        // If we have a whitelist provided add paths only in the depth that the
        // whitelist allows and strip all other paths.
        if (whitelist) {
          // Only add the deepest possible path.
          for (let i = pathArray.length; i >= 0; i--) {
            const subPath = pathArray.slice(0, i).join('.')
            if (whitelist.includes(subPath) && !paths.includes(subPath)) {
              paths.push(subPath)
              break
            }
          }
        } else {
          paths.push(this.path.join('.'))
        }
      }
    })

    return paths
  }

  /**
   * This function will convert a paths object to a proper field mask.
   *
   * @param {object} paths - The raw field mask as array and/or string.
   * @returns {object} The field mask object ready to be attached to a request.
   */
  static pathsToFieldMask(paths) {
    if (!paths) {
      return null
    }
    return { field_mask: { paths: paths.map(e => e.join('.')) } }
  }

  /**
   * This function will convert a selector parameter and convert it to a
   * streamlined array of paths.
   *
   * @param {object} selector - The raw selector passed by the user.
   * @returns {object} The field mask object ready to be attached to a request.
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

  /**
   * This function will convert a selector parameter and convert it to a
   * proper field mask object, ready to be passed to the API.
   *
   * @param {object} selector - The raw selector passed by the user.
   * @returns {object} The field mask object ready to be attached to a request.
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
