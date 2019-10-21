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

const selectApiKeysStore = (state, props) => state[props.id] || {}

export const createApiKeysStoreSelector = entity => (state, props) =>
  selectApiKeysStore(state.apiKeys[entity], props) || {}

export const createApiKeysSelector = entity =>
  function(state, props) {
    const store = selectApiKeysStore(state.apiKeys[entity], props)
    console.log(store)

    return store.keys ? store.keys : []
  }

export const createApiKeySelector = function(entity) {
  const keysSelector = createApiKeysSelector(entity)

  return function(state, props) {
    const keys = keysSelector(state, props)

    return keys.find(key => key.id === props.keyId)
  }
}

export const createTotalCountSelector = entity =>
  function(state, props) {
    const store = selectApiKeysStore(state.apiKeys[entity], props)

    return store.totalCount
  }
