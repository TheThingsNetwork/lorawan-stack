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

const selectUserPreferencesStore = state => state.userPreferences

export const selectConsolePreferences = state =>
  selectUserPreferencesStore(state).consolePreferences

export const selectBookmarksList = state => selectUserPreferencesStore(state).bookmarks.bookmarks

export const selectApplicationBookmarks = createSelector([selectBookmarksList], bookmarks =>
  bookmarks.filter(bookmark =>
    bookmark.entity_ids
      ? Object.keys(bookmark.entity_ids)[0].replace('_ids', '') === 'application'
      : [],
  ),
)

export const selectGatewayBookmarks = createSelector([selectBookmarksList], bookmarks =>
  bookmarks.filter(bookmark =>
    bookmark.entity_ids
      ? Object.keys(bookmark.entity_ids)[0].replace('_ids', '') === 'gateway'
      : [],
  ),
)

export const selectEndDeviceBookmarks = createSelector([selectBookmarksList], bookmarks =>
  bookmarks.filter(bookmark =>
    bookmark.entity_ids ? Object.keys(bookmark.entity_ids)[0].replace('_ids', '') === 'device' : [],
  ),
)

export const selectBookmarksTotalCount = state =>
  selectUserPreferencesStore(state).bookmarks.totalCount.totalCount

export const selectPerEntityTotalCount = (state, entity) =>
  selectUserPreferencesStore(state).bookmarks.totalCount.perEntityTotalCount[entity] || 0
