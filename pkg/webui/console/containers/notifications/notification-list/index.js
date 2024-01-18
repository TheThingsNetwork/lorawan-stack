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

import React, { useCallback } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import classNames from 'classnames'
import { defineMessages } from 'react-intl'
import { FixedSizeList as List } from 'react-window'
import InfiniteLoader from 'react-window-infinite-loader'
import AutoSizer from 'react-virtualized-auto-sizer'

import Icon from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { updateNotificationStatus } from '@console/store/actions/notifications'

import { selectTotalUnseenCount, selectUnseenIds } from '@console/store/selectors/notifications'
import { selectUserId } from '@console/store/selectors/logout'

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
  setSelectedNotification,
  selectedNotification,
  isArchive,
  totalCount,
  listRef,
}) => {
  const userId = useSelector(selectUserId)
  const unseenIds = useSelector(selectUnseenIds)
  const totalUnseenCount = useSelector(selectTotalUnseenCount)
  const dispatch = useDispatch()
  const isItemLoaded = useCallback(
    index => (items.length > 0 ? !hasNextPage || index < items.length : false),
    [hasNextPage, items],
  )

  // If the total count is not known, we assume that there are 100 items.
  // Otherwise, if totalCount is 0, it means the list is empty and we should not have a total count.
  const itemCount = totalCount >= 0 ? totalCount : 100

  const handleClick = useCallback(
    async (_, id) => {
      setSelectedNotification(items.find(notification => notification.id === id))
      if (!isArchive && unseenIds.includes(id)) {
        await dispatch(
          attachPromise(updateNotificationStatus(userId, [id], 'NOTIFICATION_STATUS_SEEN')),
        )
      }
    },
    [items, dispatch, userId, setSelectedNotification, isArchive, unseenIds],
  )

  const handleMarkAllAsSeen = useCallback(async () => {
    if (unseenIds.length > 0) {
      await dispatch(
        attachPromise(updateNotificationStatus(userId, unseenIds, 'NOTIFICATION_STATUS_SEEN')),
      )
    }
  }, [dispatch, userId, unseenIds])

  const classes = classNames(styles.notificationHeaderIcon)

  const Item = ({ index, style }) =>
    isItemLoaded(index) ? (
      <div style={style}>
        <NotificationListItem
          notification={items[index]}
          selectedNotification={selectedNotification}
          handleClick={handleClick}
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

  return (
    <>
      <div className={styles.notificationHeader}>
        <div className={classNames(styles.notificationHeaderTitle, 'd-flex gap-cs-xxs')}>
          <Icon icon={isArchive ? 'archive' : 'inbox'} nudgeDown className={classes} />
          <Message
            content={isArchive ? m.archived : sharedMessages.notifications}
            component="p"
            className="m-0"
          />
          {Boolean(totalUnseenCount) && !isArchive && (
            <span className={styles.totalNotifications} data-test-id="total-unseen-notifications">
              {totalUnseenCount}
            </span>
          )}
        </div>
        {!isArchive && (
          <Button
            secondary
            icon="visibility"
            onClick={handleMarkAllAsSeen}
            message={m.markAllAsRead}
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
                  height={height}
                  width={width}
                  itemSize={98}
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
  isArchive: PropTypes.bool.isRequired,
  items: PropTypes.array.isRequired,
  listRef: PropTypes.shape({ current: PropTypes.shape({}) }).isRequired,
  loadNextPage: PropTypes.func.isRequired,
  selectedNotification: PropTypes.shape({
    id: PropTypes.string,
  }),
  setSelectedNotification: PropTypes.func.isRequired,
  totalCount: PropTypes.number.isRequired,
}

NotificationList.defaultProps = {
  selectedNotification: undefined,
}

export default NotificationList
