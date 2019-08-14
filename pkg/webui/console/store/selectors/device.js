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

import { GET_DEV_BASE, UPDATE_DEV_BASE } from '../actions/device'
import { getDeviceId } from '../../../lib/selectors/id'

import {
  createEventsSelector,
  createEventsErrorSelector,
  createEventsStatusSelector,
} from './events'

import { createFetchingSelector } from './fetching'
import { createErrorSelector } from './error'

const ENTITY = 'devices'

const selectDeviceStore = store => store.device

// Device Entity
export const selectSelectedDevice = state => selectDeviceStore(state).device
export const selectSelectedDeviceId = state => getDeviceId(selectSelectedDevice(state))
export const selectDeviceFetching = createFetchingSelector(GET_DEV_BASE)
export const selectGetDeviceError = createErrorSelector(GET_DEV_BASE)
export const selectUpdateDeviceError = createErrorSelector(UPDATE_DEV_BASE)
export const selectDeviceError = createErrorSelector([GET_DEV_BASE, UPDATE_DEV_BASE])
export const selectSelectedDeviceFormatters = state => selectSelectedDevice(state).formatters

// Events
export const selectDeviceEvents = createEventsSelector(ENTITY)
export const selectDeviceEventsError = createEventsErrorSelector(ENTITY)
export const selectDeviceEventsStatus = createEventsStatusSelector(ENTITY)
