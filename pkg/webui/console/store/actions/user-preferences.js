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

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'
import {
  createPaginationBaseActionType,
  createPaginationByIdRequestActions,
} from '@ttn-lw/lib/store/actions/pagination'

export const GET_BOOKMARKS_LIST_BASE = createPaginationBaseActionType('BOOKMARKS')
export const [
  {
    request: GET_BOOKMARKS_LIST,
    success: GET_BOOKMARKS_LIST_SUCCESS,
    failure: GET_BOOKMARKS_LIST_FAILURE,
  },
  { request: getBookmarksList, success: getBookmarksListSuccess, failure: getBookmarksListFailure },
] = createPaginationByIdRequestActions('BOOKMARKS')

export const ADD_BOOKMARK_BASE = 'ADD_BOOKMARK'
export const [
  { request: ADD_BOOKMARK, success: ADD_BOOKMARK_SUCCESS, failure: ADD_BOOKMARK_FAILURE },
  { request: addBookmark, success: addBookmarkSuccess, failure: addBookmarkFailure },
] = createRequestActions(ADD_BOOKMARK_BASE, (userId, entity) => ({ userId, entity }))

export const DELETE_BOOKMARK_BASE = 'DELETE_BOOKMARK'
export const [
  { request: DELETE_BOOKMARK, success: DELETE_BOOKMARK_SUCCESS, failure: DELETE_BOOKMARK_FAILURE },
  { request: deleteBookmark, success: deleteBookmarkSuccess, failure: deleteBookmarkFailure },
] = createRequestActions(DELETE_BOOKMARK_BASE, (userId, entity) => ({ userId, entity }))
