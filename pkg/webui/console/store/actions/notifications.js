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
import {
  createPaginationBaseActionType,
  createPaginationByParentRequestActions,
} from '@ttn-lw/lib/store/actions/pagination'

export const SHARED_NAME = 'NOTIFICATIONS'

export const GET_NOTIFICATIONS_BASE = createPaginationBaseActionType(SHARED_NAME)
export const [
  {
    request: GET_NOTIFICATIONS,
    success: GET_NOTIFICATIONS_SUCCESS,
    failure: GET_NOTIFICATIONS_FAILURE,
  },
  { request: getNotifications, success: getNotificationsSuccess, failure: getNotificationsFailure },
] = createPaginationByParentRequestActions(SHARED_NAME)

export const GET_DROPDOWN_NOTIFICATIONS_BASE = 'GET_DROPDOWN_NOTIFICATIONS'
export const [
  {
    request: GET_DROPDOWN_NOTIFICATIONS,
    success: GET_DROPDOWN_NOTIFICATIONS_SUCCESS,
    failure: GET_DROPDOWN_NOTIFICATIONS_FAILURE,
  },
  {
    request: getDropdownNotifications,
    success: getDropdownNotificationsSuccess,
    failure: getDropdownNotificationsFailure,
  },
] = createRequestActions(GET_DROPDOWN_NOTIFICATIONS_BASE, userId => ({
  userId,
}))

export const GET_UNSEEN_NOTIFICATIONS_BASE = 'GET_UNSEEN_NOTIFICATIONS'
export const [
  {
    request: GET_UNSEEN_NOTIFICATIONS,
    success: GET_UNSEEN_NOTIFICATIONS_SUCCESS,
    failure: GET_UNSEEN_NOTIFICATIONS_FAILURE,
  },
  {
    request: getUnseenNotifications,
    success: getUnseenNotificationsSuccess,
    failure: getUnseenNotificationsFailure,
  },
] = createRequestActions(GET_UNSEEN_NOTIFICATIONS_BASE, id => ({
  id,
}))

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
] = createRequestActions(UPDATE_NOTIFICATION_STATUS_BASE, (id, notificationIds, newStatus) => ({
  id,
  notificationIds,
  newStatus,
}))
