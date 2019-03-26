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

import traverse from 'traverse'
import Marshaler from '../util/marshaler'
import Device from '../entity/device'
import deviceEntityMap from '../../generated/device-entity-map.json'

/**
 * Devices Class provides an abstraction on all devices and manages data
 * handling from different sources. It exposes an API to easily work with
 * device data.
 */
class Devices {
  constructor (api, { applicationId, proxy = true }) {
    this._api = api
    this._applicationId = applicationId
    this._idMask = { route: { 'end_device.ids.application_ids.application_id': this._applicationId }}
    this._entityTransform = proxy
      ? app => new Device(this, app, false)
      : undefined
  }

  _splitEntitySetPaths (paths, base = {}) {
    return this._splitEntityPaths(paths, 'set', base)
  }

  _splitEntityGetPaths (paths, base = {}) {
    return this._splitEntityPaths(paths, 'get', base)
  }

  _splitEntityPaths (paths, direction, base = {}) {
    const result = base
    const retrieveIndex = direction === 'get' ? 0 : 1

    for (const path of paths) {
      const subtree =
        traverse(deviceEntityMap).get(path)
        || traverse(deviceEntityMap).get([ path[0] ])

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

  _mergeEntity (parts, base = {}) {
    const result = base

    for (const part of parts) {
      for (const path of part.paths) {
        traverse(result).set(path, traverse(part.record).get(path))
      }
    }

    return result
  }

  async getById (deviceId) {
    const res = await this._api.EndDeviceRegistry.Get({
      ...this._idMask,
      device_id: deviceId,
    })

    return new Device(res, this)
  }

  async updateById (deviceId) {
    return this._api.EndDeviceRegistry.Get({
      ...this._idMask,
      device_id: deviceId,
    })
  }

  async create (device, applicationId = this._applicationId) {
    const paths = traverse(device).reduce(function (acc, node) {
      if (this.isLeaf) {
        acc.push(this.path)
      }
      return acc
    }, [])

    const requestTree = this._splitEntitySetPaths(paths, {
      ns: [[ 'ids' ], [ 'created_at' ], [ 'updated_at' ]],
      as: [[ 'ids' ], [ 'created_at' ], [ 'updated_at' ]],
      js: [[ 'ids' ], [ 'created_at' ], [ 'updated_at' ]],
    })

    const isResult = await this._api.EndDeviceRegistry.Create(
      { route: { 'end_device.ids.application_ids.application_id': applicationId }},
      { end_device: device }
    )

    let nsResult = {}
    let asResult = {}
    let jsResult = {}

    if ('ns' in requestTree) {
      nsResult = await this._api.NsDeviceRegistry.Set({
        route: {
          'device.ids.application_ids.application_id': applicationId,
        },
      },
      {
        device,
        field_mask: Marshaler.fieldMask(requestTree.ns),
      })
    }
    if ('as' in requestTree) {
      asResult = await this._api.AsDeviceRegistry.Set({
        route: {
          'device.ids.application_ids.application_id': applicationId,
        },
      },
      {
        device,
        field_mask: Marshaler.fieldMask(requestTree.as),
      })
    }
    if ('js' in requestTree) {
      jsResult = await this._api.JsDeviceRegistry.Set({
        route: {
          'device.ids.application_ids.application_id': applicationId,
        },
      },
      {
        device,
        field_mask: Marshaler.fieldMask(requestTree.js),
      })
    }

    const result = this._mergeEntity([
      { record: isResult.data, paths: requestTree.is },
      { record: nsResult.data, paths: requestTree.ns },
      { record: asResult.data, paths: requestTree.as },
      { record: jsResult.data, paths: requestTree.js },
    ])

    return Marshaler.unwrapDevice(
      result,
      this._entityTransform
    )
  }
}

export default Devices
