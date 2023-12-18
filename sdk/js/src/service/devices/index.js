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

import autoBind from 'auto-bind'
import traverse from 'traverse'

import { notify, EVENTS } from '../../api/stream/shared'
import Marshaler from '../../util/marshaler'
import subscribeToWebSocketStreams from '../../api/stream/subscribeToWebSocketStreams'
import deviceEntityMap from '../../../generated/device-entity-map.json'
import DownlinkQueue from '../downlink-queue'
import { STACK_COMPONENTS_MAP } from '../../util/constants'
import DeviceClaim from '../claim'

import Repository from './repository'
import { splitSetPaths, splitGetPaths, makeRequests } from './split'
import mergeDevice from './merge'

const { is: IS, ns: NS, as: AS, js: JS, gs: GS } = STACK_COMPONENTS_MAP
/**
 * Devices Class provides an abstraction on all devices and manages data
 * handling from different sources. It exposes an API to easily work with
 * device data.
 */
class Devices {
  constructor(api, { stackConfig }) {
    if (!api) {
      throw new Error('Cannot initialize device service without api object.')
    }
    this._api = api
    this._stackConfig = stackConfig

    this.DownlinkQueue = new DownlinkQueue(api.AppAs, { stackConfig })
    this.Repository = new Repository(api.DeviceRepository)
    this.DeviceClaim = new DeviceClaim(api.DeviceClaim, { stackConfig })

    this.deviceCreationAllowedFieldMaskPaths = [
      ...this._api.EndDeviceRegistry.UpdateAllowedFieldMaskPaths,
      ...this._api.NsEndDeviceRegistry.SetAllowedFieldMaskPaths,
      ...this._api.AsEndDeviceRegistry.SetAllowedFieldMaskPaths,
      ...this._api.JsEndDeviceRegistry.SetAllowedFieldMaskPaths,
      // Store unique entries only.
    ].filter((path, index, paths) => paths.indexOf(path) === index)

    autoBind(this)
  }

  _emitDefaults(paths, device) {
    // Handle zero coordinates that are swallowed by the grpc-gateway for device
    // location.
    const hasLocation = Boolean(device.locations) && Boolean(device.locations.user)
    const requestedLocation = paths.some(path => path.startsWith('location'))

    if (hasLocation && requestedLocation) {
      const { locations } = device

      if (!('altitude' in locations.user)) {
        locations.user.altitude = 0
      }

      if (!('longitude' in locations.user)) {
        locations.user.longitude = 0
      }

      if (!('latitude' in locations.user)) {
        locations.user.latitude = 0
      }
    }

    if (paths.includes('claim_authentication_code') && !Boolean(device.claim_authentication_code)) {
      device.claim_authentication_code = null
    }

    if (paths.includes('formatters') && !Boolean(device.formatters)) {
      device.formatters = null
    }

    if (paths.includes('session') && !Boolean(device.session)) {
      device.session = null
    }

    if (paths.includes('pending_session') && !Boolean(device.pending_session)) {
      device.pending_session = null
    }

    if (paths.includes('formatters.up_formatter')) {
      if (!Boolean(device.formatters)) {
        device.formatters = { up_formatter: 'FORMATTER_NONE' }
      }
      if (!Boolean(device.formatters.up_formatter)) {
        device.formatters.up_formatter = 'FORMATTER_NONE'
      }
    }

    if (paths.includes('formatters.down_formatter')) {
      if (!Boolean(device.formatters)) {
        device.formatters = { down_formatter: 'FORMATTER_NONE' }
      }
      if (!Boolean(device.formatters.down_formatter)) {
        device.formatters.down_formatter = 'FORMATTER_NONE'
      }
    }

    if (paths.includes('mac_settings')) {
      const { mac_settings = {} } = device

      if (
        Boolean(mac_settings.ping_slot_periodicity) &&
        typeof mac_settings.ping_slot_periodicity === 'undefined'
      ) {
        mac_settings.ping_slot_periodicity = 'PING_EVERY_1S'
      }

      if (
        Boolean(mac_settings.rx2_data_rate_index) &&
        typeof mac_settings.rx2_data_rate_index === 'undefined'
      ) {
        mac_settings.rx2_data_rate_index = 0
      }

      if (mac_settings.adr && 'static' in mac_settings.adr) {
        if (typeof mac_settings.adr.static.data_rate_index === 'undefined') {
          mac_settings.adr.static.data_rate_index = 0
        }

        if (typeof mac_settings.adr.static.nb_trans === 'undefined') {
          mac_settings.adr.static.nb_trans = 0
        }

        if (typeof mac_settings.adr.static.tx_power_index === 'undefined') {
          mac_settings.adr.static.tx_power_index = 0
        }
      }
    }

    return device
  }

