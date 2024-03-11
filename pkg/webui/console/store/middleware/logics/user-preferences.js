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

import tts from '@console/api/tts'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as userPreferences from '@console/store/actions/user-preferences'

const getBookmarksListLogic = createRequestLogic({
  type: userPreferences.GET_BOOKMARKS_LIST,
  process: async ({ action }) => {
    const {
      id: userId,
      params: { page, limit, order, deleted },
    } = action.payload
    const data = await tts.Users.getBookmarks(userId, { page, limit, order, deleted })

    return { entities: data.bookmarks, totalCount: data.totalCount }
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
