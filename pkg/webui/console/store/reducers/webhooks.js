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
  GET_WEBHOOKS_LIST,
  GET_WEBHOOKS_LIST_FAILURE,
  GET_WEBHOOKS_LIST_SUCCESS,
} from '../actions/webhooks'

const defaultState = {
  fetching: false,
  error: undefined,
  webhooks: undefined,
}

const webhooks = function (state = defaultState, action) {
  switch (action.type) {
  case GET_WEBHOOKS_LIST:
    return {
      ...state,
      fetching: true,
      webhooks: undefined,
    }
  case GET_WEBHOOKS_LIST_SUCCESS:
    return {
      ...state,
      fetching: false,
      webhooks: action.webhooks,
    }
  case GET_WEBHOOKS_LIST_FAILURE:
    return {
      ...state,
      fetching: false,
      error: action.error,
    }
  default:
    return state
  }
}

export default webhooks