  async _getDevice(applicationId, deviceId, paths, ignoreNotFound, mergeResult = true, components) {
    if (!applicationId) {
      throw new Error('Missing application_id for device.')
    }

    if (!deviceId) {
      throw new Error('Missing device_id for device.')
    }

    const requestTree = splitGetPaths(paths, undefined, components)

    const params = {
      routeParams: {
        'end_device_ids.application_ids.application_id': applicationId,
        'end_device_ids.device_id': deviceId,
      },
    }

    const deviceParts = await makeRequests(
      this._api,
      this._stackConfig,
      'get',
      requestTree,
      params,
      undefined,
      ignoreNotFound,
    )

    return mergeResult ? mergeDevice(deviceParts) : deviceParts
  }

  async _deleteDevice(applicationId, deviceId, components = [IS, NS, AS, JS]) {
    if (!Boolean(applicationId)) {
      throw new Error('Missing application ID for device')
    }

    if (!Boolean(deviceId)) {
      throw new Error('Missing end device ID')
    }

    const params = {
      routeParams: {
        'application_ids.application_id': applicationId,
        device_id: deviceId,
      },
    }

    const requests = []
    if (this._stackConfig.isComponentAvailable(AS) && components.includes(AS)) {
      requests.push(this._api.AsEndDeviceRegistry.Delete(params))
    }
    if (this._stackConfig.isComponentAvailable(JS) && components.includes(JS)) {
      requests.push(this._api.JsEndDeviceRegistry.Delete(params))
    }
    if (this._stackConfig.isComponentAvailable(NS) && components.includes(NS)) {
      requests.push(this._api.NsEndDeviceRegistry.Delete(params))
    }

    const responses = await Promise.all(
      // Simulate behavior of allSettled.
      requests.map(promise =>
        promise.then(
          value => ({
            status: 'fulfilled',
            value,
          }),
          reason => ({ status: 'rejected', reason }),
        ),
      ),
    )

    // Check for errors and filter out 404 errors. We do not regard 404 responses
    // from ns,as and js as failed requests.
    const errors = responses.filter(
      ({ status, reason }) => status === 'rejected' && reason.code !== 5,
    )

    // Only proceed deleting the device from IS (so it is not accessible
    // anymore) if there are no errors.
    if (errors.length > 0) {
      throw errors[0].reason
    }

    if (this._stackConfig.isComponentAvailable(IS) && components.includes(IS)) {
      const response = await this._api.EndDeviceRegistry.Delete(params)

      return Marshaler.payloadSingleResponse(response)
    }

    return {}
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

    return Marshaler.unwrapDevices(response)
  }

  async search(applicationId, params, selector) {
    const response = await this._api.EndDeviceRegistrySearch.SearchEndDevices(
      {
        routeParams: { 'application_ids.application_id': applicationId },
      },
      {
        ...params,
        ...Marshaler.selectorToFieldMask(selector),
      },
    )

    return Marshaler.payloadListResponse('end_devices', response)
  }

