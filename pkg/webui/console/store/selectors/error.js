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

const selectErrorStore = state => state.ui.error

const getErrorStoreEntrySelector = (state, baseActionType) =>
  selectErrorStore(state)[baseActionType]

/**
 * @example
 * const selectError = createErrorSelector([
 * 'GET_ENTITY_LIST',
 * 'SEARCH_ENTITY_LIST'
 * ])
 * const selectEntityError = (state) => selectError(state)
 *
 * Creates the error selector for a set of base action types.
 * @param {Array} actions - A list of base action types or a single base action type.
 * @returns {Object} The error object matching one of the base action types.
 */
export const createErrorSelector = actions => function (state) {
  if (!Array.isArray(actions)) {
    return getErrorStoreEntrySelector(state, actions)
  }

  for (const action of actions) {

    const error = getErrorStoreEntrySelector(state, action)

    if (Boolean(error)) {
      return error
    }
  }
}
