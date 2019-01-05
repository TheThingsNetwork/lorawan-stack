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
  INITIALIZE,
  INITIALIZE_FAILURE,
  INITIALIZE_SUCCESS,
} from '../actions/console'

const c = function (state = [], action) {
  switch (action.type) {
  case INITIALIZE:
    return Object.assign({}, state, { initialized: false })
  case INITIALIZE_SUCCESS:
    return Object.assign({}, state, { initialized: true })
  case INITIALIZE_FAILURE:
    return Object.assign({}, state, { initialized: false })
  default:
    return state
  }
}

export default c