  /**
   * Gets the `deviceId` end device under the `applicationId` application.
   * This method will assemble the end device from all available stack
   * components (i.e. NS, AS, IS, JS) based on the provided `selector`
   * and the end device existence in the respective components.
   * Note, this method throws an error if the requested end device does not
   * exist in the IS.
   *
   * @param {string} applicationId - The Application ID.
   * @param {string} deviceId - The Device ID.
   * @param {Array} selector - The list of end device fields to fetch.
   * @param {Array} components - A whitelist of components to source the
   * data from. Selects all by default.
   * @returns {object} - End device on successful requests, an error otherwise.
   */
  async getById(applicationId, deviceId, selector = [['ids']], components) {
    const deviceParts = await this._getDevice(
      applicationId,
      deviceId,
      Marshaler.selectorToPaths(selector),
      false,
      false,
      components,
    )

    const errors = deviceParts.filter(part => {
      // Consider all errors from IS and ignore 404 for JS, AS and NS
      if (part.hasErrored && (part.component === IS || part.error.code !== 5)) {
        return true
      }

      return false
    })

    if (errors.length > 0) {
      throw errors[0].error
    }

    const mergedDevice = mergeDevice(deviceParts)

    const { field_mask } = Marshaler.selectorToFieldMask(selector)

    return this._emitDefaults(field_mask.paths, Marshaler.unwrapDevice(mergedDevice))
  }

  /**
   * Updates the `deviceId` end device under the `applicationId` application.
   * This method will cause updates of the end device in all available stack
   * components (i.e. NS, AS, IS, JS) based on provided end device payload.
   *
   * @param {string} applicationId - The application ID.
   * @param {string} deviceId -The end device ID.
   * @param {object} patch - The end device payload.
   * @returns {object} - Updated end device on successful update, an error
   * otherwise.
   */
  async updateById(applicationId, deviceId, patch) {
    if (!Boolean(applicationId)) {
      throw new Error('Missing application ID for device')
    }

    if (!Boolean(deviceId)) {
      throw new Error('Missing end device ID')
    }

    const deviceMap = traverse(deviceEntityMap)
    const allowedPaths = this.deviceCreationAllowedFieldMaskPaths
    const paths = traverse(patch).reduce(function (acc) {
      // Disregard illegal paths.
      if (!allowedPaths.some(p => this.path.join('.').startsWith(p))) {
        return acc
      }
      // Only add the top level path for arrays, otherwise paths are generated
      // for each item in the array.
      if (Array.isArray(this.node)) {
        acc.push(this.path)
        this.update(this.node, true)
      }

      if (this.isLeaf) {
        const path = this.path

        const parentAdded = acc.some(e => path[0].startsWith(e.join()))

        // Only consider adding, if a common parent has not been already added.
        if (!parentAdded) {
          // Add only the deepest possible field mask of the patch.
          const commonPath = path.filter((_, index, array) => {
            const arr = array.slice(0, index + 1)
            return deviceMap.has(arr)
          })

          acc.push(commonPath)
        }
      }
      return acc
    }, [])

    const requestTree = splitSetPaths(paths)

    // Assemble paths for end device fields that need to be retrieved first to
    // make the update request.
    const combinePaths = []
    if (AS in requestTree && !('application_server_address' in patch)) {
      combinePaths.push(['application_server_address'])
    }
    if (JS in requestTree && !('join_server_address' in patch)) {
      combinePaths.push(['join_server_address'])
      combinePaths.push(['supports_join'])

      const { ids = {} } = patch
      if (!('dev_eui' in ids) || !('join_eui' in ids)) {
        combinePaths.push(['ids', 'dev_eui'])
        combinePaths.push(['ids', 'join_eui'])
      }
    }
    if (NS in requestTree && !('network_server_address' in patch)) {
      combinePaths.push(['network_server_address'])
    }

    const assembledValues = await this._getDevice(applicationId, deviceId, combinePaths, true)

    if (assembledValues.network_server_address !== this._stackConfig.nsHost) {
      delete requestTree.ns
    }

    if (assembledValues.application_server_address !== this._stackConfig.asHost) {
      delete requestTree.as
    }

    if (
      (this._stackConfig.isComponentAvailable(NS) && !assembledValues.supports_join) ||
      assembledValues.join_server_address !== this._stackConfig.jsHost
    ) {
      delete requestTree.js
    }

    // Make sure to include `join_eui` and `dev_eui` for js request as those are
    // required.
    if (JS in requestTree) {
      const { ids = {} } = patch
      const {
        ids: { join_eui, dev_eui },
      } = assembledValues

      patch.ids = {
        ...ids,
        join_eui,
        dev_eui,
      }
    }

    const routeParams = {
      routeParams: {
        'end_device.ids.application_ids.application_id': applicationId,
        'end_device.ids.device_id': deviceId,
      },
    }

    // Perform the requests.
    const devicePayload = Marshaler.payload(patch, 'end_device')
    const setParts = await makeRequests(
      this._api,
      this._stackConfig,
      'set',
      requestTree,
      routeParams,
      devicePayload,
    )

    // Filter out errored requests.
    const errors = setParts.filter(part => part.hasErrored)

    // Handle possible errored requests.
    if (errors.length !== 0) {
      // Throw the first error.
      throw errors[0].error
    }

    return this._emitDefaults(
      Marshaler.fieldMaskFromPatch(patch),
      Marshaler.unwrapDevice(mergeDevice(setParts)),
    )
  }

