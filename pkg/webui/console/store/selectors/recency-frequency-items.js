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

import { createSelector } from 'reselect'

import { ALL, APPLICATION, END_DEVICE, GATEWAY } from '@console/constants/entities'

const entityTypes = [APPLICATION, GATEWAY, END_DEVICE, ALL]
const MAX_FREQUENTLY_VISITED_ENTITIES = 12

const selectRecencyFrequencyStore = state => state.recencyFrequencyItems

const calculateScore = entity => {
  const now = Date.now()
  const recency = now - entity.lastAccessed
  const decay = Math.exp(-recency / 1000)

  return entity.frequency * decay
}

export const selectRecencyFrequencyItems = state => selectRecencyFrequencyStore(state).items

export const selectScoredRecencyFrequencyItems = createSelector(
  [selectRecencyFrequencyItems],
  storedData =>
    entityTypes.reduce((acc, entityType) => {
      acc[entityType] = Object.keys(storedData)
        .filter(key => key.startsWith(entityType === ALL ? '' : entityType))
        .map(key => ({
          key,
          score: calculateScore(storedData[key]),
        }))
        .sort((a, b) => b.score - a.score)
        .slice(0, MAX_FREQUENTLY_VISITED_ENTITIES)
      return acc
    }, {}),
)
