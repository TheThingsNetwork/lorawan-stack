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

import { GET_DEV_BASE } from '../actions/device'
import { GET_DEVICES_LIST_BASE } from '../actions/devices'

import {
  createEventsSelector,
  createEventsErrorSelector,
  createEventsStatusSelector,
} from './events'
import {
  createPaginationIdsSelectorByEntity,
  createPaginationTotalCountSelectorByEntity,
} from './pagination'

import { createFetchingSelector } from './fetching'
import { createErrorSelector } from './error'

const ENTITY = 'devices'

// Device
export const selectDeviceStore = state => state.devices
export const selectDeviceEntitiesStore = state => selectDeviceStore(state).entities
export const selectDeviceById = (state, id) => selectDeviceEntitiesStore(state)[id]
export const selectSelectedDeviceId = state => selectDeviceStore(state).selectedDevice
export const selectSelectedDevice = state => selectDeviceById(state, selectSelectedDeviceId(state))
export const selectSelectedDeviceFormatters = state => selectSelectedDevice(state).formatters
export const selectDeviceFetching = createFetchingSelector(GET_DEV_BASE)
export const selectDeviceError = createErrorSelector(GET_DEV_BASE)

// Devices
const selectDevsIds = createPaginationIdsSelectorByEntity(ENTITY)
const selectDevsTotalCount = createPaginationTotalCountSelectorByEntity(ENTITY)
const selectDevsFetching = createFetchingSelector(GET_DEVICES_LIST_BASE)
const selectDevsError = createErrorSelector(GET_DEVICES_LIST_BASE)

export const selectDevices = state => selectDevsIds(state).map(id => selectDeviceById(state, id))
export const selectDevicesTotalCount = state => selectDevsTotalCount(state)
export const selectDevicesFetching = state => selectDevsFetching(state)
export const selectDevicesError = state => selectDevsError(state)

// Events
export const selectDeviceEvents = createEventsSelector(ENTITY)
export const selectDeviceEventsError = createEventsErrorSelector(ENTITY)
export const selectDeviceEventsStatus = createEventsStatusSelector(ENTITY)
