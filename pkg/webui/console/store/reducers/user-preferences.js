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

import {
  ADD_BOOKMARK_SUCCESS,
  GET_ALL_BOOKMARKS_SUCCESS,
  GET_BOOKMARKS_LIST_SUCCESS,
} from '@console/store/actions/user-preferences'
import { GET_USER_ME_SUCCESS } from '@console/store/actions/logout'

// Update a range of values in an array by using another array and a start index.
const fillIntoArray = (array, start, values, totalCount) => {
  const newArray = [...array]
  const end = Math.min(start + values.length, totalCount)
  for (let i = start; i < end; i++) {
    newArray[i] = values[i - start]
  }
  return newArray
}

const pageToIndices = (page, limit) => {
  const startIndex = (page - 1) * limit
  const stopIndex = page * limit - 1
  return [startIndex, stopIndex]
}

const initialState = {
  bookmarks: {
    bookmarks: [],
    totalCount: {},
  },
  consolePreferences: {},
}

const userPreferences = (state = initialState, { type, payload }) => {
  switch (type) {
    case GET_ALL_BOOKMARKS_SUCCESS:
      return {
        ...state,
        bookmarks: {
          ...state.bookmarks,
          bookmarks: payload.entities,
          totalCount: payload.totalCount,
        },
      }
    case GET_BOOKMARKS_LIST_SUCCESS:
      return {
        ...state,
        bookmarks: {
          ...state.bookmarks,
          bookmarks: fillIntoArray(
            state.bookmarks.bookmarks,
            pageToIndices(payload.page, payload.limit)[0],
            payload.entities,
            payload.totalCount.totalCount,
          ),
          totalCount: {
            ...state.bookmarks.totalCount,
            totalCount: payload.totalCount,
          },
        },
      }
    case ADD_BOOKMARK_SUCCESS:
      return {
        ...state,
        bookmarks: {
          ...state.bookmarks,
          totalCount: {
            ...state.bookmarks.totalCount,
            totalCount: state.bookmarks.totalCount.totalCount + 1,
          },
        },
      }
    case GET_USER_ME_SUCCESS:
      return {
        ...state,
        consolePreferences: {
          ...state.consolePreferences,
          ...payload.console_preferences,
        },
      }
    default:
      return state
  }
}

export default userPreferences
