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
  GET_GTWS_LIST,
  SEARCH_GTWS_LIST,
  GET_GTWS_LIST_SUCCESS,
  GET_GTWS_LIST_FAILURE,
} from '../actions/gateways'

const defaultState = {
  fetching: false,
  fetchingSearch: false,
  error: undefined,
  gateways: [],
  totalCount: 0,
}

const gateways = function (state = defaultState, action) {
  switch (action.type) {
  case GET_GTWS_LIST:
    return {
      ...state,
      fetching: true,
    }
  case SEARCH_GTWS_LIST:
    return {
      ...state,
      fetching: true,
      fetchingSearch: true,
    }
  case GET_GTWS_LIST_SUCCESS:
    return {
      ...state,
      totalCount: action.totalCount,
      gateways: action.gateways,
      fetching: false,
      fetchingSearch: false,
    }
  case GET_GTWS_LIST_FAILURE:
    return {
      ...state,
      fetching: false,
      fetchingSearch: false,
      error: action.error,
    }
  default:
    return state
  }
}

export default gateways
