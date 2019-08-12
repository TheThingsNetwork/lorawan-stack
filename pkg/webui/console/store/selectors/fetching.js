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

/* eslint-disable import/prefer-default-export */

const selectFetchingStore = state => state.ui.fetching

const selectFetchingEntry = (state, id) => selectFetchingStore(state)[id] || false

/**
 * @example
 * const selectFetching = createFetchingSelector([
 * 'GET_ENTITY_LIST',
 * 'SEARCH_ENTITY_LIST'
 * ])
 * const selectEntityFetching = (state) => selectFetching(state)
 *
 * Creates the fetching selector for a set of base action types.
 * @param {Array} actions - A list of base action types or a single base action type.
 * @returns {boolean} `true` if one of the base action types is in the `fetching` state,
 * `false` otherwise.
 */
export const createFetchingSelector = actions =>
  function(state) {
    if (!Array.isArray(actions)) {
      return selectFetchingEntry(state, actions)
    }

    return actions.some(action => selectFetchingEntry(state, action))
  }
