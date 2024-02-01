// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

export const selectNotificationsStore = state => state.notifications
export const selectInboxNotifications = state =>
  selectNotificationsStore(state).notifications.inbox.entities
export const selectArchivedNotifications = state =>
  selectNotificationsStore(state).notifications.archived.entities

export const selectDropdownNotifications = createSelector([selectNotificationsStore], store =>
  Object.values(store.dropdownNotifications),
)
export const selectInboxNotificationsTotalCount = state =>
  selectNotificationsStore(state).notifications.inbox.totalCount
export const selectArchivedNotificationsTotalCount = state =>
  selectNotificationsStore(state).notifications.archived.totalCount
export const selectTotalUnseenCount = state => selectNotificationsStore(state).unseenTotalCount
