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

import Icon from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'

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

const NotificationList = ({
  setSelectedNotification,
  selectedNotification,
  setArchiving,
  fetchItems,
}) => {
  const userId = useSelector(selectUserId)
  const notifications = useSelector(selectNotifications)
  const unseenIds = useSelector(selectUnseenIds)
  const totalUnseenCount = useSelector(selectTotalUnseenCount)
  const dispatch = useDispatch()

  const handleClick = useCallback(
    async (e, not) => {
      setArchiving(false)
      setSelectedNotification(notifications.find(notification => notification.id === not.id))
      await dispatch(
        attachPromise(updateNotificationStatus(userId, [not.id], 'NOTIFICATION_STATUS_SEEN')),
      )
      setTimeout(async () => {
        await fetchItems()
        await dispatch(attachPromise(getUnseenNotifications(userId)))
      }, 300)
    },
    [notifications, dispatch, userId, fetchItems, setArchiving, setSelectedNotification],
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

  return (
    <>
      <Row justify="between" className={classNames(style.notificationHeader, 'm-0')}>
        <Col className="pl-cs-s pr-cs-s d-flex">
          <Icon icon="notifications" textPaddedRight nudgeDown className={style.notificationIcon} />
          <Message component="h3" content={'Notifications'} className="m-0" />
          {Boolean(totalUnseenCount) && (
            <span className={style.totalNotifications}>{totalUnseenCount}</span>
          )}
        </Col>
        <Button
          icon="visibility"
          onClick={handleMarkAllAsSeen}
          message="Mark all as read"
          className="mr-cs-s"
        />
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
  selectedNotification: PropTypes.shape({
    id: PropTypes.string,
  }),
  setArchiving: PropTypes.func.isRequired,
  setSelectedNotification: PropTypes.func.isRequired,
}

NotificationList.defaultProps = {
  selectedNotification: undefined,
}

export default NotificationList
