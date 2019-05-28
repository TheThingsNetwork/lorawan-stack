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
  GET_APP,
  GET_APP_SUCCESS,
  GET_APP_FAILURE,
} from '../actions/application'

const defaultState = {
  fetching: false,
  error: undefined,
  application: undefined,
  link: {},
}

const application = function (state = defaultState, action) {
  switch (action.type) {
  case GET_APP:
    return {
      ...state,
      fetching: true,
      application: undefined,
      error: undefined,
    }
  case GET_APP_SUCCESS:
    return {
      ...state,
      fetching: false,
      application: action.application,
    }
  case GET_APP_FAILURE:
    return {
      ...state,
      fetching: false,
      error: action.error,
    }
  default:
    return state
  }
}

export default application
