// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import { handleActions } from 'redux-actions'

import {
  GET_GLOBAL_SEARCH_RESULTS_SUCCESS,
  SET_SEARCH_OPEN,
  SET_SEARCH_SCOPE,
} from '@console/store/actions/search'

const defaultState = {
  searchOpen: false,
  results: [],
  query: '',
  scope: undefined,
}

export default handleActions(
  {
    [SET_SEARCH_OPEN]: (state, { payload: { searchOpen } }) => ({
      ...state,
      searchOpen,
    }),
    [SET_SEARCH_SCOPE]: (state, { payload: { scope } }) => ({
      ...state,
      scope,
    }),
    [GET_GLOBAL_SEARCH_RESULTS_SUCCESS]: (state, { payload: { query, results } }) => ({
      ...state,
      results,
      query,
    }),
  },
  defaultState,
)
