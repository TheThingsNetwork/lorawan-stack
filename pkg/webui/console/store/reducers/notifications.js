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

import {
  GET_NOTIFICATIONS_SUCCESS,
  GET_UNSEEN_NOTIFICATIONS_SUCCESS,
} from '@console/store/actions/notifications'

const defaultState = {
  notifications: {},
  unseenTotalCount: undefined,
  unseenIds: [],
  totalCount: undefined,
}

const notifications = (state = defaultState, { type, payload }) => {
  switch (type) {
    case GET_NOTIFICATIONS_SUCCESS:
      return {
        ...state,
        notifications: {
          ...payload.notifications.reduce((acc, not) => {
            acc[not.id] = not
            return acc
          }, {}),
        },
        totalCount: payload.totalCount,
      }
    case GET_UNSEEN_NOTIFICATIONS_SUCCESS:
      return {
        ...state,
        unseenTotalCount: payload.totalCount,
        unseenIds: payload.notifications.map(not => not.id),
      }
    default:
      return state
  }
}

export default notifications
