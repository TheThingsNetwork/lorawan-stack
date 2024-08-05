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

import React, { useCallback, useState, useRef, useEffect, useMemo } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'
import classNames from 'classnames'
import { defineMessages } from 'react-intl'

import LAYOUT from '@ttn-lw/constants/layout'

import Button from '@ttn-lw/components/button'
import Spinner from '@ttn-lw/components/spinner'
import Icon, { IconInbox } from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import NOTIFICATION_STATUS from '@console/containers/notifications/notification-status'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import useRequest from '@ttn-lw/lib/hooks/use-request'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  getArchivedNotifications,
  getInboxNotifications,
  getUnseenNotifications,
  refreshNotifications,
  updateNotificationStatus,
} from '@console/store/actions/notifications'

import {
  selectArchivedNotifications,
  selectArchivedNotificationsTotalCount,
  selectInboxNotifications,
  selectInboxNotificationsTotalCount,
  selectTotalUnseenCount,
} from '@console/store/selectors/notifications'

import NotificationList from './notification-list'
import NotificationContent from './notification-content'

import style from './notifications.styl'

const BATCH_SIZE = 50

const indicesToPage = (startIndex, stopIndex, limit) => {
  const startPage = Math.floor(startIndex / limit) + 1
  const stopPage = Math.floor(stopIndex / limit) + 1
  return [startPage, stopPage]
}

const m = defineMessages({
  seeArchived: 'See archived messages',
  seeAll: 'See all messages',
  noNotificationsSubtitle: 'Once you receive a notification, it can be viewed here',
  unreadMessagesTitle: '{value} unread notifications',
  archivedMessageTitle: '{value} archived notifications',
  unreadMessagesSubtitle: 'Select a notification to display the content here.',
  noArchivedNotificationsSubtitle: 'Once you archive a notification, it can be viewed here',
})

