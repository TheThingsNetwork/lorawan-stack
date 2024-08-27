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

import NOTIFICATION_STATUS from '@console/containers/notifications/notification-status'

import {
  GET_ARCHIVED_NOTIFICATIONS_SUCCESS,
  GET_INBOX_NOTIFICATIONS_SUCCESS,
  GET_UNSEEN_NOTIFICATIONS_SUCCESS,
  MARK_ALL_AS_SEEN_SUCCESS,
  REFRESH_NOTIFICATIONS_SUCCESS,
  UPDATE_NOTIFICATION_STATUS_SUCCESS,
} from '@console/store/actions/notifications'

const defaultState = {
  notifications: {
    inbox: { entities: [], totalCount: 0 },
    archived: { entities: [], totalCount: 0 },
  },
  unseenTotalCount: undefined,
}

const notifications = (state = defaultState, { type, payload }) => {
  switch (type) {
    case GET_INBOX_NOTIFICATIONS_SUCCESS:
      return {
        ...state,
        notifications: {
          ...state.notifications,
          inbox: {
            entities: fillIntoArray(
              state.notifications.inbox.entities,
              pageToIndices(payload.page, payload.limit)[0],
              payload.notifications,
              payload.totalCount,
            ),
            totalCount: payload.totalCount,
          },
        },
      }
    case REFRESH_NOTIFICATIONS_SUCCESS:
      if (
        payload.unseenTotalCount === undefined ||
        payload.unseenTotalCount === state.unseenTotalCount
      ) {
        return state
      }
      return {
        ...state,
        unseenTotalCount: payload.unseenTotalCount,
      }

    case GET_ARCHIVED_NOTIFICATIONS_SUCCESS:
      return {
        ...state,
        notifications: {
          ...state.notifications,
          archived: {
            entities: fillIntoArray(
              state.notifications.archived.entities,
              pageToIndices(payload.page, payload.limit)[0],
              payload.notifications,
              payload.totalCount,
            ),
            totalCount: payload.totalCount,
          },
        },
      }
    case GET_UNSEEN_NOTIFICATIONS_SUCCESS:
      return {
        ...state,
        unseenTotalCount: payload.totalCount,
      }
    case UPDATE_NOTIFICATION_STATUS_SUCCESS:
      return {
        ...state,
        unseenTotalCount:
          payload.status === NOTIFICATION_STATUS.SEEN && state.unseenTotalCount > 0
            ? state.unseenTotalCount - payload.ids.length
            : state.unseenTotalCount,
        notifications: {
          ...state.notifications,
          inbox: {
            ...state.notifications.inbox,
            entities: state.notifications.inbox.entities.map(entity =>
              payload.ids.includes(entity.id) ? { ...entity, status: payload.status } : entity,
            ),
          },
        },
      }
    case MARK_ALL_AS_SEEN_SUCCESS:
      return {
        ...state,
        unseenTotalCount: 0,
        notifications: {
          ...state.notifications,
          inbox: {
            ...state.notifications.inbox,
            entities: state.notifications.inbox.entities.map(entity => ({
              ...entity,
              status: NOTIFICATION_STATUS.SEEN,
            })),
          },
        },
      }
    default:
      return state
  }
}

export default notifications
