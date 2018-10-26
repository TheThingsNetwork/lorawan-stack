// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
  GET_APPS_LIST,
  SEARCH_APPS_LIST,
  GET_APPS_LIST_SUCCESS,
  GET_APPS_LIST_FAILURE,
  CHANGE_APPS_PAGE,
  CHANGE_APPS_TAB,
} from '../actions/applications'

const defaultState = {
  fetching: false,
  fetchingSearch: false,
  error: undefined,
  applications: [],
  totalCount: 0,
}

const applications = function (state = defaultState, action) {
  switch (action.type) {
  case GET_APPS_LIST:
    return {
      ...state,
      fetching: true,
    }
  case SEARCH_APPS_LIST:
    return {
      ...state,
      fetching: true,
      fetchingSearch: true,
    }
  case GET_APPS_LIST_SUCCESS:
    return {
      ...state,
      totalCount: action.totalCount,
      applications: action.applications,
      fetching: false,
      fetchingSearch: false,
    }
  case GET_APPS_LIST_FAILURE:
    return {
      ...state,
      fetching: false,
      fetchingSearch: false,
      error: action.error,
    }
  case CHANGE_APPS_PAGE:
    return {
      ...state,
      fetching: true,
    }
  case CHANGE_APPS_TAB:
    return {
      ...state,
      fetching: true,
    }
  default:
    return state
  }
}

export default applications
