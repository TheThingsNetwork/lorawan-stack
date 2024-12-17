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

import { defineMessage } from 'react-intl'
import { createLogic } from 'redux-logic'

import tts from '@console/api/tts'

import toast from '@ttn-lw/components/toast'

import { validateNotification } from '@console/components/notifications/utils'

import NOTIFICATION_STATUS from '@console/containers/notifications/notification-status'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { selectIsOnlineStatus } from '@ttn-lw/lib/store/selectors/status'

import * as notifications from '@console/store/actions/notifications'

import { selectUserId } from '@console/store/selectors/user'
import { selectTotalUnseenCount } from '@console/store/selectors/notifications'

const m = defineMessage({
  newNotifications: 'You have new notifications',
})

const updateThroughPagination = async (totalCount, userId) => {
  let page = 1
  const limit = 1000
  let result = []

  while ((page - 1) * limit < totalCount) {
    // Get the next page of notifications.
    // eslint-disable-next-line no-await-in-loop
    const notifications = await tts.Notifications.getAllNotifications(
      userId,
      [NOTIFICATION_STATUS.UNSEEN],
      page,
      limit,
    )
    // Get the notification ids.
    const notificationIds = notifications.notifications.map(notification => notification.id)
    // Make the update request.
    // eslint-disable-next-line no-await-in-loop
    await tts.Notifications.updateNotificationStatus(
      userId,
      notificationIds,
      NOTIFICATION_STATUS.SEEN,
    )

    result = [...result, ...notificationIds]
    page += 1
  }

  return result
}

const getInboxNotificationsLogic = createRequestLogic({
  type: notifications.GET_INBOX_NOTIFICATIONS,
  process: async ({ action, getState }) => {
    const {
      payload: { page = 1, limit = 1000 },
    } = action
    const filter = [NOTIFICATION_STATUS.UNSEEN, NOTIFICATION_STATUS.SEEN]
    const userId = selectUserId(getState())
    const result = await tts.Notifications.getAllNotifications(userId, filter, page, limit)
    // Remove unknown notification types from the result and send error to Sentry.
    // When a new notification type is introduced in the backend, this will prevent
    // the Console from crashing.
    const validatedResult = result.notifications.filter(notification =>
      validateNotification(notification.notification_type, 'getInboxNotificationsLogic'),
    )

    return {
      notifications: validatedResult,
      // Subtract the number of unknown notification types from the total count.
      totalCount: result.totalCount - (result.notifications.length - validatedResult.length),
      page,
      limit,
    }
  },
})

const refreshNotificationsLogic = createRequestLogic({
  type: notifications.REFRESH_NOTIFICATIONS,
  validate: ({ getState }, allow, reject) => {
    // Avoid refreshing notifications while the Console is offline.
    const isOnline = selectIsOnlineStatus(getState())
    if (isOnline) {
      allow()
    } else {
      reject()
    }
  },
  process: async ({ getState }, dispatch) => {
    const state = getState()
    const userId = selectUserId(state)
    const prevTotalUnseenCount = selectTotalUnseenCount(state)

    const unseen = await tts.Notifications.getAllNotifications(
      userId,
      [NOTIFICATION_STATUS.UNSEEN],
      1,
      1,
    )
    // Remove unknown notification types from the result and send error to Sentry.
    // When a new notification type is introduced in the backend, this will prevent
    // the Console from crashing.
    const validatedUnseen = unseen.notifications.filter(notification =>
      validateNotification(notification.notification_type, 'refreshNotificationsLogic'),
    )

    // Subtract the number of unknown notification types from the total count.
    const validatedUnseenCount =
      unseen.totalCount - (unseen.notifications.length - validatedUnseen.length)
    // If there are new unseen notifications, show a toast and fetch the notifications.
    if (unseen && validatedUnseenCount > prevTotalUnseenCount) {
      toast({
        title: m.newNotifications,
        type: toast.types.INFO,
      })
      await dispatch(attachPromise(notifications.getInboxNotifications()))
    }

    return { unseenTotalCount: validatedUnseenCount }
  },
})

const getArchivedNotificationsLogic = createRequestLogic({
  type: notifications.GET_ARCHIVED_NOTIFICATIONS,
  process: async ({ action, getState }) => {
    const {
      payload: { page, limit },
    } = action
    const filter = [NOTIFICATION_STATUS.ARCHIVED]
    const userId = selectUserId(getState())
    const result = await tts.Notifications.getAllNotifications(userId, filter, page, limit)
    // Remove unknown notification types from the result and send error to Sentry.
    // When a new notification type is introduced in the backend, this will prevent
    // the Console from crashing.
    const validatedResult = result.notifications.filter(notification =>
      validateNotification(notification.notification_type, 'getArchivedNotificationsLogic'),
    )

    return {
      notifications: validatedResult,
      // Subtract the number of unknown notification types from the total count.
      totalCount: result.totalCount - (result.notifications.length - validatedResult.length),
      page,
      limit,
    }
  },
})

const getUnseenNotificationsLogic = createRequestLogic({
  type: notifications.GET_UNSEEN_NOTIFICATIONS,
  process: async ({ action, getState }) => {
    const {
      payload: { page, limit },
    } = action
    const filter = [NOTIFICATION_STATUS.UNSEEN]
    const userId = selectUserId(getState())
    const result = await tts.Notifications.getAllNotifications(userId, filter, page, limit)
    // Remove unknown notification types from the result and send error to Sentry.
    // When a new notification type is introduced in the backend, this will prevent
    // the Console from crashing.
    const validatedResult = result.notifications.filter(notification =>
      validateNotification(notification.notification_type, 'getUnseenNotificationsLogic'),
    )

    return {
      notifications: validatedResult,
      // Subtract the number of unknown notification types from the total count.
      totalCount: result.totalCount - (result.notifications.length - validatedResult.length),
      page,
      limit,
    }
  },
})

const getUnseenNotificationsPeriodicallyLogic = createLogic({
  type: notifications.GET_UNSEEN_NOTIFICATIONS_PERIODICALLY,
  processOptions: {
    dispatchMultiple: true,
  },
  warnTimeout: 0,
  process: async (_, dispatch) => {
    // Fetch once initially.
    await dispatch(notifications.refreshNotifications())

    setInterval(
      async () => {
        dispatch(notifications.refreshNotifications())
      },
      // Refresh notifications every 15 minutes.
      1000 * 60 * 15,
    )
  },
})

const updateNotificationStatusLogic = createRequestLogic({
  type: notifications.UPDATE_NOTIFICATION_STATUS,
  process: async ({ action, getState }) => {
    const {
      payload: { notificationIds, newStatus },
    } = action
    const id = selectUserId(getState())
    await tts.Notifications.updateNotificationStatus(id, notificationIds, newStatus)

    return { ids: notificationIds, status: newStatus }
  },
})

const markAllAsSeenLogic = createRequestLogic({
  type: notifications.MARK_ALL_AS_SEEN,
  process: async ({ getState }) => {
    const id = selectUserId(getState())
    const totalUnseenCount = selectTotalUnseenCount(getState())

    return updateThroughPagination(totalUnseenCount, id)
  },
})

export default [
  getInboxNotificationsLogic,
  refreshNotificationsLogic,
  getArchivedNotificationsLogic,
  getUnseenNotificationsLogic,
  getUnseenNotificationsPeriodicallyLogic,
  updateNotificationStatusLogic,
  markAllAsSeenLogic,
]
