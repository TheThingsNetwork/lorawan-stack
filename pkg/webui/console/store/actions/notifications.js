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

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'

export const SHARED_NAME = 'NOTIFICATIONS'

export const GET_INBOX_NOTIFICATIONS_BASE = 'GET_INBOX_NOTIFICATIONS'
export const [
  {
    request: GET_INBOX_NOTIFICATIONS,
    success: GET_INBOX_NOTIFICATIONS_SUCCESS,
    failure: GET_INBOX_NOTIFICATIONS_FAILURE,
  },
  {
    request: getInboxNotifications,
    success: getInboxNotificationsSuccess,
    failure: getInboxNotificationsFailure,
  },
] = createRequestActions(GET_INBOX_NOTIFICATIONS_BASE, ({ page, limit } = {}) => ({
  page,
  limit,
}))

export const GET_ARCHIVED_NOTIFICATIONS_BASE = 'GET_ARCHIVED_NOTIFICATIONS'
export const [
  {
    request: GET_ARCHIVED_NOTIFICATIONS,
    success: GET_ARCHIVED_NOTIFICATIONS_SUCCESS,
    failure: GET_ARCHIVED_NOTIFICATIONS_FAILURE,
  },
  {
    request: getArchivedNotifications,
    success: getArchivedNotificationsSuccess,
    failure: getArchivedNotificationsFailure,
  },
] = createRequestActions(GET_ARCHIVED_NOTIFICATIONS_BASE, ({ page, limit } = {}) => ({
  page,
  limit,
}))

export const GET_UNSEEN_NOTIFICATIONS_PERIODICALLY_BASE = 'GET_UNSEEN_NOTIFICATIONS_PERIODICALLY'
export const [
  {
    request: GET_UNSEEN_NOTIFICATIONS_PERIODICALLY,
    success: GET_UNSEEN_NOTIFICATIONS_PERIODICALLY_SUCCESS,
    failure: GET_UNSEEN_NOTIFICATIONS_PERIODICALLY_FAILURE,
  },
  {
    request: getUnseenNotificationsPeriodically,
    success: getUnseenNotificationsPeriodicallySuccess,
    failure: getUnseenNotificationsPeriodicallyFailure,
  },
] = createRequestActions(GET_UNSEEN_NOTIFICATIONS_PERIODICALLY_BASE)

export const UPDATE_NOTIFICATION_STATUS_BASE = 'UPDATE_NOTIFICATION_STATUS'
export const [
  {
    request: UPDATE_NOTIFICATION_STATUS,
    success: UPDATE_NOTIFICATION_STATUS_SUCCESS,
    failure: UPDATE_NOTIFICATION_STATUS_FAILURE,
  },
  {
    request: updateNotificationStatus,
    success: updateNotificationStatusSuccess,
    failure: updateNotificationStatusFailure,
  },
] = createRequestActions(UPDATE_NOTIFICATION_STATUS_BASE, (notificationIds, newStatus) => ({
  notificationIds,
  newStatus,
}))

export const MARK_ALL_AS_SEEN_BASE = 'MARK_ALL_AS_SEEN'
export const [
  {
    request: MARK_ALL_AS_SEEN,
    success: MARK_ALL_AS_SEEN_SUCCESS,
    failure: MARK_ALL_AS_SEEN_FAILURE,
  },
  { request: markAllAsSeen, success: markAllAsSeenSuccess, failure: markAllAsSeenFailure },
] = createRequestActions(MARK_ALL_AS_SEEN_BASE)

export const REFRESH_NOTIFICATIONS_BASE = 'REFRESH_NOTIFICATIONS'
export const [
  {
    request: REFRESH_NOTIFICATIONS,
    success: REFRESH_NOTIFICATIONS_SUCCESS,
    failure: REFRESH_NOTIFICATIONS_FAILURE,
  },
  {
    request: refreshNotifications,
    success: refreshNotificationsSuccess,
    failure: refreshNotificationsFailure,
  },
] = createRequestActions(REFRESH_NOTIFICATIONS_BASE)
