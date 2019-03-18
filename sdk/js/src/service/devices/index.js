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
import { splitSetPath, splitGetPath } from './split'
import { mergeDevice } from './merge'

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

    if (deviceId && ids && 'device_id' in deviceId && deviceId !== ids.device_id) {
      throw new Error('Device ID mismatch.')
    }

    if (!create && !devId) {
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
      routeParams: {
        'end_device.ids.application_ids.application_id': appId,
      },
    }

    if (!create) {
      params.routeParams['end_device.ids.device_id'] = devId
    }

    // Extract the paths from the patch
    const paths = traverse(device).reduce(function (acc, node) {
      if (this.isLeaf) {
        acc.push(this.path)
      }
      return acc
    }, [])

    const requestTree = splitSetPath(paths, mergeBase)
    const devicePayload = Marshaler.payload(device, 'end_device')

    let isResult = {}
    if (create) {
      isResult = await this._api.EndDeviceRegistry.Create(params, devicePayload)
    }

    params.routeParams['end_device.ids.device_id'] = 'data' in isResult ? isResult.data.ids.device_id : deviceId

    const requests = new Array(3)

    if (!create && 'is' in requestTree) {
      isResult = await this._api.EndDeviceRegistry.Update(params, {
        ...devicePayload,
        ...Marshaler.pathsToFieldMask(requestTree.is),
      })
    }

    if ('ns' in requestTree) {
      requests[0] = this._api.NsEndDeviceRegistry.Set(params, {
        ...devicePayload,
        ...Marshaler.pathsToFieldMask(requestTree.ns),
      })
    }
    if ('as' in requestTree) {
      requests[1] = this._api.AsEndDeviceRegistry.Set(params, {
        ...devicePayload,
        ...Marshaler.pathsToFieldMask(requestTree.as),
      })
    }
    if ('js' in requestTree) {
      requests[2] = this._api.JsEndDeviceRegistry.Set(params, {
        ...devicePayload,
        ...Marshaler.pathsToFieldMask(requestTree.js),
      })
    }

    try {
      const setResults = (await Promise.all(requests))
        .map(e => e ? Marshaler.payloadSingleResponse(e) : undefined)

      const result = mergeDevice([
        { record: setResults[0], paths: requestTree.ns },
        { record: setResults[1], paths: requestTree.as },
        { record: setResults[2], paths: requestTree.js },
        { record: Marshaler.payloadSingleResponse(isResult), paths: requestTree.is },
      ])

      return result
    } catch (err) {
      // Roll back changes
      if (create) {
        this._deleteDevice(appId, devId, Object.keys(requestTree))
      }
      throw new Error('Could not create device.')
    }
  }

  async _getDevice (applicationId, deviceId, paths, ignoreNotFound) {

    if (!applicationId) {
      throw new Error('Missing application_id for device.')
    }

    const requestTree = splitGetPaths(paths)

    const params = {
      routeParams: {
        'end_device_ids.application_ids.application_id': applicationId,
        'end_device_ids.device_id': deviceId,
      },
    }

    let isResult = {}
    const requests = new Array(3)

    // Wrap the request to allow ignoring not found errors
    const requestWrapper = async function (call, params, paths) {
      try {
        const res = await call(params, Marshaler.pathsToFieldMask(paths))
        return res
      } catch (err) {
        if (ignoreNotFound && err.code === 5) {
          return { end_device: {}}
        }
        throw err
      }
    }

    if ('is' in requestTree) {
      isResult = await this._api.EndDeviceRegistry.Get(
        params,
        Marshaler.pathsToFieldMask(requestTree.is),
      )
    }

    if ('ns' in requestTree) {
      requests[0] = await requestWrapper(
        this._api.NsEndDeviceRegistry.Get,
        params,
        requestTree.ns,
      )
    }
    if ('as' in requestTree) {
      requests[1] = await requestWrapper(
        this._api.AsEndDeviceRegistry.Get,
        params,
        requestTree.as,
      )
    }
    if ('js' in requestTree) {
      requests[2] = await requestWrapper(
        this._api.NsEndDeviceRegistry.Get,
        params,
        requestTree.js,
      )
    }

    const getResults = (await Promise.all(requests))
      .map(e => e ? Marshaler.payloadSingleResponse(e) : undefined)

    const result = mergeDevice([
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

    if (setDefaults) {
      dev = {
        application_server_address: new URL(this._stackConfig.as).host,
        join_server_address: new URL(this._stackConfig.js).host,
        network_server_address: new URL(this._stackConfig.ns).host,
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
    const response = await this._setDevice(applicationId, undefined, dev, true)

    return this._responseTransform(response)
  }

  async deleteById (applicationId, deviceId) {
    const result = this._deleteDevice(applicationId, deviceId)

    return result
  }
}

export default Devices
