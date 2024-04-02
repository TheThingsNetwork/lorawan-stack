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

const getPerEntityTotalCountThroughPagination = async (totalCount, userId) => {
  let page = 1
  const limit = 1000
  const result = {}

  while ((page - 1) * limit < totalCount) {
    // Get the next page of notifications.
    // eslint-disable-next-line no-await-in-loop
    const response = await tts.Users.getBookmarks(userId, { page, limit })
    response.bookmarks.forEach(element => {
      const entityIds = element.entity_ids
      const entity = Object.keys(entityIds)[0].replace('_ids', '')
      if (!result[entity]) {
        result[entity] = 1
      } else {
        result[entity] += 1
      }
    })

    page += 1
  }

  return result
}

const getBookmarksListLogic = createRequestLogic({
  type: userPreferences.GET_BOOKMARKS_LIST,
  process: async ({ action }) => {
    const {
      id: userId,
      params: { page, limit, order, deleted },
    } = action.payload
    const data = await tts.Users.getBookmarks(userId, { page, limit, order, deleted })
    const perEntityTotalCount = await getPerEntityTotalCountThroughPagination(
      data.totalCount,
      userId,
    )
    return {
      entities: data.bookmarks,
      totalCount: { totalCount: data.totalCount, perEntityTotalCount },
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

    return await tts.Users.deleteBookmark(userId, entity)
  },
})

export default [getBookmarksListLogic, addBookmarkLogic, deleteBookmarkLogic]
