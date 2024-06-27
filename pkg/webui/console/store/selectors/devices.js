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

import { createSelector } from 'reselect'

import getHostnameFromUrl from '@ttn-lw/lib/host-from-url'
import { combineDeviceIds, extractDeviceIdFromCombinedId } from '@ttn-lw/lib/selectors/id'
import {
  createPaginationIdsSelectorByEntity,
  createPaginationTotalCountSelectorByEntity,
} from '@ttn-lw/lib/store/selectors/pagination'
import { selectAsConfig, selectJsConfig, selectNsConfig } from '@ttn-lw/lib/selectors/env'
import { createFetchingSelector } from '@ttn-lw/lib/store/selectors/fetching'

import { GET_DEV_BASE } from '../actions/devices'

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
export const selectDeviceFetching = createFetchingSelector(GET_DEV_BASE)
export const selectDeviceById = (state, id) => selectDeviceEntitiesStore(state)[id]
export const selectDeviceDerivedById = (state, id) => selectDeviceDerivedStore(state)[id]
export const selectSelectedDeviceId = state =>
  extractDeviceIdFromCombinedId(selectDeviceStore(state).selectedDevice)
export const selectSelectedCombinedDeviceId = state => selectDeviceStore(state).selectedDevice
export const selectSelectedDevice = state =>
  selectDeviceById(state, selectSelectedCombinedDeviceId(state))
export const selectSelectedDeviceFormatters = state => selectSelectedDevice(state).formatters
export const isOtherClusterDevice = device => {
  const isOtherCluster =
    getHostnameFromUrl(selectAsConfig().base_url) !== device.application_server_address &&
    getHostnameFromUrl(selectNsConfig().base_url) !== device.network_server_address &&
    getHostnameFromUrl(selectJsConfig().base_url) !== device.join_server_address

  return isOtherCluster
}
export const selectDeviceLastSeen = (state, appId, devId) => {
  const device = selectDeviceById(state, combineDeviceIds(appId, devId))
  if (!Boolean(device)) return undefined

  return device.last_seen_at
}

// Derived.
export const selectDeviceDerivedUplinkFrameCount = (state, appId, devId) => {
  const derived = selectDeviceDerivedById(state, combineDeviceIds(appId, devId))
  if (!Boolean(derived)) return undefined

  return derived.uplinkFrameCount
}
export const selectDeviceDerivedAppDownlinkFrameCount = (state, appId, devId) => {
  const derived = selectDeviceDerivedById(state, combineDeviceIds(appId, devId))
  if (!Boolean(derived)) return undefined

  return derived.downlinkAppFrameCount
}
export const selectDeviceDerivedNwkDownlinkFrameCount = (state, appId, devId) => {
  const derived = selectDeviceDerivedById(state, combineDeviceIds(appId, devId))
  if (!Boolean(derived)) return undefined

  return derived.downlinkNwkFrameCount
}

export const selectSelectedDeviceClaimable = state =>
  selectDeviceStore(state).selectedDeviceClaimable

// Devices.
const selectDevsIds = createPaginationIdsSelectorByEntity(ENTITY)
const selectDevsTotalCount = createPaginationTotalCountSelectorByEntity(ENTITY)

export const selectDevices = createSelector(
  [selectDevsIds, selectDeviceEntitiesStore],
  (ids, entities) => ids.map(id => entities[id]),
)
export const selectDevicesTotalCount = state => selectDevsTotalCount(state)

export const selectDevicesWithLastSeen = createSelector(
  [selectDevices, state => state],
  (devices, state) =>
    devices.map(device => ({
      ...device,
      _lastSeen: selectDeviceLastSeen(state, device.application_ids, device.ids.device_id),
    })),
)

// Events.
export const selectDeviceEvents = createEventsSelector(ENTITY)
export const selectDeviceEventsError = createEventsErrorSelector(ENTITY)
export const selectDeviceEventsStatus = createEventsStatusSelector(ENTITY)
export const selectDeviceEventsInterruptted = createEventsInterruptedSelector(ENTITY)
export const selectDeviceEventsPaused = createEventsPausedSelector(ENTITY)
export const selectDeviceEventsTruncated = createEventsTruncatedSelector(ENTITY)
export const selectDeviceEventsFilter = createEventsFilterSelector(ENTITY)
