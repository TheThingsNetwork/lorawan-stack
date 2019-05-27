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
  GET_WEBHOOK_FORMATS,
  GET_WEBHOOK_FORMATS_FAILURE,
  GET_WEBHOOK_FORMATS_SUCCESS,
} from '../actions/webhook-formats'

const defaultState = {
  fetching: false,
  error: undefined,
  formats: undefined,
}

const webhooks = function (state = defaultState, action) {
  switch (action.type) {
  case GET_WEBHOOK_FORMATS:
    return {
      ...state,
      fetching: true,
      formats: undefined,
    }
  case GET_WEBHOOK_FORMATS_SUCCESS:
    return {
      ...state,
      fetching: false,
      formats: action.formats,
    }
  case GET_WEBHOOK_FORMATS_FAILURE:
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
