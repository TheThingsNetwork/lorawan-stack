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
import { useParams } from 'react-router-dom'

import Icon from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { markAllAsSeen, updateNotificationStatus } from '@console/store/actions/notifications'

import { selectTotalUnseenCount } from '@console/store/selectors/notifications'

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
}) => {
  const totalUnseenCount = useSelector(selectTotalUnseenCount)
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

  const handleClick = useCallback(
    async (_, id) => {
      const clickedNotification = items.find(notification => notification.id === id)
      if (!isArchive && !('status' in clickedNotification) && totalUnseenCount > 0) {
        await dispatch(attachPromise(updateNotificationStatus([id], 'NOTIFICATION_STATUS_SEEN')))
      }
    },
    [items, dispatch, isArchive, totalUnseenCount],
  )

  const handleMarkAllAsSeen = useCallback(async () => {
    if (totalUnseenCount > 0) {
      await dispatch(attachPromise(markAllAsSeen()))
    }
  }, [dispatch, totalUnseenCount])

  const classes = classNames(styles.notificationHeaderIcon)

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
          onClick={handleClick}
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
}

NotificationList.defaultProps = {
  selectedNotification: undefined,
}

export default NotificationList
