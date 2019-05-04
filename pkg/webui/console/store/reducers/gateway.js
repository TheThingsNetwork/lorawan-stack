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

import {
  GET_GTW,
  GET_GTW_SUCCESS,
  GET_GTW_FAILURE,
  START_GTW_STATS,
  UPDATE_GTW_STATS,
  UPDATE_GTW_STATS_SUCCESS,
  UPDATE_GTW_STATS_FAILURE,
  UPDATE_GTW_STATS_UNAVAILABLE,
  STOP_GTW_STATS,
} from '../actions/gateway'

const statsDefaultState = {
  fetching: false,
  available: true,
  error: undefined,
  stats: undefined,
}

const defaultState = {
  fetching: false,
  error: undefined,
  gateway: undefined,
  statistics: statsDefaultState,
}

const statistics = function (state = statsDefaultState, action) {
  switch (action.type) {
  case START_GTW_STATS:
    return {
      ...state,
      fetching: true,
      error: undefined,
    }
  case UPDATE_GTW_STATS_SUCCESS:
    return {
      ...state,
      fetching: false,
      available: true,
      error: undefined,
      stats: action.statistics,
    }
  case UPDATE_GTW_STATS_FAILURE:
    return {
      ...state,
      fetching: false,
      error: action.error,
    }
  case UPDATE_GTW_STATS_UNAVAILABLE:
    return {
      ...state,
      fetching: false,
      available: false,
    }
  default:
    return state
  }
}

const gateway = function (state = defaultState, action) {
  switch (action.type) {
  case GET_GTW:
    return {
      ...state,
      fetching: true,
      gateway: undefined,
      error: undefined,
    }
  case GET_GTW_SUCCESS:
    return {
      ...state,
      fetching: false,
      gateway: action.gateway,
    }
  case GET_GTW_FAILURE:
    return {
      ...state,
      fetching: false,
      error: action.error,
    }
  case START_GTW_STATS:
  case UPDATE_GTW_STATS:
  case UPDATE_GTW_STATS_SUCCESS:
  case UPDATE_GTW_STATS_FAILURE:
  case UPDATE_GTW_STATS_UNAVAILABLE:
  case STOP_GTW_STATS:
    return {
      ...state,
      statistics: statistics(state.statistics, action),
    }
  default:
    return state
  }
}

export default gateway
