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
  GET_DEV,
  UPDATE_DEV,
  GET_DEV_SUCCESS,
  GET_DEV_FAILURE,
} from '../actions/device'

const defaultState = {
  fetching: true,
  error: undefined,
  device: undefined,
}

const device = function (state = defaultState, action) {
  switch (action.type) {
  case GET_DEV:
    return {
      ...state,
      fetching: true,
      device: undefined,
      error: false,
    }
  case UPDATE_DEV:
    return {
      ...state,
      device: {
        ...state.device,
        ...action.patch,
      },
    }
  case GET_DEV_SUCCESS:
    return {
      ...state,
      fetching: false,
      error: false,
      device: action.device,
    }
  case GET_DEV_FAILURE:
    return {
      ...state,
      fetching: false,
      error: action.error,
      device: undefined,
    }
  default:
    return state
  }
}

export default device
