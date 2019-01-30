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
  GET_USER_ME,
  GET_USER_ME_FAILURE,
  GET_USER_ME_SUCCESS,
  LOGOUT,
  LOGOUT_SUCCESS,
} from '../actions/user'

const defaultState = {
  fetching: false,
  user: undefined,
}

const user = function (state = defaultState, action) {
  switch (action.type) {
  case GET_USER_ME:
    return Object.assign({}, state, {
      fetching: true,
    })
  case GET_USER_ME_SUCCESS:
    return Object.assign({}, state, {
      fetching: false,
      user: action.userData.user,
    })
  case GET_USER_ME_FAILURE:
    return Object.assign({}, state, {
      fetching: false,
      user: undefined,
    })
  case LOGOUT:
    return Object.assign({}, state, {
      fetching: false,
    })
  case LOGOUT_SUCCESS:
    return Object.assign({}, state, {
      user: undefined,
    })
  default:
    return state
  }
}

export default user
