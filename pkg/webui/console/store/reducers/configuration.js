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
  GET_NS_FREQUENCY_PLANS,
  GET_NS_FREQUENCY_PLANS_FAILURE,
  GET_NS_FREQUENCY_PLANS_SUCCESS,
  GET_GS_FREQUENCY_PLANS,
  GET_GS_FREQUENCY_PLANS_FAILURE,
  GET_GS_FREQUENCY_PLANS_SUCCESS,
} from '../actions/configuration'

const defaultState = {
  fetching: false,
  error: undefined,
  nsFrequencyPlans: undefined,
  gsFrequencyPlans: undefined,
}

const configuration = function (state = defaultState, action) {
  switch (action.type) {
  case GET_NS_FREQUENCY_PLANS:
    return {
      ...state,
      fetching: true,
      nsFrequencyPlans: undefined,
    }
  case GET_NS_FREQUENCY_PLANS_SUCCESS:
    return {
      ...state,
      fetching: false,
      nsFrequencyPlans: action.frequencyPlans,
    }
  case GET_NS_FREQUENCY_PLANS_FAILURE:
    return {
      ...state,
      fetching: false,
      error: action.error,
    }
  case GET_GS_FREQUENCY_PLANS:
    return {
      ...state,
      fetching: true,
      gsFrequencyPlans: undefined,
    }
  case GET_GS_FREQUENCY_PLANS_SUCCESS:
    return {
      ...state,
      fetching: false,
      gsFrequencyPlans: action.frequencyPlans,
    }
  case GET_GS_FREQUENCY_PLANS_FAILURE:
    return {
      ...state,
      fetching: false,
      error: action.error,
    }
  default:
    return state
  }
}

export default configuration