  async resetById(applicationId, deviceId) {
    const result = await this._api.NsEndDeviceRegistry.ResetFactoryDefaults({
      routeParams: {
        'end_device_ids.application_ids.application_id': applicationId,
        'end_device_ids.device_id': deviceId,
      },
    })

    return Marshaler.payloadSingleResponse(result)
  }

  /**
   * Creates an end device under the `applicationId` application.
   * This method will cause creating the end device in all available stack
   * components (i.e. NS, AS, IS, JS) based on provided end device payload
   * (`device`) or on field mask paths (`mask`).
   *
   * @param {string} applicationId - Application ID.
   * @param {object} device - The end device payload.
   * @param {Array} mask -The field mask paths (by default is generated from
   * `device` payload).
   * @returns {object} - Created end device on successful creation, an error
   * otherwise.
   */
  async create(
    applicationId,
    device,
    mask = Marshaler.fieldMaskFromPatch(device, this.deviceCreationAllowedFieldMaskPaths),
  ) {
    if (!Boolean(applicationId)) {
      throw new Error('Missing application ID for device')
    }

    const { authenticated_identifiers, target_device_id, ...submitValues } = device
    const { supports_join = false, ids = {} } = submitValues

    // Initiate claiming, if the device is claimable.
    const hasAuthenticatedIdentifiers = Boolean(authenticated_identifiers)
    const claimInfoResponse = await this._api.EndDeviceClaimingServer.GetInfoByJoinEUI(undefined, {
      join_eui: ids.join_eui,
    })
    const claim = Marshaler.payloadSingleResponse(claimInfoResponse)
    const supportsClaiming = claim?.supports_claiming ?? false
    let claimDeviceIds
    if (supportsClaiming) {
      // Since this device is claimable, the creation on the join server needs to be skipped.
      submitValues.join_server_address = undefined
      const claimPayload = hasAuthenticatedIdentifiers
        ? {
            authenticated_identifiers,
            target_device_id,
            target_application_ids: {
              application_id: applicationId,
            },
          }
        : {
            authenticated_identifiers: {
              dev_eui: ids.dev_eui,
              authentication_code: device.claim_authentication_code?.value,
              join_eui: ids.join_eui,
            },
            target_device_id: ids.device_id,
            target_application_ids: {
              application_id: applicationId,
            },
          }
      const claimResponse = await this._api.EndDeviceClaimingServer.Claim(undefined, claimPayload)
      claimDeviceIds = Marshaler.payloadSingleResponse(claimResponse)
    }

    let newFieldmasks = mask
    // Apply the resulting IDs to the end_device.
    if (claimDeviceIds) {
      submitValues.ids = { ...ids, ...claimDeviceIds }
      newFieldmasks = Marshaler.fieldMaskFromPatch(
        submitValues,
        this.deviceCreationAllowedFieldMaskPaths,
      )
    }

    const deviceId = submitValues.ids.device_id
    if (!Boolean(deviceId)) {
      throw new Error('Missing end device ID')
    }

    const requestTree = splitSetPaths(Marshaler.selectorToPaths(newFieldmasks))

    if (!supports_join || device.join_server_address !== this._stackConfig.jsHost) {
      delete requestTree.js
    }

    if (device.network_server_address !== this._stackConfig.nsHost) {
      delete requestTree.ns
    }

    if (device.application_server_address !== this._stackConfig.asHost) {
      delete requestTree.as
    }

    const devicePayload = Marshaler.payload(submitValues, 'end_device')
    const routeParams = {
      routeParams: {
        'end_device.ids.application_ids.application_id': applicationId,
      },
    }

    const setParts = await makeRequests(
      this._api,
      this._stackConfig,
      'create',
      requestTree,
      routeParams,
      devicePayload,
    )

    // Filter out errored requests.
    const errors = setParts.filter(part => part.hasErrored)

    // Handle possible errored requests.
    if (errors.length !== 0) {
      // Roll back successfully created registry entries.
      const rollbackComponents = setParts.reduce((components, part) => {
        if (part.hasAttempted && !part.hasErrored) {
          components.push(part.component)
        }

        return components
      }, [])

      await this._deleteDevice(applicationId, deviceId, rollbackComponents)

      // Throw the first error.
      throw errors[0].error
    }

    return mergeDevice(setParts)
  }

