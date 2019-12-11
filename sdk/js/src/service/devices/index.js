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

/* eslint-disable no-invalid-this, no-await-in-loop */

import traverse from 'traverse'
import Marshaler from '../../util/marshaler'
import Device from '../../entity/device'
import { notify, EVENTS } from '../../api/stream/shared'
import deviceEntityMap from '../../../generated/device-entity-map.json'
import { splitSetPaths, splitGetPaths, makeRequests } from './split'
import mergeDevice from './merge'

/**
 * Devices Class provides an abstraction on all devices and manages data
 * handling from different sources. It exposes an API to easily work with
 * device data.
 */
class Devices {
  constructor(api, { proxy = true, ignoreDisabledComponents = true, stackConfig }) {
    if (!api) {
      throw new Error('Cannot initialize device service without api object.')
    }
    this._api = api
    this._stackConfig = stackConfig
    this._proxy = proxy
    this._ignoreDisabledComponents = ignoreDisabledComponents
  }

  _responseTransform(response, single = true) {
    return Marshaler[single ? 'unwrapDevice' : 'unwrapDevices'](
      response,
      this._proxy ? device => new Device(device, this._api) : undefined,
    )
  }

  async _setDevice(applicationId, deviceId, device, create = false, requestTreeOverwrite) {
    const ids = device.ids
    const devId = deviceId || ('device_id' in ids && ids.device_id)
    const appId = applicationId || ('application_ids' in ids && ids.application_ids.application_id)

    if (deviceId && ids && 'device_id' in ids && deviceId !== ids.device_id) {
      throw new Error('Device ID mismatch.')
    }

    if (!create && !devId) {
      throw new Error('Missing device_id for update operation.')
    }

    if (!appId) {
      throw new Error('Missing application_id for device.')
    }

    // Ensure proper id object
    if (!('ids' in device)) {
      device.ids = { device_id: deviceId, application_ids: { application_id: applicationId } }
    } else if (!device.ids.device_id) {
      device.ids.device_id = deviceId
    } else if (!device.ids.application_ids || !device.ids.application_ids.application_id) {
      device.ids.application_ids = { application_id: applicationId }
    }

    const params = {
      routeParams: {
        'end_device.ids.application_ids.application_id': appId,
      },
    }

    // Extract the paths from the patch
    const deviceMap = traverse(deviceEntityMap)

    const commonPathFilter = function(element, index, array) {
      return deviceMap.has(array.slice(0, index + 1))
    }
    const paths = traverse(device).reduce(function(acc, node) {
      if (this.isLeaf) {
        const path = this.path

        // Only consider adding, if a common parent has not been already added
        if (
          acc.every(
            e =>
              !path
                .slice(-1)
                .join()
                .startsWith(e.join()),
          )
        ) {
          // Add only the deepest possible field mask of the patch
          const commonPath = path.filter(commonPathFilter)
          acc.push(commonPath)
        }
      }
      return acc
    }, [])

    // Make sure to write at least the ids, in case of creation
    const mergeBase = create
      ? {
          ns: [['ids']],
          as: [['ids']],
          js: [['ids']],
        }
      : {}

    const requestTree = requestTreeOverwrite
      ? requestTreeOverwrite
      : splitSetPaths(paths, mergeBase)

    // Retrieve join information if not present
    if (!create && !('supports_join' in device)) {
      const res = await this._getDevice(
        appId,
        devId,
        [['supports_join'], ['join_server_address']],
        true,
      )
      if ('supports_join' in res && res.supports_join) {
        // The NS registry entry exists
        device.supports_join = true
      } else if (res.join_server_address) {
        // The NS registry entry does not exist, but a join_server_address
        // setting suggests that join is supported, so we add the path
        // to the request tree to ensure that it will be set on creation
        device.supports_join = true
        requestTree.ns.push(['supports_join'])
      }
    }

    // Do not query JS when the device is ABP
    if (!device.supports_join) {
      delete requestTree.js
    }

    if (!create) {
      const { network_server_address, application_server_address } = await this._getDevice(
        appId,
        devId,
        [['application_server_address'], ['network_server_address']],
        false,
      )

      try {
        const nsHost = new URL(this._stackConfig.ns).hostname

        if (network_server_address !== nsHost) {
          delete requestTree.as
        }
      } catch (e) {}

      try {
        const asHost = new URL(this._stackConfig.as).hostname

        if (application_server_address !== asHost) {
          delete requestTree.as
        }
      } catch (e) {}
    }

    // Retrieve necessary EUIs in case of a join server query being necessary
    if ('js' in requestTree) {
      if (!create && (!ids || !ids.join_eui || !ids.dev_eui)) {
        const res = await this._getDevice(
          appId,
          devId,
          [['ids', 'join_eui'], ['ids', 'dev_eui']],
          true,
        )
        if (!res.ids || !res.ids.join_eui || !res.ids.dev_eui) {
          throw new Error(
            'Could not update Join Server data on a device without Join EUI or Dev EUI',
          )
        }
        device.ids = {
          ...device.ids,
          join_eui: res.ids.join_eui,
          dev_eui: res.ids.dev_eui,
        }
      }
    }

    // Perform the requests
    const devicePayload = Marshaler.payload(device, 'end_device')
    const setParts = await makeRequests(
      this._api,
      this._stackConfig,
      this._ignoreDisabledComponents,
      create ? 'create' : 'set',
      requestTree,
      params,
      devicePayload,
    )

    // Filter out errored requests
    const errors = setParts.filter(part => part.hasErrored)

    // Handle possible errored requests
    if (errors.length !== 0) {
      // Roll back successfully created registry entries
      if (create) {
        this._deleteDevice(appId, devId, setParts.map(e => e.hasAttempted && !e.hasErrored))
      }

      // Throw the first error
      throw errors[0].error
    }

    const result = mergeDevice(setParts)
    return result
  }

