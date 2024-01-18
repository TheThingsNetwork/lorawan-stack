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
import { useDispatch, useSelector } from 'react-redux'
import classNames from 'classnames'
import { defineMessages } from 'react-intl'

import Button from '@ttn-lw/components/button'
import Spinner from '@ttn-lw/components/spinner'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import useQueryState from '@ttn-lw/lib/hooks/use-query-state'

import { getNotifications, updateNotificationStatus } from '@console/store/actions/notifications'

import { selectUserId } from '@console/store/selectors/logout'

import NotificationList from './notification-list'
import NotificationContent from './notification-content'

import style from './notifications.styl'

const BATCH_SIZE = 50

// Update a range of values in an array by using another array and a start index.
const fillIntoArray = (array, start, values, totalCount) => {
  const newArray = [...array]
  const end = Math.min(start + values.length, totalCount)
  for (let i = start; i < end; i++) {
    newArray[i] = values[i - start]
  }
  return newArray
}

const indicesToPage = (startIndex, stopIndex, limit) => {
  const startPage = Math.floor(startIndex / limit) + 1
  const stopPage = Math.floor(stopIndex / limit) + 1
  return [startPage, stopPage]
}

const pageToIndices = (page, limit) => {
  const startIndex = (page - 1) * limit
  const stopIndex = page * limit - 1
  return [startIndex, stopIndex]
}

const m = defineMessages({
  seeArchived: 'See archived messages',
  seeAll: 'See all messages',
})

const Notifications = () => {
  const listRef = useRef(null)
  const userId = useSelector(selectUserId)
  const dispatch = useDispatch()
  const [selectedNotification, setSelectedNotification] = useState(undefined)
  const [hasNextPage, setHasNextPage] = useState(true)
  const [items, setItems] = useState(undefined)
  const [showArchived, setShowArchived] = useQueryState('archived', 'false')
  const [totalCount, setTotalCount] = useState(0)
  const [fetching, setFetching] = useState(false)

  const loadNextPage = useCallback(
    async (startIndex, stopIndex, archived) => {
      if (fetching) return
      setFetching(true)
      const composedArchived =
        archived === undefined ? showArchived === 'true' : archived === 'true'

      // Determine filters based on whether archived notifications should be shown.
      const filters = composedArchived
        ? ['NOTIFICATION_STATUS_ARCHIVED']
        : ['NOTIFICATION_STATUS_UNSEEN', 'NOTIFICATION_STATUS_SEEN']
      // Calculate the number of items to fetch.
      const limit = Math.max(BATCH_SIZE, stopIndex - startIndex + 1)
      const [startPage, stopPage] = indicesToPage(startIndex, stopIndex, limit)

      // Fetch new notifications with a maximum of 1000 items.
      const newItems = await dispatch(
        attachPromise(
          getNotifications(userId, filters, {
            limit: Math.min((stopPage - startPage + 1) * BATCH_SIZE, 1000),
            page: startPage,
          }),
        ),
      )

      // Update the total count of notifications.
      setTotalCount(newItems.totalCount)

      // Integrate the new items into the existing list.
      const updatedItems = fillIntoArray(
        items,
        pageToIndices(startPage, limit)[0],
        newItems.notifications,
        newItems.totalCount,
      )
      setItems(updatedItems)

      // Set the first notification as selected if none is currently selected.
      if (!selectedNotification) {
        setSelectedNotification(updatedItems[0])
      }

      // Determine if there are more pages to load.
      setHasNextPage(updatedItems.length < newItems.totalCount)
      setFetching(false)
    },
    [fetching, showArchived, dispatch, userId, items, selectedNotification],
  )

  const handleShowArchived = useCallback(async () => {
    // Toggle the showArchived state.
    const newShowArchivedValue = showArchived === 'false' ? 'true' : 'false'
    await setShowArchived(newShowArchivedValue)
    // Reset items and selected notification.
    setItems([])
    setSelectedNotification(undefined)

    // Load the first page of archived notifications.
    // When handleShowArchived is defined, it captures the current value of showArchived.
    // Even though showArchived is updated later, the captured value inside handleShowArchived remains the same.
    // So loadNextPage() is called with the old value of showArchived, showing the same notifications again.
    // To avoid this, we pass the new value of showArchived to loadNextPage() as an argument.
    loadNextPage(0, BATCH_SIZE, newShowArchivedValue)
  }, [loadNextPage, setShowArchived, showArchived])

  const handleArchive = useCallback(
    async (_, id) => {
      // Determine the filter to apply based on the showArchived state.
      const updateFilter =
        showArchived === 'true' ? 'NOTIFICATION_STATUS_SEEN' : 'NOTIFICATION_STATUS_ARCHIVED'

      // Update the status of the notification.
      await dispatch(attachPromise(updateNotificationStatus(userId, [id], updateFilter)))

      // Find the index of the archived notification.
      const index = items.findIndex(item => item.id === id)

      // Reload notifications starting from the archived one.
      await loadNextPage(
        index,
        index + BATCH_SIZE > items.length - 1 ? items.length - 1 : index + BATCH_SIZE,
      )

      // Update the selected notification to the one above the archived one,
      // unless there is only one item in the list.
      setSelectedNotification(totalCount === 1 ? undefined : items[Math.max(1, index - 1)])

      // Reset the list cache if available so that old items are discarded.
      if (listRef.current && listRef.current.resetloadMoreItemsCache) {
        listRef.current.resetloadMoreItemsCache()
      }
    },
    [showArchived, dispatch, userId, items, loadNextPage, totalCount],
  )

  // Load the first page of notifications when the component mounts.
  useEffect(() => {
    loadNextPage(0, BATCH_SIZE)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

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
          setSelectedNotification={setSelectedNotification}
          selectedNotification={selectedNotification}
          isArchive={showArchived === 'true'}
          listRef={listRef}
        />
        <div className="d-flex j-center">
          <Button
            onClick={handleShowArchived}
            naked
            message={showArchived === 'true' ? m.seeAll : m.seeArchived}
            className={style.notificationListChangeButton}
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
            setSelectedNotification={setSelectedNotification}
            selectedNotification={selectedNotification}
            isArchive={showArchived === 'true'}
            onArchive={handleArchive}
          />
        )}
      </div>
    </div>
  )
}

export default Notifications