  /**
   * Deletes the `deviceId` end device under the `applicationId` application.
   * This method will cause deletion of the end device in all available stack
   * components (i.e. NS, AS, IS, JS).
   *
   * @param {string} applicationId - The application ID.
   * @param {string} deviceId - The end evice ID.
   * @returns {object} - Empty object on successful update, an error otherwise.
   */
  async deleteById(applicationId, deviceId) {
    return this._deleteDevice(applicationId, deviceId)
  }

  // End Device Template Converter.

  async listTemplateFormats() {
    const result = await this._api.EndDeviceTemplateConverter.ListFormats()
    const payload = Marshaler.payloadSingleResponse(result)

    return payload.formats
  }

  convertTemplate(formatId, data) {
    // This is a stream endpoint.
    return this._api.EndDeviceTemplateConverter.Convert(undefined, {
      format_id: formatId,
      data,
    })
  }

  bulkCreate(applicationId, deviceOrDevices) {
    const devices = !(deviceOrDevices instanceof Array) ? [deviceOrDevices] : deviceOrDevices
    let listeners = Object.values(EVENTS).reduce((acc, curr) => ({ ...acc, [curr]: null }), {})
    let finishedCount = 0
    let stopRequested = false

    const runTasks = async function () {
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

          const result = await this.create(applicationId, end_device, paths)

          notify(listeners[EVENTS.MESSAGE], result)
          finishedCount++
        } catch (error) {
          notify(listeners[EVENTS.ERROR], error)
          finishedCount++
        } finally {
          if (finishedCount === devices.length) {
            notify(listeners[EVENTS.CLOSE])
            listeners = null
          }
        }
      }
    }

    return {
      start: runTasks.bind(this),
      on(eventName, callback) {
        if (listeners[eventName] === undefined) {
          throw new Error(
            `${eventName} event is not supported. Should be one of: open, error, message or close`,
          )
        }

        listeners[eventName] = callback

        return this
      },
      abort: () => {
        stopRequested = true
      },
    }
  }

  // Events Stream

  async openStream(identifiers, names, tail, after, listeners) {
    const payload = {
      identifiers: identifiers.map(ids => ({
        device_ids: ids,
      })),
      names,
      tail,
      after,
    }

    // Event streams can come from multiple stack components. It is necessary to
    // check for stack components on different hosts and open distinct stream
    // connections for any distinct host if need be.
    const distinctComponents = this._stackConfig.getComponentsWithDistinctBaseUrls([IS, GS])

    const baseUrls = new Set(
      distinctComponents.map(component => this._stackConfig.getComponentUrlByName(component)),
    )
    // Combine all stream sources to one subscription generator.
    return subscribeToWebSocketStreams(payload, [...baseUrls], listeners)
  }

  async simulateUplink(applicationId, deviceId, uplink) {
    const result = await this._api.AppAs.SimulateUplink(
      {
        routeParams: {
          'end_device_ids.application_ids.application_id': applicationId,
          'end_device_ids.device_id': deviceId,
        },
      },
      {
        uplink_message: uplink,
      },
    )

    return Marshaler.payloadSingleResponse(result)
  }
}

export default Devices
