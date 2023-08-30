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
import { Col, Row } from 'react-grid-system'
import classNames from 'classnames'
import { defineMessages } from 'react-intl'

import Icon from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  getUnseenNotifications,
  updateNotificationStatus,
} from '@console/store/actions/notifications'

import {
  selectNotifications,
  selectTotalUnseenCount,
  selectUnseenIds,
} from '@console/store/selectors/notifications'
import { selectUserId } from '@console/store/selectors/logout'

import style from '../notifications.styl'

import NotificationListItem from './list-item'

const m = defineMessages({
  archived: 'Archived notifications',
  markAllAsRead: 'Mark all as read',
})

const NotificationList = ({
  setSelectedNotification,
  selectedNotification,
  setShowContent,
  fetchItems,
  isArchive,
  setShowListColumn,
}) => {
  const userId = useSelector(selectUserId)
  const notifications = useSelector(selectNotifications)
  const unseenIds = useSelector(selectUnseenIds)
  const totalUnseenCount = useSelector(selectTotalUnseenCount)
  const dispatch = useDispatch()
  const isMobile = window.innerWidth < 768

  const handleClick = useCallback(
    async (e, id) => {
      setShowListColumn(!isMobile)
      setShowContent(true)
      setSelectedNotification(notifications.find(notification => notification.id === id))
      if (!isArchive) {
        await dispatch(
          attachPromise(updateNotificationStatus(userId, [id], 'NOTIFICATION_STATUS_SEEN')),
        )
        setTimeout(async () => {
          await fetchItems()
          await dispatch(attachPromise(getUnseenNotifications(userId)))
        }, 300)
      }
    },
    [
      notifications,
      dispatch,
      userId,
      fetchItems,
      setShowContent,
      setSelectedNotification,
      isArchive,
      setShowListColumn,
      isMobile,
    ],
  )

  const handleMarkAllAsSeen = useCallback(async () => {
    await dispatch(
      attachPromise(updateNotificationStatus(userId, unseenIds, 'NOTIFICATION_STATUS_SEEN')),
    )
    setTimeout(async () => {
      await fetchItems()
      await dispatch(attachPromise(getUnseenNotifications(userId)))
    }, 300)
  }, [dispatch, userId, unseenIds, fetchItems])

  const classes = classNames(style.notificationHeaderIcon, 'm-0', {
    [style.notifications]: !isArchive,
    [style.archived]: isArchive,
  })

  return (
    <>
      <Row justify="between" className={classNames(style.notificationHeader, 'm-0')}>
        <Col className="pl-cs-s pr-cs-s d-flex">
          <Icon icon={isArchive ? 'archive' : 'notifications'} nudgeDown className={classes} />
          <Message
            component="h3"
            content={isArchive ? m.archived : sharedMessages.notifications}
            className="m-0"
          />
          {Boolean(totalUnseenCount) && !isArchive && (
            <span className={style.totalNotifications} data-test-id="total-unseen-notifications">
              {totalUnseenCount}
            </span>
          )}
        </Col>
        {!isArchive && (
          <Button
            icon="visibility"
            onClick={handleMarkAllAsSeen}
            message={m.markAllAsRead}
            className="mr-cs-s"
          />
        )}
      </Row>
      {notifications.map(notification => (
        <Row direction="column" key={notification.id} className="m-0 p-0">
          <NotificationListItem
            notification={notification}
            selectedNotification={selectedNotification}
            handleClick={handleClick}
          />
        </Row>
      ))}
    </>
  )
}

NotificationList.propTypes = {
  fetchItems: PropTypes.func.isRequired,
  isArchive: PropTypes.bool.isRequired,
  selectedNotification: PropTypes.shape({
    id: PropTypes.string,
  }),
  setSelectedNotification: PropTypes.func.isRequired,
  setShowContent: PropTypes.func.isRequired,
  setShowListColumn: PropTypes.func.isRequired,
}

NotificationList.defaultProps = {
  selectedNotification: undefined,
}

export default NotificationList
