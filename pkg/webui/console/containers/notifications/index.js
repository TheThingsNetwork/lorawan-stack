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

import React, { useCallback, useState, useRef, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'
import classNames from 'classnames'
import { defineMessages } from 'react-intl'

import LAYOUT from '@ttn-lw/constants/layout'

import Button from '@ttn-lw/components/button'
import Spinner from '@ttn-lw/components/spinner'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import {
  getArchivedNotifications,
  getInboxNotifications,
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
})

const Notifications = React.memo(() => {
  const listRef = useRef(null)
  const dispatch = useDispatch()
  const navigate = useNavigate()
  const { id: notificationId, category = 'inbox' } = useParams()
  const showArchived = category === 'archived'
  const items = useSelector(showArchived ? selectArchivedNotifications : selectInboxNotifications)
  const totalCount = useSelector(
    showArchived ? selectArchivedNotificationsTotalCount : selectInboxNotificationsTotalCount,
  )
  const totalUnseenCount = useSelector(selectTotalUnseenCount)
  const hasNextPage = items.length < totalCount
  const [fetching, setFetching] = useState(false)
  const [selectionCache, setSelectionCache] = useState({ archived: undefined, inbox: undefined })
  const [isSmallScreen, setIsSmallScreen] = useState(window.innerWidth < LAYOUT.BREAKPOINTS.M)

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
      const updateFilter = showArchived
        ? 'NOTIFICATION_STATUS_SEEN'
        : 'NOTIFICATION_STATUS_ARCHIVED'

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
        navigate(`/notifications/${category}/${previousNotification.id}`)
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

  // Update the selection cache when a notification is selected.
  // This is used to restore the selection when switching between categories.
  useEffect(() => {
    setSelectionCache(v => ({ ...v, [category]: notificationId }))
  }, [category, notificationId])

  // Set the first notification as selected if none is currently selected.
  useEffect(() => {
    if (
      !isSmallScreen &&
      (!notificationId || !items.find(i => i.id === notificationId)) &&
      items.length > 0
    ) {
      const selectedId = selectionCache[category] || items[0].id
      navigate(`/notifications/${category}/${selectedId}`)
    }
  }, [category, isSmallScreen, items, navigate, notificationId, selectionCache])

  // Update the status of the selected notification to seen.
  useEffect(() => {
    const clickedNotification = items.find(notification => notification.id === notificationId)
    const index = items.findIndex(notification => notification.id === notificationId)
    const timer = setTimeout(() => {
      if (category === 'inbox' && !clickedNotification?.status && totalUnseenCount > 0) {
        dispatch(
          attachPromise(
            updateNotificationStatus([clickedNotification.id], 'NOTIFICATION_STATUS_SEEN'),
          ),
        )
        loadNextPage(index, index + 1)
      }
    }, 500)

    return () => clearTimeout(timer)
  }, [category, items, dispatch, loadNextPage, totalUnseenCount, notificationId])

  const selectedNotification = items?.find(item => item.id === notificationId)

  if (!items) {
    return (
      <div className="d-flex flex-grow">
        <Spinner center />
      </div>
    )
  }

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
        {selectedNotification && (
          <NotificationContent
            selectedNotification={selectedNotification}
            onArchive={handleArchive}
            isSmallScreen={isSmallScreen}
          />
        )}
      </div>
    </div>
  )
})

export default Notifications
