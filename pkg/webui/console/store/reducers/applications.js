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

import { EVENT_END_DEVICE_HEARTBEAT_FILTERS_REGEXP } from '@console/constants/event-filters'

import { getApplicationId } from '@ttn-lw/lib/selectors/id'
import getByPath from '@ttn-lw/lib/get-by-path'

import {
  GET_APP,
  GET_APP_SUCCESS,
  GET_APP_DEV_COUNT_SUCCESS,
  GET_APP_DEV_EUI_COUNT_SUCCESS,
  GET_APPS_LIST_SUCCESS,
  UPDATE_APP_SUCCESS,
  DELETE_APP_SUCCESS,
  GET_APP_EVENT_MESSAGE_SUCCESS,
  GET_MQTT_INFO_SUCCESS,
} from '@console/store/actions/applications'

const application = (state = {}, application) => ({
  ...state,
  ...application,
})

const defaultState = {
  entities: {},
  derived: {},
  selectedApplication: null,
  applicationDeviceCounts: {},
  mqtt: {},
}

const applications = (state = defaultState, { type, payload, event }) => {
  switch (type) {
    case GET_APP:
      return {
        ...state,
        selectedApplication: payload.id,
      }
    case GET_APPS_LIST_SUCCESS:
      const entities = payload.entities.reduce(
        (acc, app) => {
          const id = getApplicationId(app)

          acc[id] = application(acc[id], app)
          return acc
        },
        { ...state.entities },
      )

      return {
        ...state,
        entities,
      }
    case GET_APP_DEV_COUNT_SUCCESS:
      return {
        ...state,
        applicationDeviceCounts: {
          ...state.applicationDeviceCounts,
          [payload.id]: payload.applicationDeviceCount,
        },
      }
    case GET_APP_SUCCESS:
    case UPDATE_APP_SUCCESS:
      const id = getApplicationId(payload)

      return {
        ...state,
        entities: {
          ...state.entities,
          [id]: application(state.entities[id], payload),
        },
      }
    case GET_APP_DEV_EUI_COUNT_SUCCESS:
      return {
        ...state,
        entities: {
          ...state.entities,
          [payload.id]: {
            ...state.entities[payload.id],
            dev_eui_counter: payload.dev_eui_counter,
          },
        },
      }
    case DELETE_APP_SUCCESS:
      const { [payload.id]: deleted, ...rest } = state.entities

      return {
        ...defaultState,
        entities: rest,
      }
    case GET_APP_EVENT_MESSAGE_SUCCESS:
      if (EVENT_END_DEVICE_HEARTBEAT_FILTERS_REGEXP.test(event.name)) {
        const lastSeen = getByPath(event, 'data.received_at') || event.time
        const id = getApplicationId(event.identifiers[0].device_ids)

        // Update the application's derived last seen value, if the current
        // heartbeat event is more recent than the currently stored one.
        if (!(id in state.derived) || lastSeen > state.derived[id].lastSeen) {
          return {
            ...state,
            derived: {
              ...state.derived,
              [id]: {
                ...(state.derived[id] || {}),
                lastSeen,
              },
            },
          }
        }
      }
      return state
    case GET_MQTT_INFO_SUCCESS:
      return {
        ...state,
        mqtt: payload,
      }
    default:
      return state
  }
}

export default applications
