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

import {
  combineDeviceIds,
  extractApplicationIdFromCombinedId,
  extractDeviceIdFromCombinedId,
} from '@ttn-lw/lib/selectors/id'

import { getTypeAndId } from '@console/lib/recency-frequency-entities'

import { selectDeviceEntitiesStore } from './devices'
import { selectApplicationCombinedStore } from './applications'
import { selectGatewayEntitiesStore } from './gateways'
import { selectScoredRecencyFrequencyItems } from './recency-frequency-items'
import { selectBookmarksList } from './user-preferences'

const MAX_TOP_ENTITIES = 10

const getBookmarkType = bookmark => {
  const type = Object.keys(bookmark.entity_ids)[0].replace('_ids', '').toUpperCase()

  if (type === 'DEVICE') {
    return END_DEVICE
  }

  return type
}

const getEntityPath = (id, type) => {
  switch (type) {
    case APPLICATION:
      return `/applications/${id}`
    case END_DEVICE:
      return `/applications/${extractApplicationIdFromCombinedId(id)}/devices/${extractDeviceIdFromCombinedId(id)}`
    case GATEWAY:
      return `/gateways/${id}`
  }
}

const getBookmarkEntityId = bookmark => {
  const type = getBookmarkType(bookmark)
  switch (type) {
    case APPLICATION:
      return bookmark.entity_ids.application_ids.application_id
    case END_DEVICE:
      return combineDeviceIds(
        bookmark.entity_ids.device_ids.application_ids.application_id,
        bookmark.entity_ids.device_ids.device_id,
      )
    case GATEWAY:
      return bookmark.entity_ids.gateway_ids.gateway_id
  }
}

/*
 * Builds an entity object, by providing the id, type and state.
 * An entity object uses a streamlined object shape to represent an entity.
 * It is decorated with the respective entity record from the global store.
 *
 * @param {string} id - The entity id.
 * @param {string} type - The entity type.
 * @param {Object} state - The current state.
 * @param {string} source - The source of the entity.
 *
 * @returns {Entity} - The entity object or null if the connected entity is not found.
 */
const buildEntity = (id, type, source, entities) => ({
  id,
  type,
  path: getEntityPath(id, type),
  entity: entities?.[type]?.[id] || null,
  source,
})

const selectEntityMap = createSelector(
  [selectApplicationCombinedStore, selectDeviceEntitiesStore, selectGatewayEntitiesStore],
  (applications, devices, gateways) => ({
    [APPLICATION]: applications,
    [END_DEVICE]: devices,
    [GATEWAY]: gateways,
  }),
)

export const selectTopEntitiesStore = state => state.topEntities

export const selectTopEntities = createSelector(
  [selectEntityMap, selectScoredRecencyFrequencyItems, selectBookmarksList],
  (entities, topRecencyFrequencyItems, bookmarks) => {
    // Combine bookmarks with frequency items if bookmarks are exhausted.
    const entityTypes = [APPLICATION, GATEWAY, END_DEVICE, ALL]
    const items = entityTypes.reduce((acc, entityType) => {
      const entityRelevantBookmarks =
        entityType === ALL
          ? bookmarks
          : bookmarks.filter(bookmark => getBookmarkType(bookmark) === entityType)

      // Always put the bookmarks first.
      acc[entityType] = entityRelevantBookmarks.map(bookmark =>
        buildEntity(getBookmarkEntityId(bookmark), getBookmarkType(bookmark), 'bookmark', entities),
      )

      // Sort the bookmarks by their score in the recency frequency items.
      acc[entityType].sort((a, b) => {
        const aScore =
          topRecencyFrequencyItems[a.type].find(i => i.key === `${a.type}:${a.id}`)?.score || 0
        const bScore =
          topRecencyFrequencyItems[b.type].find(i => i.key === `${b.type}:${b.id}`)?.score || 0

        return bScore - aScore
      })

      // After the bookmarks, add frequency recency items.
      if (acc[entityType].length < MAX_TOP_ENTITIES) {
        const dedupedFrequencyRecencyItems = topRecencyFrequencyItems[entityType].filter(
          item => !acc[entityType].find(e => e.id === item.key.split(':')[1]),
        )
        acc[entityType].push(
          ...dedupedFrequencyRecencyItems
            .slice(0, MAX_TOP_ENTITIES - acc[entityType].length)
            .map(item => {
              const { entityType, entityId } = getTypeAndId(item)
              return buildEntity(entityId, entityType, 'frequency-recency-item', entities)
            }),
        )
      }

      // Fill up the rest with recently created entities.
      if (acc[entityType].length < MAX_TOP_ENTITIES) {
        // Concatenate entities from all entity types and add them to a new object sorted by created_at.
        const relevantEntities =
          entityType === ALL ? entities : { [entityType]: entities[entityType] }
        const allEntities = Object.keys(relevantEntities).reduce(
          (acc, entityType) =>
            acc.concat(
              Object.keys(entities[entityType]).map(id =>
                buildEntity(id, entityType, 'recently-created-entity', entities),
              ),
            ),
          [],
        )

        const sortedEntities = allEntities.sort((a, b) => {
          const aCreatedAt = a.entity?.created_at || 0
          const bCreatedAt = b.entity?.created_at || 0

          return bCreatedAt - aCreatedAt
        })

        acc[entityType].push(
          ...sortedEntities
            .filter(e => !acc[entityType].find(i => i.id === e.id))
            .slice(0, MAX_TOP_ENTITIES - acc[entityType].length),
        )
      }

      return acc
    }, {})

    return items
  },
)

const createTopEntityByTypeSelector = entityType =>
  createSelector([selectTopEntities, (_, filter) => filter], (topEntities, filter) => {
    if (!topEntities[entityType]) {
      return []
    }

    const filteredEntities = filter
      ? topEntities[entityType].filter(filter)
      : topEntities[entityType]

    return filteredEntities
  })

export const selectTopEntitiesAll = state => selectTopEntities(state)[ALL]

// Explicitly named selectors for each entity type.
export const selectApplicationTopEntities = createTopEntityByTypeSelector(APPLICATION)
export const selectEndDeviceTopEntities = createTopEntityByTypeSelector(END_DEVICE)
export const selectGatewayTopEntities = createTopEntityByTypeSelector(GATEWAY)

export const selectTopEntitiesLastFetched = state => selectTopEntitiesStore(state).lastFetched
