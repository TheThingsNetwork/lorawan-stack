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

import { getGatewayId } from '@ttn-lw/lib/selectors/id'

import {
  GET_GTW,
  GET_GTW_SUCCESS,
  UPDATE_GTW_SUCCESS,
  UPDATE_GTW_LOCATION_SUCCESS,
  DELETE_GTW_SUCCESS,
  GET_GTWS_LIST_SUCCESS,
  UPDATE_GTW_STATS,
  UPDATE_GTW_STATS_SUCCESS,
  UPDATE_GTW_STATS_FAILURE,
  START_GTW_STATS_SUCCESS,
  START_GTW_STATS_FAILURE,
  FETCH_GTWS_LIST_SUCCESS,
} from '@console/store/actions/gateways'

const defaultStatisticsState = {
  error: undefined,
  stats: undefined,
}

const defaultState = {
  entities: {},
  selectedGateway: null,
  statistics: defaultStatisticsState,
}

const gateway = (state = {}, gateway) => ({
  ...state,
  ...gateway,
})

const statistics = (state = defaultStatisticsState, { type, payload }) => {
  switch (type) {
    case UPDATE_GTW_STATS_SUCCESS:
      return {
        ...state,
        error: undefined,
        stats: payload.stats,
      }
    case UPDATE_GTW_STATS_FAILURE:
    case START_GTW_STATS_FAILURE:
      return {
        ...state,
        stats: undefined,
        error: payload,
      }
    case UPDATE_GTW_STATS:
    case START_GTW_STATS_SUCCESS:
      return {
        ...state,
        error: undefined,
      }
    default:
      return state
  }
}

const gateways = (state = defaultState, action) => {
  const { type, payload, meta } = action

  switch (type) {
    case GET_GTW:
      return {
        ...state,
        statistics: defaultStatisticsState,
        selectedGateway: meta.options.noSelect ? state.selectedGateway : payload.id,
      }
    case GET_GTW_SUCCESS:
    case UPDATE_GTW_SUCCESS:
      const id = getGatewayId(payload)

      return {
        ...state,
        entities: {
          ...state.entities,
          [id]: gateway(state.entities[id], payload),
        },
      }
    case UPDATE_GTW_LOCATION_SUCCESS: {
      const { id } = payload
      const antennaLocations = payload.event.data.antenna_locations

      const composedLocations = antennaLocations.map(antennaLocation => ({
        location: {
          ...antennaLocation,
          // Locations from status messages can currently not be trusted
          // in terms of integrity since they are not sent over a secure connection.
          trusted: false,
        },
      }))

      return {
        ...state,
        entities: {
          ...state.entities,
          [id]: {
            ...state.entities[id],
            antennas: composedLocations,
          },
        },
      }
    }
    case DELETE_GTW_SUCCESS:
      const { [payload.id]: deleted, ...rest } = state.entities

      return {
        ...state,
        selectedGateway: null,
        entities: rest,
      }
    case FETCH_GTWS_LIST_SUCCESS:
    case GET_GTWS_LIST_SUCCESS:
      const entities = payload.entities.reduce(
        (acc, gtw) => {
          const id = getGatewayId(gtw)

          acc[id] = gateway(acc[id], gtw)
          return acc
        },
        { ...state.entities },
      )

      return {
        ...state,
        entities,
      }
    case START_GTW_STATS_SUCCESS:
    case START_GTW_STATS_FAILURE:
    case UPDATE_GTW_STATS:
    case UPDATE_GTW_STATS_SUCCESS:
    case UPDATE_GTW_STATS_FAILURE:
      return {
        ...state,
        statistics: statistics(state.statistics, action),
      }
    default:
      return state
  }
}

export default gateways
