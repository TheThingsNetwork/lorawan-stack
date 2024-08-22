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

import { chunk } from 'lodash'

import { APPLICATION, END_DEVICE, GATEWAY } from '@console/constants/entities'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { isNotFoundError, isPermissionDeniedError } from '@ttn-lw/lib/errors/utils'
import {
  extractApplicationIdFromCombinedId,
  extractDeviceIdFromCombinedId,
} from '@ttn-lw/lib/selectors/id'

import { getTypeAndId } from '@console/lib/recency-frequency-entities'

import { GET_TOP_ENTITIES } from '@console/store/actions/top-entities'
import { getBookmarksList } from '@console/store/actions/user-preferences'
import { fetchApplicationsList, getApplication } from '@console/store/actions/applications'
import { fetchGatewaysList, getGateway } from '@console/store/actions/gateways'
import { fetchDevicesList, getDevice } from '@console/store/actions/devices'

import { selectUserId } from '@console/store/selectors/user'
import {
  selectTopEntities,
  selectTopEntitiesLastFetched,
} from '@console/store/selectors/top-entities'
import { selectScoredRecencyFrequencyItems } from '@console/store/selectors/recency-frequency-items'

const requestMap = {
  [APPLICATION]: getApplication,
  [GATEWAY]: getGateway,
  [END_DEVICE]: getDevice,
}

/** Limit the number of concurrent promises.
 * @param {Array<Function>} promises - An array of functions that return promises.
 * @param {number} limit - The number of promises to run concurrently.
 * @returns {Promise<Array>} - A promise that resolves to an array of results.
 */
const limitedPromiseAll = async (promises, limit) => {
  const chunks = chunk(promises, limit)
  const results = []

  for (const chunk of chunks) {
    /* eslint-disable no-await-in-loop */
    const chunkResults = await Promise.all(chunk.map(p => p()))
    results.push(...chunkResults)
  }

  return results
}

/** Fetch missing entities.
 * @param {object} topEntities - The top entities.
 * @param {Function} dispatch - The dispatch function.
 * @returns {Promise<Array>} - A promise that resolves to an array of results.
 */
const fetchMissingEntities = (topEntities, dispatch) => {
  const missingEntities = []
  Object.values(topEntities).forEach(entities => {
    entities.forEach(entity => {
      if (!entity.entity && !missingEntities.some(({ path }) => path === entity.path)) {
        missingEntities.push(entity)
      }
    })
  })

  // Fetch the missing entities but limit the number of concurrent requests to 5.
  return limitedPromiseAll(
    missingEntities.map(({ id, type }) => async () => {
      try {
        const idParams =
          type === END_DEVICE
            ? [extractApplicationIdFromCombinedId(id), extractDeviceIdFromCombinedId(id)]
            : [id]
        await dispatch(
          attachPromise(
            requestMap[type](...idParams, ['name'], { noSelect: true, startStream: false }),
          ),
        )
      } catch (error) {
        if (isPermissionDeniedError(error)) {
          // Permission denied is not necessarily conclusive that the entity cannot be
          // accessed at all, since it's possible that only certain parts of the entity
          // are not accessible. In any case, it should not cause the logic to fail.
          return
        }
        if (!isNotFoundError(error)) {
          throw error
        }
      }
    }),
    5,
  )
}

// In the Console, top entities refer to a collection of items that
// are deemed important to the user. These items are a mix of bookmarks
// which are stored in the backend and frequency recency items which are
// calculated based on the user's usage of the Console and stored in the
// local storage.
// Since for these entities, only the type and id are stored, we need to
// fetch the entity names from the backend to display them in the UI.
// This should however be done in a way that is most efficient and does
// overload the backend.
// Since the top entities are entirely a derivative of other store items,
// the composition is done on the selector level.
const fetchTopEntities = async (getState, dispatch) => {
  const state = getState()
  const prefetchLimit = 100
  const order = '-created_at'
  const lastFetched = selectTopEntitiesLastFetched(state)

  // Only refetch entity names every 5 minutes.
  const shouldRefetch = lastFetched && Date.now() - lastFetched < 1000 * 60 * 5 // 5 minutes.

  if (!shouldRefetch) {
    // Fetch 100 items of all entity types to have some initial data in the store
    // to source the entity names from. This is a best effort to avoid making
    // requests for every single entity.
    try {
      await Promise.all([
        dispatch(
          attachPromise(
            fetchApplicationsList({ page: 1, limit: prefetchLimit, order }, ['name'], {
              withDeviceCount: true,
            }),
          ),
        ),
        dispatch(
          attachPromise(
            fetchGatewaysList(
              { page: 1, limit: prefetchLimit, order },
              ['name', 'gateway_server_address'],
              {
                withStatus: true,
              },
            ),
          ),
        ),
      ])
    } catch (error) {
      // Ignore any errors that occur during the prefetch.
      // The entity names will not be displayed in the worst case,
      // but the logic should not fail.
    }
  }

  const topRecencyFrequencyItems = selectScoredRecencyFrequencyItems(state)
  const userId = selectUserId(state)

  await dispatch(attachPromise(getBookmarksList(userId, { page: 1, limit: 100 })))

  // Get the top 3 frequency recency items that are applications.
  const topRecencyFrequencyApplicationIds = topRecencyFrequencyItems.APPLICATION.slice(0, 3).map(
    getTypeAndId,
  )

  if (!shouldRefetch) {
    // Fetch the 1000 last registered devices for the top 3 applications.
    // This is a best effort to get the device names for the top entities.
    await Promise.all(
      topRecencyFrequencyApplicationIds.map(async ({ entityId }) => {
        try {
          return await dispatch(
            attachPromise(
              fetchDevicesList(entityId, { page: 1, limit: 1000, order }, ['name', 'last_seen_at']),
            ),
          )
        } catch (error) {
          // Ignore any errors that occur during the prefetch.
          // The entity names will not be displayed in the worst case,
          // but the logic should not fail.
        }
      }),
    )
  }

  const topEntities = selectTopEntities(getState())

  // Fetch the missing entities.
  await fetchMissingEntities(topEntities, dispatch)

  // Get the composed top entities.
  return topEntities
}

// This promise is used to prevent multiple concurrent requests for the top entities.
let topEntitiesPromise = null

const getTopEntitiesLogic = createRequestLogic({
  type: GET_TOP_ENTITIES,
  process: async ({ getState }, dispatch) => {
    if (topEntitiesPromise) {
      return topEntitiesPromise
    }

    topEntitiesPromise = fetchTopEntities(getState, dispatch)
    const result = await topEntitiesPromise

    topEntitiesPromise = null

    return result
  },
})

export default [getTopEntitiesLogic]
