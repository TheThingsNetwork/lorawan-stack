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
import randomByteString from '../util/random-bytes'

/**
 * Devices Class provides an abstraction on all devices and manages data
 * handling from different sources. It exposes an API to easily work with
 * device data.
 */
class Devices {
  constructor (api, { proxy = true, stackConfig }) {
    if (!api) {
      throw new Error('Cannot initialize device service without api object.')
    }
    this._api = api
    this._stackConfig = stackConfig
  }

  _responseTransform (response) {
    const isList = response instanceof Array
    return Marshaler[isList ? 'unwrapDevices' : 'unwrapDevice'](
      response,
      this._proxy
        ? app => new Device(this, app, false)
        : undefined
    )
  }

  _splitEntitySetPaths (paths, base) {
    return this._splitEntityPaths(paths, 'set', base)
  }

  _splitEntityGetPaths (paths, base) {
    return this._splitEntityPaths(paths, 'get', base)
  }

  _splitEntityPaths (paths = [], direction, base = {}) {
    const result = base
    const retrieveIndex = direction === 'get' ? 0 : 1

    for (const path of paths) {
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

  _mergeEntity (
    parts,
    base = {},
    minimum = [[ 'ids' ], [ 'created_at' ], [ 'updated_at' ]]
  ) {
    const result = base

    for (const part of parts) {
      for (const path of part.paths ? [ ...minimum, ...part.paths ] : []) {
        const val = traverse(part.record).get(path)
        if (val) {
          if (typeof val === 'object') {
            // In case of a whole sub-object being selected, write each leaf node
            // explicitly to achieve a deep merge instead of object overrides
            if (Object.keys(val).length === 0) {
              // Ignore empty object values
              continue
            }
            traverse(val).forEach(function (e) {
              if (this.isLeaf) {
                if (typeof e === 'object' && Object.keys(e).length === 0) {
                  // Ignore empty object values
                  return
                }
                traverse(result).set([ ...path, ...this.path ], e)
              }
            })
          } else {
            traverse(result).set(path, val)
          }
        }
      }
    }

    return result
  }

  async _setDevice (applicationId, device, create = false) {
    const ids = device.ids
    const deviceId = 'device_id' in ids && ids.device_id
    const appId = applicationId || 'application_ids' in ids && ids.application_ids.application_id

    if (!create && !deviceId) {
      throw new Error('Missing device_id for update operation.')
    }

    if (!appId) {
      throw new Error('Missing application_id for device.')
    }

    const mergeBase = create ? {
      ns: [[ 'ids' ]],
      as: [[ 'ids' ]],
      js: [[ 'ids' ]],
    } : {}

    const params = {
      routeParams: { 'end_device.ids.application_ids.application_id': appId },
    }

    // Extract the paths from the patch
    const paths = traverse(device).reduce(function (acc, node) {
      if (this.isLeaf) {
        acc.push(this.path)
      }
      return acc
    }, [])

    const requestTree = this._splitEntitySetPaths(paths, mergeBase)

    let isResult = {}
    if (create) {
      isResult = await this._api.EndDeviceRegistry.Create(params, device)
    }

    params.routeParams['end_device.ids.device_id'] = 'data' in isResult ? isResult.data.ids.device_id : deviceId

    const requests = new Array(3)

    if (!create && 'is' in requestTree) {
      isResult = await this._api.EndDeviceRegistry.Update({
        ...params,
        ...Marshaler.pathsToFieldMask(requestTree.is),
      }, device)
    }

    if ('ns' in requestTree) {
      requests[0] = this._api.NsEndDeviceRegistry.Set({
        ...params,
        ...Marshaler.pathsToFieldMask(requestTree.ns),
      }, device)
    }
    if ('as' in requestTree) {
      requests[1] = this._api.AsEndDeviceRegistry.Set({
        ...params,
        ...Marshaler.pathsToFieldMask(requestTree.as),
      }, device)
    }
    if ('js' in requestTree) {
      requests[2] = this._api.JsEndDeviceRegistry.Set({
        ...params,
        ...Marshaler.pathsToFieldMask(requestTree.js),
      }, device)
    }

    try {
      const setResults = (await Promise.all(requests))
        .map(e => e ? Marshaler.payloadSingleResponse(e) : undefined)

      const result = this._mergeEntity([
        { record: setResults[0], paths: requestTree.ns },
        { record: setResults[1], paths: requestTree.as },
        { record: setResults[2], paths: requestTree.js },
        { record: Marshaler.payloadSingleResponse(isResult), paths: requestTree.is },
      ])

      return result
    } catch (err) {
      // Roll back changes
      this._deleteDevice(appId, deviceId, Object.keys(requestTree))
      throw new Error('Could not create device.')
    }
  }

  async _getDevice (applicationId, deviceId, paths) {

    if (!applicationId) {
      throw new Error('Missing application_id for device.')
    }

    const requestTree = this._splitEntityGetPaths(paths)

    const params = {
      routeParams: {
        'end_device_ids.application_ids.application_id': applicationId,
        'end_device_ids.device_id': deviceId,
      },
    }

    let isResult = {}
    const requests = new Array(3)

    if ('is' in requestTree) {
      isResult = await this._api.EndDeviceRegistry.Get({
        ...params,
        ...Marshaler.pathsToFieldMask(requestTree.is),
      })
    }

    if ('ns' in requestTree) {
      requests[0] = this._api.NsEndDeviceRegistry.Get({
        ...params,
        ...Marshaler.pathsToFieldMask(requestTree.ns),
      })
    }
    if ('as' in requestTree) {
      requests[1] = this._api.AsEndDeviceRegistry.Get({
        ...params,
        ...Marshaler.pathsToFieldMask(requestTree.as),
      })
    }
    if ('js' in requestTree) {
      requests[2] = this._api.JsEndDeviceRegistry.Get({
        ...params,
        ...Marshaler.pathsToFieldMask(requestTree.js),
      })
    }

    const getResults = (await Promise.all(requests))
      .map(e => e ? Marshaler.payloadSingleResponse(e) : undefined)

    const result = this._mergeEntity([
      { record: getResults[0], paths: requestTree.ns },
      { record: getResults[1], paths: requestTree.as },
      { record: getResults[2], paths: requestTree.js },
      { record: Marshaler.payloadSingleResponse(isResult), paths: requestTree.is },
    ])

    return result
  }

  async _deleteDevice (applicationId, deviceId, components = [ 'is', 'ns', 'as', 'js' ]) {
    const requests = Array(4)

    const params = {
      routeParams: {
        'application_ids.application_id': applicationId,
        device_id: deviceId,
      },
    }

    if (components.includes('is')) {
      requests[0] = this._api.EndDeviceRegistry.Delete(params)
    }
    if (components.includes('ns')) {
      requests[1] = this._api.NsEndDeviceRegistry.Delete(params)
    }
    if (components.includes('as')) {
      requests[2] = this._api.AsEndDeviceRegistry.Delete(params)
    }
    if (components.includes('js')) {
      requests[3] = this._api.JsEndDeviceRegistry.Delete(params)
    }

    const deleteResults = (await Promise.all(requests)).map(e => e ? e.status : false)
    return deleteResults
  }

  async getAll (applicationId, params) {
    const response = await this._api.EndDeviceRegistry.List({ queryParams: params })

    return this._responseTransform(response)
  }

  async getById (applicationId, deviceId, selector) {
    const response = await this._getDevice(applicationId, deviceId, Marshaler.selectorToPaths(selector))

    return this._responseTransform(response)
  }

  async updateById (applicationId, deviceId, patch) {
    const response = await this._setDevice(applicationId, patch, true)

    return this._responseTransform(response)
  }

  async create (applicationId, device, { abp = false, setDefaults = true, withRootKeys = false } = {}) {
    let dev = device

    if (setDefaults) {
      dev = {
        application_server_address: this._stackConfig.as,
        join_server_address: this._stackConfig.js,
        network_server_address: this._stackConfig.ns,
        ...device,
      }
    }

    if (abp) {
      const session = {
        dev_addr: randomByteString(8), // TODO: Replace with proper generator
        keys: {
          session_key_id: randomByteString(16),
          f_nwk_s_int_key: {
            key: randomByteString(16, 'base64'),
            kek_label: '',
          },
          app_s_key: {
            key: randomByteString(16),
            kek_label: '',
          },
        },
      }
      if (parseInt(device.lorawan_version.replace(/\D/g, '').padEnd(3, 0)) >= 110) {
        session.keys.s_nwk_s_int_key = {
          key: randomByteString(16, 'base64'),
          kek_label: '',
        }
        session.keys.nwk_s_enc_key = {
          key: randomByteString(16, 'base64'),
          kek_label: '',
        }
      }

      dev.session = {
        ...session,
        ...dev.session,
      }

      dev.supports_join = false

    } else {
      if ('provisioner_id' in dev && dev.provisioner_id !== '') {
        throw new Error('Setting a provisioner with end device keys is not allowed.')
      }
      let root_keys = {}
      if (withRootKeys) {
        root_keys = {
          root_key_id: 'ttn-lw-js-sdk-generated',
          app_key: {
            key: randomByteString(16),
            kek_label: 'default',
          },
          nwk_key: {
            key: randomByteString(16),
            kek_label: 'default',
          },
        }
      }

      dev.root_keys = {
        ...root_keys,
        ...dev.root_keys,
      }

      dev.supports_join = true

    }

    const response = await this._setDevice(applicationId, dev, true)

    return this._responseTransform(response)
  }

  async deleteById (applicationId, deviceId) {
    const result = this._deleteDevice(applicationId, deviceId)

    return result
  }
}

export default Devices
