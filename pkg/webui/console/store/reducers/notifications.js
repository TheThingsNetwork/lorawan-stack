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

import { GET_NOTIFICATIONS_SUCCESS } from '@console/store/actions/notifications'

const defaultState = {
  notifications: {
    seen: {},
    unseen: {},
    archived: {},
  },
}

const getUnseenNotifications = notifications =>
  notifications.reduce((acc, not) => {
    if (!('status' in not) || not.status === 'NOTIFICATION_STATUS_UNSEEN') {
      acc[not.id] = not
    }

    return acc
  }, {})

const getNotifications = (notifications, status) =>
  notifications.reduce((acc, not) => {
    if (not.status === status) {
      acc[not.id] = not
    }

    return acc
  }, {})

const notifications = (state = defaultState, { type, payload }) => {
  switch (type) {
    case GET_NOTIFICATIONS_SUCCESS:
      return {
        ...state.notifications,
        seen: {
          ...state?.notifications?.seen,
          ...getNotifications(payload.notifications, 'NOTIFICATION_STATUS_SEEN'),
        },
        unseen: {
          ...state?.notifications?.unseen,
          ...getUnseenNotifications(payload.notifications),
        },
        archived: {
          ...state?.notifications?.archived,
          ...getNotifications(payload.notifications, 'NOTIFICATION_STATUS_ARCHIVED'),
        },
      }
    default:
      return state
  }
}

export default notifications