const Notifications = React.memo(() => {
  const listRef = useRef(null)
  const dispatch = useDispatch()
  const navigate = useNavigate()
  const { id: notificationId, category } = useParams()
  const showArchived = category === 'archived'
  const items = useSelector(showArchived ? selectArchivedNotifications : selectInboxNotifications)
  const totalCount = useSelector(
    showArchived ? selectArchivedNotificationsTotalCount : selectInboxNotificationsTotalCount,
  )
  const totalUnseenCount = useSelector(selectTotalUnseenCount)
  const hasNextPage = items.length < totalCount
  const [fetching, setFetching] = useState(false)
  const [updatePendingNotificationIds, setUpdatePendingNotificationIds] = useState([])
  const [updateNotificationStatusLoading, setUpdateNotificationStatusLoading] = useState(false)
  const [isSmallScreen, setIsSmallScreen] = useState(window.innerWidth < LAYOUT.BREAKPOINTS.M)
  const [unseenNotificationsLoading] = useRequest(getUnseenNotifications({ page: 1, limit: 1 }))

  const unreadNotificationsCount = unseenNotificationsLoading
    ? items.filter(
        n =>
          ![NOTIFICATION_STATUS.SEEN, NOTIFICATION_STATUS.ARCHIVED].includes(n.status) &&
          !updatePendingNotificationIds.includes(n.id),
      ).length
    : totalUnseenCount

  const loadNextPage = useCallback(
    async (startIndex, stopIndex) => {
      if (fetching) return
      setFetching(true)

      // Determine filter based on whether archived notifications should be shown.
      const action = showArchived ? getArchivedNotifications : getInboxNotifications
      // Calculate the number of items to fetch.
      const limit = Math.max(BATCH_SIZE, stopIndex - startIndex + 1)
      const [startPage, stopPage] = indicesToPage(startIndex, stopIndex, limit)

      // Fetch new notifications with a maximum of 1000 items.
      await dispatch(
        attachPromise(
          action({
            limit: Math.min((stopPage - startPage + 1) * BATCH_SIZE, 1000),
            page: startPage,
          }),
        ),
      )

      setFetching(false)
    },
    [fetching, showArchived, dispatch],
  )

  const handleArchive = useCallback(
    async (_, id) => {
      // Determine the filter to apply based on the showArchived state.
      const updateFilter = showArchived ? NOTIFICATION_STATUS.SEEN : NOTIFICATION_STATUS.ARCHIVED

      // Update the status of the notification.
      await dispatch(attachPromise(updateNotificationStatus([id], updateFilter)))

      // Find the index of the archived notification.
      const index = items.findIndex(item => item.id === id)

      // Update the selected notification to the one above the archived one,
      // unless there is only one item in the list.
      const previousNotification = totalCount === 1 ? undefined : items[Math.max(0, index - 1)]

      // Reload notifications starting from the archived one.
      await loadNextPage(
        index,
        index + BATCH_SIZE > items.length - 1 ? items.length - 1 : index + BATCH_SIZE,
      )

      if (isSmallScreen) {
        navigate(`/notifications/${category}`)
      } else {
        navigate(
          `/notifications/${category}${previousNotification ? `/${previousNotification.id}` : ''}`,
        )
      }

      // Reset the list cache if available so that old items are discarded.
      if (listRef.current && listRef.current.resetloadMoreItemsCache) {
        listRef.current.resetloadMoreItemsCache()
      }
    },
    [showArchived, dispatch, items, totalCount, navigate, category, loadNextPage, isSmallScreen],
  )

  // Add a resize handler to detect mobile experiences.
  useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth < LAYOUT.BREAKPOINTS.M) {
        setIsSmallScreen(true)
      }
    }
    window.addEventListener('resize', handleResize)

    return () => window.removeEventListener('resize', handleResize)
  }, [category, dispatch, showArchived])

  const selectedNotification = useMemo(
    () => items?.find(item => item.id === notificationId),
    [items, notificationId],
  )

  const isUpdateStatusPending = useMemo(
    () => updatePendingNotificationIds.some(id => id === selectedNotification?.id),
    [selectedNotification, updatePendingNotificationIds],
  )

  const handleUpdateNotificationStatus = useCallback(async () => {
    setUpdateNotificationStatusLoading(true)
    await dispatch(
      attachPromise(
        updateNotificationStatus(updatePendingNotificationIds, NOTIFICATION_STATUS.SEEN),
      ),
    )
    setUpdatePendingNotificationIds([])
    setUpdateNotificationStatusLoading(false)
  }, [dispatch, updatePendingNotificationIds])

  useEffect(() => {
    if (
      selectedNotification?.id &&
      ![NOTIFICATION_STATUS.SEEN, NOTIFICATION_STATUS.ARCHIVED].includes(
        selectedNotification?.status,
      ) &&
      !isUpdateStatusPending
    ) {
      setUpdatePendingNotificationIds(ids => [...ids, selectedNotification.id])
    }
  }, [isUpdateStatusPending, selectedNotification])

  useEffect(() => {
    if (
      !updateNotificationStatusLoading &&
      updatePendingNotificationIds.length !== 0 &&
      updatePendingNotificationIds.length <= unreadNotificationsCount
    ) {
      handleUpdateNotificationStatus()
    }
  }, [
    handleUpdateNotificationStatus,
    unreadNotificationsCount,
    updateNotificationStatusLoading,
    updatePendingNotificationIds.length,
  ])

  if (!items) {
    return (
      <div className="d-flex flex-grow">
        <Spinner center />
      </div>
    )
  }

  const messageBox = useCallback(
    ({ icon, title, subtitle, titleValues }) => (
      <div className={classNames(style.messageBox, 'h-full d-flex al-center j-center gap-cs-s')}>
        <Icon size={64} icon={icon} className="c-icon-neutral-extralight" />
        <div className="d-flex flex-column al-center gap-cs-xxs">
          <Message content={title} className="fw-bold fs-l" values={{ ...titleValues }} />
          <Message content={subtitle} className="c-text-neutral-light" />
        </div>
      </div>
    ),
    [],
  )

  const mainContent = useMemo(() => {
    if (selectedNotification) {
      return (
        <NotificationContent
          selectedNotification={selectedNotification}
          onArchive={handleArchive}
          isSmallScreen={isSmallScreen}
        />
      )
    }
    if (!totalCount) {
      return messageBox({
        icon: IconInbox,
        title: sharedMessages.noNotifications,
        subtitle: showArchived ? m.noArchivedNotificationsSubtitle : m.noNotificationsSubtitle,
      })
    }
    return messageBox({
      icon: IconInbox,
      title: showArchived ? m.archivedMessageTitle : m.unreadMessagesTitle,
      subtitle: m.unreadMessagesSubtitle,
      titleValues: { value: showArchived ? totalCount : unreadNotificationsCount },
    })
  }, [
    handleArchive,
    isSmallScreen,
    messageBox,
    selectedNotification,
    showArchived,
    totalCount,
    unreadNotificationsCount,
  ])

  return (
    <div className="d-flex flex-grow">
      <div
        className={classNames(style.notificationList, 'flex-grow', {
          [style.notificationSelected]: selectedNotification,
        })}
      >
        <NotificationList
          hasNextPage={hasNextPage}
          loadNextPage={loadNextPage}
          items={items}
          totalCount={totalCount}
          selectedNotification={selectedNotification}
          updatePendingNotificationIds={updatePendingNotificationIds}
          unreadNotificationsCount={unreadNotificationsCount}
          listRef={listRef}
        />
        <div className="d-flex j-center">
          <Button.Link
            to={showArchived ? '/notifications/inbox' : '/notifications/archived'}
            message={showArchived ? m.seeAll : m.seeArchived}
            className={style.notificationListChangeButton}
            naked
          />
        </div>
      </div>
      <div
        className={classNames(style.notificationContent, {
          [style.notificationSelected]: selectedNotification,
        })}
      >
        {mainContent}
      </div>
    </div>
  )
})

export default Notifications