  async _getDevice(applicationId, deviceId, paths, ignoreNotFound) {
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

    const deviceParts = await makeRequests(
      this._api,
      this._stackConfig,
      this._ignoreDisabledComponents,
      'get',
      requestTree,
      params,
      undefined,
      ignoreNotFound,
    )
    const result = mergeDevice(deviceParts)

    return result
  }

  async _deleteDevice(applicationId, deviceId, components = ['is', 'ns', 'as', 'js']) {
    const params = {
      routeParams: {
        'application_ids.application_id': applicationId,
        device_id: deviceId,
      },
    }

    // Compose a request tree
    const requestTree = components.reduce(function(acc, val) {
      acc[val] = undefined
      return acc
    }, {})

    const deleteParts = await makeRequests(
      this._api,
      this._stackConfig,
      this._ignoreDisabledComponents,
      'delete',
      requestTree,
      params,
      undefined,
      true,
    )
    return deleteParts.every(e => Boolean(e.device) && Object.keys(e.device).length === 0)
      ? {}
      : deleteParts
  }

  async getAll(applicationId, params, selector) {
    const response = await this._api.EndDeviceRegistry.List(
      {
        routeParams: { 'application_ids.application_id': applicationId },
      },
      {
        ...params,
        ...Marshaler.selectorToFieldMask(selector),
      },
    )

    return this._responseTransform(response, false)
  }

  async getById(applicationId, deviceId, selector = [['ids']], { ignoreNotFound = false } = {}) {
    const response = await this._getDevice(
      applicationId,
      deviceId,
      Marshaler.selectorToPaths(selector),
      ignoreNotFound,
    )

    return this._responseTransform(response)
  }

  async updateById(applicationId, deviceId, patch) {
    const response = await this._setDevice(applicationId, deviceId, patch)

    if ('root_keys' in patch) {
      patch.supports_join = true
    }

    return this._responseTransform(response)
  }

  async create(applicationId, device, { abp = false } = {}) {
    const dev = device

    if (abp) {
      dev.supports_join = false
    } else {
      if ('provisioner_id' in dev && dev.provisioner_id !== '') {
        throw new Error('Setting a provisioner with end device keys is not allowed.')
      }

      dev.supports_join = true
    }
    const response = await this._setDevice(applicationId, undefined, dev, true)

    return this._responseTransform(response)
  }

  async deleteById(applicationId, deviceId) {
    const result = this._deleteDevice(applicationId, deviceId)

    return result
  }

  // End Device Template Converter

  async listTemplateFormats() {
    const result = await this._api.EndDeviceTemplateConverter.ListFormats()
    const payload = Marshaler.payloadSingleResponse(result)

    return payload.formats
  }

  convertTemplate(formatId, data) {
    // This is a stream endpoint
    return this._api.EndDeviceTemplateConverter.Convert(undefined, {
      format_id: formatId,
      data,
    })
  }

  bulkCreate(applicationId, deviceOrDevices, components = ['is', 'ns', 'as', 'js']) {
    const devices = !(deviceOrDevices instanceof Array) ? [deviceOrDevices] : deviceOrDevices
    let listeners = Object.values(EVENTS).reduce((acc, curr) => ({ ...acc, [curr]: null }), {})
    let finishedCount = 0
    let stopRequested = false

    const runTasks = async function() {
      for (const device of devices) {
        if (stopRequested) {
          notify(listeners[EVENTS.CLOSE])
          listeners = null
          break
        }

        try {
          const {
            field_mask: { paths },
            end_device,
          } = device

          const requestTree = splitSetPaths(Marshaler.selectorToPaths(paths), undefined, components)

          const result = await this._setDevice(
            applicationId,
            undefined,
            end_device,
            true,
            requestTree,
          )
          notify(listeners[EVENTS.CHUNK], result)
          finishedCount++
          if (finishedCount === devices.length) {
            notify(listeners[EVENTS.CLOSE])
            listeners = null
          }
        } catch (error) {
          notify(listeners[EVENTS.ERROR], error)
          listeners = null
          break
        }
      }
    }

    runTasks.bind(this)()

    return {
      on(eventName, callback) {
        if (listeners[eventName] === undefined) {
          throw new Error(
            `${eventName} event is not supported. Should be one of: start, error, chunk or close`,
          )
        }

        listeners[eventName] = callback

        return this
      },
      abort() {
        stopRequested = true
      },
    }
  }

  // Events Stream

  async openStream(identifiers, tail, after) {
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
