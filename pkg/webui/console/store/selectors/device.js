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
import { getDeviceId } from '../../../lib/selectors/id'

import {
  eventsSelector,
  errorSelector as eventsErrorSelector,
  statusSelector as eventsStatusSelector,
} from './events'

import { createFetchingSelector } from './fetching'
import { createErrorSelector } from './error'

const ENTITY = 'devices'

const storeSelector = store => store.device

export const deviceSelector = state => storeSelector(state).device

export const selectDeviceFetching = createFetchingSelector(GET_DEV_BASE)
export const selectDeviceError = createErrorSelector(GET_DEV_BASE)
export const selectSelectedDevice = state => deviceSelector(state)

export const selectSelectedDeviceId = state => getDeviceId(selectSelectedDevice(state))

export const selectSelectedDeviceFormatters = state => selectSelectedDevice(state).formatters

export const fetchingSelector = function (state) {
  const store = storeSelector(state)

  return store.fetching || false
}

export const errorSelector = function (state) {
  const store = storeSelector(state)

  return store.error
}

export const deviceEventsSelector = eventsSelector(ENTITY)

export const deviceEventsErrorSelector = eventsErrorSelector(ENTITY)

export const deviceEventsStatusSelector = eventsStatusSelector(ENTITY)
