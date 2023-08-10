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

import tts from '@console/api/tts'

import toast from '@ttn-lw/components/toast'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'

import * as notifications from '@console/store/actions/notifications'

const getNotificationsLogic = createRequestLogic({
  type: notifications.GET_NOTIFICATIONS,
  process: async ({ action }) => {
    const {
      payload: { parentType, parentId, params },
    } = action
    const { page, limit } = params
    clearInterval()
    const result = await tts.Notifications.getAllNotifications(parentType, parentId, page, limit)
    let totalCount
    setInterval(async () => {
      const newResult = await tts.Notifications.getAllNotifications(parentType, parentId, 1, limit)
      if (newResult.totalCount > totalCount) {
        toast({ message: 'You have new notifications', type: toast.types.INFO })
      }
      totalCount = newResult.totalCount
    }, [300000]) // 5 minutes

    return { notifications: result.notifications, totalCount: result.totalCount }
  },
})

const updateNotificationStatusLogic = createRequestLogic({
  type: notifications.UPDATE_NOTIFICATION_STATUS,
  process: async ({ action }) => {
    const {
      payload: { id, notificationIds, newStatus },
    } = action

    return await tts.Notifications.updateNotificationStatus(id, notificationIds, newStatus)
  },
})

export default [getNotificationsLogic, updateNotificationStatusLogic]
