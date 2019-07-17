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

import { getGatewayId } from '../../../lib/selectors/id'
import {
  GET_GTW,
  GET_GTW_SUCCESS,
  UPDATE_GTW_SUCCESS,
  DELETE_GTW_SUCCESS,
  GET_GTWS_LIST_SUCCESS,
  UPDATE_GTW_STATS,
  UPDATE_GTW_STATS_SUCCESS,
  UPDATE_GTW_STATS_FAILURE,
  START_GTW_STATS_FAILURE,
} from '../actions/gateways'

const defaultState = {
  entities: {},
  selectedGateway: null,
  statistics: {},
}

const gateway = function (state = {}, gateway) {
  return {
    ...state,
    ...gateway,
  }
}

const statistics = function (state = defaultState.statistics, { type, payload }) {
  const { id } = payload
  const stats = state[id] || {}

  switch (type) {
  case UPDATE_GTW_STATS_SUCCESS:
    return {
      ...state,
      [id]: {
        ...stats,
        error: undefined,
        stats: payload.stats,
      },
    }
  case UPDATE_GTW_STATS_FAILURE:
    return {
      ...state,
      [id]: {
        ...stats,
        error: payload.error,
      },
    }
  case UPDATE_GTW_STATS:
    return {
      ...state,
      [id]: {
        ...stats,
        error: undefined,
      },
    }
  default:
    return state
  }
}

const gateways = function (state = defaultState, action) {
  const { type, payload } = action

  switch (type) {
  case GET_GTW:
    return {
      ...state,
      selectedGateway: payload.id,
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
  case DELETE_GTW_SUCCESS:
    const { [payload.id]: deleted, ...rest } = state.entities

    return {
      selectedGateway: null,
      entities: rest,
    }
  case GET_GTWS_LIST_SUCCESS:
    const entities = payload.entities.reduce(function (acc, gtw) {
      const id = getGatewayId(gtw)

      acc[id] = gateway(acc[id], gtw)
      return acc
    }, { ...state.entities })

    return {
      ...state,
      entities,
    }
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
