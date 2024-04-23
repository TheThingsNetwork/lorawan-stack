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

// This module keeps track of the entities that are frequently visited by the user
// and stores it in the local storage. We use the Recency and Frequency model to
// determine the entities that are frequently visited by the user. The model is
// based on the assumption that the more frequently and recently an entity is
// visited, the more important it is to the user.

const FREQUENTLY_VISITED_ENTITIES_KEY = 'frequentlyVisitedEntities'
const MAX_FREQUENTLY_VISITED_ENTITIES = 5

const getFrequentlyVisitedEntities = () =>
  JSON.parse(localStorage.getItem(FREQUENTLY_VISITED_ENTITIES_KEY)) || {}

const setFrequentlyVisitedEntities = entities => {
  localStorage.setItem(FREQUENTLY_VISITED_ENTITIES_KEY, JSON.stringify(entities))
}

const getTypeAndId = item => {
  const [entityType, entityId] = item.key.split(':')

  return { entityType, entityId }
}

const trackEntityAccess = (entityType, entityId) => {
  const storedData = getFrequentlyVisitedEntities()
  const entityKey = `${entityType}:${entityId}`
  const entity = storedData[entityKey]

  if (entity) {
    storedData[entityKey] = {
      ...entity,
      frequency: entity.frequency + 1,
      lastAccessed: Date.now(),
    }
  }

  if (!entity) {
    storedData[entityKey] = {
      frequency: 1,
      lastAccessed: Date.now(),
    }
  }

  setFrequentlyVisitedEntities(storedData)
}

const calculateScore = entity => {
  const now = Date.now()
  const recency = now - entity.lastAccessed
  const decay = Math.exp(-recency / 1000)

  return entity.frequency * decay
}

const getTopFrequencyRecencyItems = () => {
  const storedData = getFrequentlyVisitedEntities()
  const entities = Object.keys(storedData).map(key => ({
    key,
    score: calculateScore(storedData[key]),
  }))

  entities.sort((a, b) => b.score - a.score)

  return entities.slice(0, MAX_FREQUENTLY_VISITED_ENTITIES)
}

export { trackEntityAccess, getTopFrequencyRecencyItems, getTypeAndId }
