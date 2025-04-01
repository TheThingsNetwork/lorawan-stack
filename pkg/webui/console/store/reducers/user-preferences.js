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

import { fillIntoArray, pageToIndices } from '@console/store/utils'

import {
  ADD_BOOKMARK_SUCCESS,
  DELETE_BOOKMARK_SUCCESS,
  GET_ALL_BOOKMARKS_SUCCESS,
  GET_BOOKMARKS_LIST_SUCCESS,
} from '@console/store/actions/user-preferences'
import { APPLY_PERSISTED_STATE_SUCCESS, GET_USER_ME_SUCCESS } from '@console/store/actions/user'

import { UPDATE_USER_SUCCESS } from '../actions/users'

const initialState = {
  bookmarks: {
    bookmarks: [],
    totalCount: {},
    perEntityBookmarks: {},
  },
  consolePreferences: {
    console_theme: 'CONSOLE_THEME_SYSTEM',
  },
}

const userPreferences = (state = initialState, { type, payload }) => {
  switch (type) {
    case GET_ALL_BOOKMARKS_SUCCESS:
      return {
        ...state,
        bookmarks: {
          ...state.bookmarks,
          bookmarks: payload.bookmarks,
          perEntityBookmarks: payload.perEntityBookmarks,
          totalCount: payload.totalCount,
        },
      }
    case GET_BOOKMARKS_LIST_SUCCESS:
      if ('perEntityBookmarks' in payload) {
        return {
          ...state,
          bookmarks: {
            ...state.bookmarks,
            perEntityBookmarks: {
              ...state.bookmarks.perEntityBookmarks,
              [payload.entity]: fillIntoArray(
                state.bookmarks.perEntityBookmarks[payload.entity],
                pageToIndices(payload.page, payload.limit)[0],
                payload.perEntityBookmarks[payload.entity],
                payload.perEntityTotalCount[payload.entity],
              ),
            },
            totalCount: {
              ...state.bookmarks.totalCount,
              perEntityTotalCount: payload.perEntityTotalCount,
            },
          },
        }
      }

      return {
        ...state,
        bookmarks: {
          ...state.bookmarks,
          bookmarks: fillIntoArray(
            state.bookmarks.bookmarks,
            pageToIndices(payload.page, payload.limit)[0],
            payload.entities,
            payload.totalCount,
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
          bookmarks: [...state.bookmarks.bookmarks, payload],
          totalCount: {
            ...state.bookmarks.totalCount,
            totalCount: state.bookmarks.totalCount.totalCount + 1,
          },
        },
      }
    case DELETE_BOOKMARK_SUCCESS:
      return {
        ...state,
        bookmarks: {
          ...state.bookmarks,
          bookmarks: [...state.bookmarks.bookmarks].filter(
            b => b.entity_ids?.[`${payload.name}_ids`]?.[`${payload.name}_id`] !== payload.id,
          ),
          totalCount: {
            ...state.bookmarks.totalCount,
            totalCount: state.bookmarks.totalCount.totalCount - 1,
          },
        },
      }
    case GET_USER_ME_SUCCESS:
    case UPDATE_USER_SUCCESS:
      return {
        ...state,
        consolePreferences: {
          ...state.consolePreferences,
          ...payload.console_preferences,
        },
      }
    case 'SET_PAGE_SIZE':
      return {
        ...state,
        consolePreferences: {
          ...state.consolePreferences,
          pageSize: payload,
        },
      }
    case APPLY_PERSISTED_STATE_SUCCESS:
      return {
        ...state,
        consolePreferences: {
          ...state.consolePreferences,
          pageSize: payload.userPreferences?.consolePreferences?.pageSize,
        },
      }
    default:
      return state
  }
}

export default userPreferences
