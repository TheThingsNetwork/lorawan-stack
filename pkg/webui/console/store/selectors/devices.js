// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import getHostnameFromUrl from '@ttn-lw/lib/host-from-url'
import { combineDeviceIds, extractDeviceIdFromCombinedId } from '@ttn-lw/lib/selectors/id'
import { createFetchingSelector } from '@ttn-lw/lib/store/selectors/fetching'
import { createErrorSelector } from '@ttn-lw/lib/store/selectors/error'
import {
  createPaginationIdsSelectorByEntity,
  createPaginationTotalCountSelectorByEntity,
} from '@ttn-lw/lib/store/selectors/pagination'
import { selectAsConfig, selectJsConfig, selectNsConfig } from '@ttn-lw/lib/selectors/env'

import { GET_DEV_BASE, GET_DEVICES_LIST_BASE, GET_DEV_VERSION_IDS_BASE } from '@console/store/actions/devices'

import {
  createEventsSelector,
  createEventsErrorSelector,
  createEventsStatusSelector,
  createEventsInterruptedSelector,
  createEventsPausedSelector,
  createEventsTruncatedSelector,
  createEventsFilterSelector,
} from './events'

const ENTITY = 'devices'

// Device.
export const selectDeviceStore = state => state.devices
export const selectDeviceEntitiesStore = state => selectDeviceStore(state).entities
export const selectDeviceDerivedStore = state => selectDeviceStore(state).derived
export const selectDeviceByIds = (state, appId, devId) =>
  selectDeviceById(state, combineDeviceIds(appId, devId))
export const selectDeviceById = (state, id) => selectDeviceEntitiesStore(state)[id]
export const selectDeviceDerivedById = (state, id) => selectDeviceDerivedStore(state)[id]
export const selectSelectedDeviceId = state =>
  extractDeviceIdFromCombinedId(selectDeviceStore(state).selectedDevice)
export const selectSelectedCombinedDeviceId = state => selectDeviceStore(state).selectedDevice
export const selectSelectedDevice = state =>
  selectDeviceById(state, selectSelectedCombinedDeviceId(state))
export const selectSelectedDeviceFormatters = state => selectSelectedDevice(state).formatters
export const selectDeviceFetching = createFetchingSelector(GET_DEV_BASE)
export const selectDeviceError = createErrorSelector(GET_DEV_BASE)
export const isOtherClusterDevice = device => {
  const isOtherCluster =
    getHostnameFromUrl(selectAsConfig().base_url) !== device.application_server_address &&
    getHostnameFromUrl(selectNsConfig().base_url) !== device.network_server_address &&
    getHostnameFromUrl(selectJsConfig().base_url) !== device.join_server_address

  return isOtherCluster
}
export const selectVersionIds = state => selectDeviceStore(state).version_ids

// Derived.
export const selectDeviceDerivedUplinkFrameCount = (state, appId, devId) => {
  const derived = selectDeviceDerivedById(state, combineDeviceIds(appId, devId))
  if (!Boolean(derived)) return undefined

  return derived.uplinkFrameCount
}
export const selectDeviceDerivedDownlinkFrameCount = (state, appId, devId) => {
  const derived = selectDeviceDerivedById(state, combineDeviceIds(appId, devId))
  if (!Boolean(derived)) return undefined

  return derived.downlinkFrameCount
}
export const selectDeviceDerivedLastSeen = (state, appId, devId) => {
  const derived = selectDeviceDerivedById(state, combineDeviceIds(appId, devId))
  if (!Boolean(derived)) return undefined

  return derived.lastSeen
}

// Devices.
const selectDevsIds = createPaginationIdsSelectorByEntity(ENTITY)
const selectDevsTotalCount = createPaginationTotalCountSelectorByEntity(ENTITY)
const selectDevsFetching = createFetchingSelector(GET_DEVICES_LIST_BASE)
const selectDevsError = createErrorSelector(GET_DEVICES_LIST_BASE)

export const selectDevices = state => selectDevsIds(state).map(id => selectDeviceById(state, id))
export const selectDevicesTotalCount = state => selectDevsTotalCount(state)
export const selectDevicesFetching = state => selectDevsFetching(state)
export const selectDevicesError = state => selectDevsError(state)

// Events.
export const selectDeviceEvents = createEventsSelector(ENTITY)
export const selectDeviceEventsError = createEventsErrorSelector(ENTITY)
export const selectDeviceEventsStatus = createEventsStatusSelector(ENTITY)
export const selectDeviceEventsInterruptted = createEventsInterruptedSelector(ENTITY)
export const selectDeviceEventsPaused = createEventsPausedSelector(ENTITY)
export const selectDeviceEventsTruncated = createEventsTruncatedSelector(ENTITY)
export const selectDeviceEventsFilter = createEventsFilterSelector(ENTITY)
