// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

const selectTopEntitiesStore = state => state.topEntities

export const selectTopEntitiesData = state => selectTopEntitiesStore(state).data
export const selectConcatenatedTopEntitiesData = state =>
  selectTopEntitiesData(state).reduce((acc, entity) => acc.concat(entity.items), [])
export const selectConcatenatedTopEntitiesByType = (state, type) =>
  selectConcatenatedTopEntitiesData(state).filter(entity => entity.type === type)
export const selectTopEntitiesLastFetched = state => selectTopEntitiesStore(state).lastFetched
