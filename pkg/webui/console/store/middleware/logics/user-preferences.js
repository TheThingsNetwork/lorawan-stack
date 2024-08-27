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

import tts from '@console/api/tts'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as userPreferences from '@console/store/actions/user-preferences'

const addBookmarkToResult = (result, entity, element) => {
  if (!(entity in result.perEntityBookmarks)) {
    result.perEntityBookmarks[entity] = []
    result.perEntityBookmarks[entity].push(element)
  } else {
    result.perEntityBookmarks[entity].push(element)
  }
  if (!(entity in result.totalCount.perEntityTotalCount)) {
    result.totalCount.perEntityTotalCount[entity] = 1
  } else {
    result.totalCount.perEntityTotalCount[entity] += 1
  }

  return result
}

const getBookmarksThroughPagination = async userId => {
  let page = 1
  const limit = 1000
  let totalCount = Infinity
  let result = {
    bookmarks: [],
    perEntityBookmarks: {},
    totalCount: {
      totalCount: 0,
      perEntityTotalCount: {},
    },
  }

  while ((page - 1) * limit < totalCount) {
    // Get the next page of bookmarks.
    // eslint-disable-next-line no-await-in-loop
    const response = await tts.Users.getBookmarks(userId, { page, limit, order: '-created_at' })
    response.bookmarks.forEach(element => {
      const entityIds = element.entity_ids
      const entity = Object.keys(entityIds)[0].replace('_ids', '')
      result = addBookmarkToResult(result, entity, element)
    })

    result = {
      bookmarks: [...result.bookmarks, ...response.bookmarks],
      perEntityBookmarks: { ...result.perEntityBookmarks },
      totalCount: {
        ...result.totalCount,
        totalCount: response.totalCount,
      },
    }

    totalCount = response.totalCount
    page += 1
  }

  return result
}

const getBookmarksPerEntityThroughPagination = async (
  userId,
  requestedEntity,
  start,
  end,
  limit,
) => {
  let page = 1
  let totalCount = limit
  const result = {
    perEntityBookmarks: {
      [requestedEntity]: [],
    },
    perEntityTotalCount: {
      [requestedEntity]: 0,
    },
  }

  while (result.perEntityBookmarks[requestedEntity].length < totalCount) {
    // eslint-disable-next-line no-await-in-loop
    const response = await tts.Users.getBookmarks(userId, { page, limit, order: '-created_at' })
    if (totalCount > response.totalCount) {
      totalCount = response.totalCount
    }

    response.bookmarks.forEach(element => {
      const entityIds = element.entity_ids
      const entity = Object.keys(entityIds)[0].replace('_ids', '')
      if (entity === requestedEntity) {
        result = addBookmarkToResult(result, requestedEntity, element)
      }
    })

    result.perEntityBookmarks[requestedEntity] = result.perEntityBookmarks[requestedEntity].slice(
      start,
      end,
    )

    page += 1
  }

  return result
}

const getBookmarksListLogic = createRequestLogic({
  type: userPreferences.GET_BOOKMARKS_LIST,
  process: async ({ action }) => {
    const {
      meta: { selectors },
    } = action
    const {
      id: userId,
      params: { page, limit, deleted },
    } = action.payload

    if ('entity' in selectors && selectors.entity !== 'all') {
      const data = await getBookmarksPerEntityThroughPagination(
        userId,
        selectors.entity,
        selectors.start,
        selectors.end,
        limit,
      )

      return {
        perEntityBookmarks: data.perEntityBookmarks,
        perEntityTotalCount: data.perEntityTotalCount,
        page,
        limit,
        entity: selectors.entity,
      }
    }

    const data = await tts.Users.getBookmarks(userId, {
      page,
      limit,
      order: '-created_at',
      deleted,
    })
    return {
      entities: data.bookmarks,
      totalCount: data.totalCount,
      page,
      limit,
    }
  },
})

const getAllBookmarksLogic = createRequestLogic({
  type: userPreferences.GET_ALL_BOOKMARKS,
  process: async ({ action }) => {
    const { id: userId } = action.payload
    const data = await getBookmarksThroughPagination(userId)
    return {
      bookmarks: data.bookmarks,
      perEntityBookmarks: data.perEntityBookmarks,
      totalCount: data.totalCount,
    }
  },
})

const addBookmarkLogic = createRequestLogic({
  type: userPreferences.ADD_BOOKMARK,
  process: async ({ action }) => {
    const {
      payload: { userId, entity },
    } = action

    return await tts.Users.addBookmark(userId, entity)
  },
})

const deleteBookmarkLogic = createRequestLogic({
  type: userPreferences.DELETE_BOOKMARK,
  process: async ({ action }) => {
    const {
      payload: { userId, entity },
    } = action

    await tts.Users.deleteBookmark(userId, entity)

    return entity
  },
})

export default [getBookmarksListLogic, getAllBookmarksLogic, addBookmarkLogic, deleteBookmarkLogic]
