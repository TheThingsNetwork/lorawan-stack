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

import { mergeWith, merge } from 'lodash'

import { getCombinedDeviceId, combineDeviceIds } from '@ttn-lw/lib/selectors/id'
import getByPath from '@ttn-lw/lib/get-by-path'

import {
  GET_DEV,
  GET_DEVICES_LIST_SUCCESS,
  GET_DEV_SUCCESS,
  UPDATE_DEV_SUCCESS,
} from '@console/store/actions/devices'
import { GET_APP_EVENT_MESSAGES_SUCCESS } from '@console/store/actions/applications'

const defaultState = {
  entities: {},
  derived: {},
  selectedDevice: undefined,
}

const heartbeatEvents = ['ns.up.data.receive', 'ns.up.join.receive', 'ns.up.rejoin.receive']
const uplinkFrameCountEventName = 'ns.up.data.process'

const mergeDerived = (state, id, derived) =>
  Object.keys(derived).length > 0
    ? merge({}, state, {
        derived: {
          [id]: derived,
        },
      })
    : state

const devices = function(state = defaultState, { type, payload, events, event }) {
  switch (type) {
    case GET_DEV:
      return {
        ...state,
        selectedDevice: combineDeviceIds(payload.appId, payload.deviceId),
      }
    case UPDATE_DEV_SUCCESS:
    case GET_DEV_SUCCESS:
      const updatedState = { ...state }
      const id = getCombinedDeviceId(payload)
      const mergedDevice = mergeWith({}, state.entities[id], payload, (_, __, key, ___, source) => {
        // Always set location from the payload.
        if (source === payload && key === 'locations') {
          return source.locations
        }

        if (source === payload && key === 'attributes') {
          return source.attributes
        }
      })

      updatedState.entities = {
        ...state.entities,
        [id]: mergedDevice,
      }

      // Update derived last seen value if possible.
      const { recent_uplinks } = payload
      const derived = {}
      if (recent_uplinks) {
        const last_uplink = Boolean(recent_uplinks)
          ? recent_uplinks[recent_uplinks.length - 1]
          : undefined
        if (last_uplink) {
          derived.lastSeen = last_uplink.received_at
        }
      }

      return mergeDerived(updatedState, id, derived)
    case GET_DEVICES_LIST_SUCCESS:
      const entities = payload.entities.reduce(
        function(acc, dev) {
          const id = getCombinedDeviceId(dev)

          acc[id] = dev
          return acc
        },
        { ...state.entities },
      )

      return {
        ...state,
        entities,
      }
    case GET_APP_EVENT_MESSAGES_SUCCESS: {
      // Detect heartbeat events to update last seen state.
      const derived = {}
      const id = events[0].identifiers[0].device_ids.device_id
      const heartbeatEvent = events.find(event => heartbeatEvents.includes(event.name))
      if (heartbeatEvent) {
        const receivedAt = getByPath(heartbeatEvent, 'data.received_at')
        if (receivedAt) {
          const currentDerived = state.derived[id]
          if (currentDerived) {
            // Only update if the event was actually more recent than the current value.
            if (currentDerived.lastSeen && currentDerived.lastSeen < receivedAt) {
              derived.lastSeen = receivedAt
            }
          } else {
            derived.lastSeen = receivedAt
          }
        }
      }

      // Detect uplink process event to update uplink frame count state.
      const uplinkFrameCountEvent = events.find(event => event.name === uplinkFrameCountEventName)
      if (uplinkFrameCountEvent) {
        derived.uplinkFrameCount = getByPath(
          uplinkFrameCountEvent,
          'data.payload.mac_payload.full_f_cnt',
        )
      }

      if (Object.keys(derived).length > 0) {
        return mergeDerived(state, id, derived)
      }

      return state
    }
    default:
      return state
  }
}

export default devices
