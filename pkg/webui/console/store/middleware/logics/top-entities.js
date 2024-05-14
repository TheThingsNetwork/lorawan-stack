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

import { APPLICATION, END_DEVICE, GATEWAY, ORGANIZATION } from '@console/constants/entities'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import {
  combineDeviceIds,
  extractApplicationIdFromCombinedId,
  extractDeviceIdFromCombinedId,
} from '@ttn-lw/lib/selectors/id'

import { getTopFrequencyRecencyItems, getTypeAndId } from '@console/lib/frequently-visited-entities'

import * as actions from '@console/store/actions/top-entities'
import { getAllBookmarks } from '@console/store/actions/user-preferences'
import { getApplicationsList } from '@console/store/actions/applications'
import { getOrganizationsList } from '@console/store/actions/organizations'
import { getGatewaysList } from '@console/store/actions/gateways'
import { getDevicesList } from '@console/store/actions/devices'

import { selectUserId } from '@account/store/selectors/user'
import { selectApplicationById } from '@console/store/selectors/applications'
import { selectGatewayById } from '@console/store/selectors/gateways'
import { selectOrganizationById } from '@console/store/selectors/organizations'
import { selectDeviceByIds } from '@console/store/selectors/devices'
import { selectTopEntitiesLastFetched } from '@console/store/selectors/top-entities'

const MAX_ENTITIES = 15

const getBookmarkType = bookmark =>
  Object.keys(bookmark.entity_ids)[0].replace('_ids', '').toUpperCase()
const getEntityPath = (id, type) => {
  switch (type) {
    case APPLICATION:
      return `/applications/${id}`
    case END_DEVICE:
      return `/applications/${extractApplicationIdFromCombinedId(id)}/devices/${extractDeviceIdFromCombinedId(id)}`
    case GATEWAY:
      return `/gateways/${id}`
    case ORGANIZATION:
      return `/organizations/${id}`
  }
}
const getEntityName = (id, type, state) => {
  switch (type) {
    case APPLICATION:
      return selectApplicationById(state, id)?.name
    case END_DEVICE:
      return selectDeviceByIds(
        state,
        extractApplicationIdFromCombinedId(id),
        extractDeviceIdFromCombinedId(id),
      )?.name
    case GATEWAY:
      return selectGatewayById(state, id)?.name
    case ORGANIZATION:
      return selectOrganizationById(state, id)?.name
    default:
      return ''
  }
}
const getBookmarkEntityId = bookmark => {
  const type = getBookmarkType(bookmark)
  switch (type) {
    case APPLICATION:
      return bookmark.entity_ids.application_ids.application_id
    case END_DEVICE:
      return combineDeviceIds(
        bookmark.entity_ids.end_device_ids.application_ids.application_id,
        bookmark.entity_ids.end_device_ids.device_id,
      )
    case GATEWAY:
      return bookmark.entity_ids.gateway_ids.gateway_id
    case ORGANIZATION:
      return bookmark.entity_ids.organization_ids.organization_id
  }
}

const getTopEntitiesLogic = createRequestLogic({
  type: actions.GET_TOP_ENTITIES,
  process: async ({ getState }, dispatch) => {
    const state = getState()
    const limit = 100
    const order = '-created_at'
    const lastFetched = selectTopEntitiesLastFetched(state)

    // Only refetch entity names every 5 minutes.
    const shouldRefetch = lastFetched && Date.now() - lastFetched < 1000 * 60 * 5 // 5 minutes

    if (!shouldRefetch) {
      // Fetch 100 items of all entity types to have some initial data in the store
      // to source the entity names from. This is a best effort to avoid making
      // requests for every single entity.
      await Promise.all([
        dispatch(attachPromise(getApplicationsList({ page: 1, limit, order }, ['name']))),
        dispatch(attachPromise(getGatewaysList({ page: 1, limit, order }, ['name']))),
        dispatch(attachPromise(getOrganizationsList({ page: 1, limit, order }, ['name']))),
      ])
    }

    const topEntities = []
    const topFrequencyRecencyItems = getTopFrequencyRecencyItems()
    const userId = selectUserId(state)

    const { bookmarks } = await dispatch(attachPromise(getAllBookmarks(userId)))

    // Get the top 3 frequency recency items that are applications.
    const topFrequencyRecencyApplicationIds = topFrequencyRecencyItems
      .filter(item => item.key.startsWith(APPLICATION))
      .map(getTypeAndId)
      .slice(0, 3)

    if (!shouldRefetch) {
      // Fetch the 1000 last registered devices for the top 3 applications.
      // This is a best effort to get the device names for the top entities.
      await Promise.all(
        topFrequencyRecencyApplicationIds.map(({ entityId }) =>
          dispatch(
            attachPromise(getDevicesList(entityId, { page: 1, limit: 1000, order }, ['name'])),
          ),
        ),
      )
    }

    // Always put the bookmarks first.
    topEntities.push({
      category: 'bookmarks',
      source: 'top-entities',
      items: bookmarks.slice(0, MAX_ENTITIES).map(bookmark => {
        const id = getBookmarkEntityId(bookmark)
        const type = getBookmarkType(bookmark)
        return {
          id: getBookmarkEntityId(bookmark),
          name: getEntityName(id, type, state),
          type: getBookmarkType(bookmark),
          path: getEntityPath(id, type, state),
        }
      }),
    })

    // Then add the top frequency recency items, but only if we have not already.
    const slicedTopFrequencyRecencyItems = topFrequencyRecencyItems
      .slice(0, MAX_ENTITIES - bookmarks.length)
      .reduce((acc, item) => {
        const { entityType, entityId } = getTypeAndId(item)
        const path = getEntityPath(entityId, entityType, state)
        // Skip if the entity is already in the list.
        if (topEntities[0].items.find(e => e.path === path)) {
          return acc
        }
        acc.push({
          id: entityId,
          name: getEntityName(entityId, entityType, state),
          type: entityType,
          path,
        })
        return acc
      }, [])

    topEntities.push({
      category: 'recency',
      source: 'top-entities',
      items: slicedTopFrequencyRecencyItems,
    })

    return topEntities
  },
})

export default [getTopEntitiesLogic]
