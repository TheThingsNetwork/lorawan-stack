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

import React, { useCallback, useMemo } from 'react'
import { useDispatch } from 'react-redux'
import classNames from 'classnames'
import { defineMessages } from 'react-intl'
import { FixedSizeList as List } from 'react-window'
import InfiniteLoader from 'react-window-infinite-loader'
import AutoSizer from 'react-virtualized-auto-sizer'
import { useParams } from 'react-router-dom'

import { IconEye } from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { markAllAsSeen } from '@console/store/actions/notifications'

import styles from '../notifications.styl'

import { NotificationListItem, NotificationListSpinner } from './list-item'

const m = defineMessages({
  archived: 'Archived notifications',
  markAllAsRead: 'Mark all as read',
})

const NotificationList = ({
  items,
  hasNextPage,
  loadNextPage,
  selectedNotification,
  totalCount,
  listRef,
  updatePendingNotificationIds,
  unreadNotificationsCount,
}) => {
  const { category } = useParams()
  const isArchive = category === 'archived'
  const dispatch = useDispatch()
  const isItemLoaded = useCallback(
    index => (items.length > 0 ? !hasNextPage || index < items.length : false),
    [hasNextPage, items],
  )

  // If the total count is not known, we assume that there are 100 items.
  // Otherwise, if totalCount is 0, it means the list is empty and we should not have a total count.
  const itemCount = totalCount >= 0 ? totalCount : 100

  const handleMarkAllAsSeen = useCallback(async () => {
    await dispatch(attachPromise(markAllAsSeen()))
  }, [dispatch])

  const isSelected = notification =>
    notification && selectedNotification && selectedNotification.id === notification.id
  const isNextSelected = notification => {
    const index = items.findIndex(item => item.id === notification.id)
    return notification && index + 1 < items.length && isSelected(items[index + 1])
  }

  const Item = ({ index, style }) =>
    isItemLoaded(index) ? (
      <div style={style}>
        <NotificationListItem
          notification={items[index]}
          isSelected={isSelected(items[index])}
          isNextSelected={isNextSelected(items[index])}
          isUpdatePending={Boolean(updatePendingNotificationIds.find(id => id === items[index].id))}
        />
      </div>
    ) : (
      <div style={style}>
        <NotificationListSpinner />
      </div>
    )

  Item.propTypes = {
    index: PropTypes.number.isRequired,
    style: PropTypes.shape({}).isRequired,
  }

  const notificationCount = useMemo(() => {
    if (!isArchive && unreadNotificationsCount) {
      return (
        <span className={styles.totalNotifications} data-test-id="total-unseen-notifications">
          {unreadNotificationsCount}
        </span>
      )
    }

    return null
  }, [isArchive, unreadNotificationsCount])

  return (
    <>
      <div className={styles.notificationHeader}>
        <div className={classNames(styles.notificationHeaderTitle, 'd-flex gap-cs-xxs')}>
          <Message
            content={isArchive ? m.archived : sharedMessages.notifications}
            component="p"
            className="m-0 fs-l"
          />
          {notificationCount}
        </div>
        {!isArchive && (
          <Button
            secondary
            icon={IconEye}
            onClick={handleMarkAllAsSeen}
            message={m.markAllAsRead}
            disabled={!unreadNotificationsCount}
          />
        )}
      </div>
      <div className="flex-grow">
        <AutoSizer>
          {({ height, width }) => (
            <InfiniteLoader
              ref={listRef}
              loadMoreItems={loadNextPage}
              isItemLoaded={isItemLoaded}
              itemCount={itemCount}
              minimumBatchSize={50}
            >
              {({ onItemsRendered, ref }) => (
                <List
                  className={styles.notificationListList}
                  height={height}
                  width={width}
                  itemSize={88}
                  ref={ref}
                  itemCount={itemCount}
                  onItemsRendered={onItemsRendered}
                >
                  {Item}
                </List>
              )}
            </InfiniteLoader>
          )}
        </AutoSizer>
      </div>
    </>
  )
}

NotificationList.propTypes = {
  hasNextPage: PropTypes.bool.isRequired,
  items: PropTypes.array.isRequired,
  listRef: PropTypes.shape({ current: PropTypes.shape({}) }).isRequired,
  loadNextPage: PropTypes.func.isRequired,
  selectedNotification: PropTypes.shape({
    id: PropTypes.string,
  }),
  totalCount: PropTypes.number.isRequired,
  unreadNotificationsCount: PropTypes.number.isRequired,
  updatePendingNotificationIds: PropTypes.arrayOf(PropTypes.string).isRequired,
}

NotificationList.defaultProps = {
  selectedNotification: undefined,
}

export default NotificationList
