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

import { URL } from 'url'
import traverse from 'traverse'
import Marshaler from '../../util/marshaler'
import Device from '../../entity/device'
import randomByteString from '../../util/random-bytes'
import deviceEntityMap from '../../../generated/device-entity-map.json'
import { splitSetPaths, splitGetPaths, makeRequests } from './split'
import mergeDevice from './merge'


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
    this._proxy = proxy
  }

  _responseTransform (response, single = true) {
    return Marshaler[single ? 'unwrapDevice' : 'unwrapDevices'](
      response,
      this._proxy
        ? device => new Device(device, this._api)
        : undefined
    )
  }

  async _setDevice (applicationId, deviceId, device, create = false) {
    const ids = device.ids
    const devId = deviceId || 'device_id' in ids && ids.device_id
    const appId = applicationId || 'application_ids' in ids && ids.application_ids.application_id

    if (deviceId && ids && 'device_id' in ids && deviceId !== ids.device_id) {
      throw new Error('Device ID mismatch.')
    }

    if (!create && !devId) {
      throw new Error('Missing device_id for update operation.')
    }

    if (!appId) {
      throw new Error('Missing application_id for device.')
    }

    // Make sure to write at least the ids, in case of creation
    const mergeBase = create ? {
      ns: [[ 'ids' ]],
      as: [[ 'ids' ]],
      js: [[ 'ids' ]],
    } : {}

    const params = {
      routeParams: {
        'end_device.ids.application_ids.application_id': appId,
      },
    }

    // Extract the paths from the patch
    const deviceMap = traverse(deviceEntityMap)

    const commonPathFilter = function (element, index, array) {
      return deviceMap.has(array.slice(0, index + 1))
    }
    const paths = traverse(device).reduce(function (acc, node) {
      if (this.isLeaf) {
        const path = this.path

        // Only consider adding, if a common parent has not been already added
        if (acc.every(e => !path.slice(-1).join().startsWith(e.join()))) {
          // Add only the deepest possible field mask of the patch
          const commonPath = path.filter(commonPathFilter)
          acc.push(commonPath)
        }
      }
      return acc
    }, [])

    const requestTree = splitSetPaths(paths, mergeBase)
    const devicePayload = Marshaler.payload(device, 'end_device')

    let isResult = {}
    if (create) {
      isResult = await this._api.EndDeviceRegistry.Create(params, devicePayload)
      isResult = Marshaler.payloadSingleResponse(isResult)
      delete requestTree.is
    }

    // Retrieve join information if not present
    if (!create && !('supports_join' in device)) {
      try {
        const res = await this._getDevice(appId, devId, [[ 'supports_join' ]])
        device.supports_join = res.supports_join
      } catch (err) {
        throw new Error('Could not retrieve join information of the device')
      }
    }

    // Do not query JS when the device is ABP
    if (!device.supports_join) {
      delete requestTree.js
    }

    // Retrieve necessary EUIs in case of a join server query being necessary
    if ('js' in requestTree) {
      if (!create && (!ids || !ids.join_eui || !ids.dev_eui)) {
        try {
          const res = await this._getDevice(appId, devId, [[ 'ids', 'join_eui' ], [ 'ids', 'dev_eui' ]])
          device.ids = {
            ...device.ids,
            join_eui: res.ids.join_eui,
            dev_eui: res.ids.dev_eui,
          }
        } catch (err) {
          throw new Error('Could not update Join Server data on a device without Join EUI or Dev EUI')
        }
      }
    }

    // Write the device id param based on either the id of the newly created
    // device, or the passed id argument
    params.routeParams['end_device.ids.device_id'] = 'data' in isResult ? isResult.ids.device_id : devId

    try {
      const setParts = await makeRequests(this._api, 'set', requestTree, params, devicePayload)
      const result = mergeDevice(setParts, isResult)
      return result
    } catch (err) {
      // Roll back changes
      if (create) {
        this._deleteDevice(appId, devId, Object.keys(requestTree))
      }
      throw new Error(`Could not ${create ? 'create' : 'update'} device.`)
    }
  }

  async _getDevice (applicationId, deviceId, paths, ignoreNotFound) {

    if (!applicationId) {
      throw new Error('Missing application_id for device.')
    }

    if (!deviceId) {
      throw new Error('Missing device_id for device.')
    }

    const requestTree = splitGetPaths(paths)

    const params = {
      routeParams: {
        'end_device_ids.application_ids.application_id': applicationId,
        'end_device_ids.device_id': deviceId,
      },
    }

    const deviceParts = await makeRequests(this._api, 'get', requestTree, params, undefined, ignoreNotFound)
    const result = mergeDevice(deviceParts)

    return result
  }

  async _deleteDevice (applicationId, deviceId, components = [ 'is', 'ns', 'as', 'js' ]) {
    const params = {
      routeParams: {
        'application_ids.application_id': applicationId,
        device_id: deviceId,
      },
    }

    // Compose a request tree
    const requestTree = components.reduce(function (acc, val) {
      acc[val] = undefined
      return acc
    }, {})

    const deleteParts = await makeRequests(this._api, 'delete', requestTree, params)
    return deleteParts.every(e => Object.keys(e.device).length === 0) ? {} : deleteParts
  }

  async getAll (applicationId, params, selector) {
    const response = await this._api.EndDeviceRegistry.List({
      routeParams: { 'application_ids.application_id': applicationId },
    }, {
      ...params,
      ...Marshaler.selectorToFieldMask(selector),
    })

    return this._responseTransform(response, false)
  }

  async getById (applicationId, deviceId, selector = [[ 'ids' ]], { ignoreNotFound = false } = {}) {
    const response = await this._getDevice(applicationId, deviceId, Marshaler.selectorToPaths(selector), ignoreNotFound)

    return this._responseTransform(response)
  }

  async updateById (applicationId, deviceId, patch) {
    const response = await this._setDevice(applicationId, deviceId, patch)

    if ('root_keys' in patch) {
      patch.supports_join = true
    }

    return this._responseTransform(response)
  }

  async create (applicationId, device, { abp = false, setDefaults = true, withRootKeys = false } = {}) {
    let dev = device
    const Url = URL ? URL : window.URL

    if (setDefaults) {
      dev = {
        application_server_address: new Url(this._stackConfig.as).host,
        join_server_address: new Url(this._stackConfig.js).host,
        network_server_address: new Url(this._stackConfig.ns).host,
        ...device,
      }
    }

    if (abp) {
      const session = {
        dev_addr: randomByteString(8), // TODO: Replace with proper generator
        keys: {
          session_key_id: randomByteString(16),
          f_nwk_s_int_key: {
            key: randomByteString(32),
          },
          app_s_key: {
            key: randomByteString(32),
          },
        },
      }
      if (parseInt(device.lorawan_version.replace(/\D/g, '').padEnd(3, 0)) >= 110) {
        session.keys.s_nwk_s_int_key = {
          key: randomByteString(32),
        }
        session.keys.nwk_s_enc_key = {
          key: randomByteString(32),
        }
      }

      let providedKeys = {}
      if (dev.session && dev.session.keys) {
        providedKeys = dev.session.keys
      }

      dev.session = {
        ...session,
        ...dev.session,
        keys: {
          ...session.keys,
          ...providedKeys,
        },
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
            key: randomByteString(32),
          },
          nwk_key: {
            key: randomByteString(32),
          },
        }
      }

      dev.root_keys = {
        ...root_keys,
        ...dev.root_keys,
      }

      dev.supports_join = true

    }
    const response = await this._setDevice(applicationId, undefined, dev, true)

    return this._responseTransform(response)
  }

  async deleteById (applicationId, deviceId) {
    const result = this._deleteDevice(applicationId, deviceId)

    return result
  }

  // Events Stream

  async openStream (identifiers, tail, after) {
    const payload = {
      identifiers: identifiers.map(ids => ({
        device_ids: ids,
      })),
      tail,
      after,
    }

    return this._api.Events.Stream(undefined, payload)
  }
}

export default Devices
