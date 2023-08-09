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

import Button from '@ttn-lw/components/button'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import { getNotifications, updateNotificationStatus } from '@console/store/actions/notifications'

import { selectUserId } from '@console/store/selectors/logout'
import {
  selectNotifications,
  selectTotalUnseenCount,
  selectUnseenNotifications,
} from '@console/store/selectors/notifications'

import style from './notifications.styl'
import Status from '@ttn-lw/components/status'

const NotificationsContainer = () => {
  const userId = useSelector(selectUserId)
  const notifications = useSelector(selectNotifications)
  const totalUnseenCount = useSelector(selectTotalUnseenCount)
  const unseenNotifications = useSelector(selectUnseenNotifications)
  const unseenIds = Object.keys(unseenNotifications)
  const dispatch = useDispatch()
  const [archiving, setArchiving] = React.useState(false)
  const [selectedNotification, setSelectedNotification] = React.useState(undefined)

  const handleClick = useCallback(
    async (e, id) => {
      setArchiving(false)
      setSelectedNotification(notifications.find(notification => notification.id === id))
      await dispatch(updateNotificationStatus(userId, [id], 'NOTIFICATION_STATUS_SEEN'))
      setTimeout(async () => await dispatch(getNotifications(userId)), 1000)
    },
    [notifications, dispatch, userId],
  )

  const handleArchive = useCallback(
    async (e, id) => {
      setArchiving(true)
      await dispatch(updateNotificationStatus(userId, [id], 'NOTIFICATION_STATUS_ARCHIVED'))
      setTimeout(async () => await dispatch(getNotifications(userId)), 300)
    },
    [dispatch, userId],
  )

  const handleMarkAllAsSeen = useCallback(async () => {
    await dispatch(updateNotificationStatus(userId, unseenIds, 'NOTIFICATION_STATUS_SEEN'))
    setTimeout(async () => await dispatch(getNotifications(userId)), 300)
  }, [dispatch, userId, unseenIds])

  return (
    <Row sm={12} lg={20} className={classNames(style.notificationsContainer, 'm-0')}>
      <Col lg={6} xl={5} className={classNames(style.notificationList, 'mt-cs-l', 'mb-cs-l')}>
        <Row justify="between" className={classNames(style.notificationHeader, 'm-0')}>
          <div>
            <Icon icon="notifications" />
            <Message component="strong" content={'Notifications'} />
            {Boolean(totalUnseenCount) && (
              <span className={style.totalNotifications}>{totalUnseenCount}</span>
            )}
          </div>
          <Button onClick={handleMarkAllAsSeen} message="Mark all as read" />
        </Row>
        <Col className={classNames(style.notificationItems, 'm-0', 'p-0')}>
          {notifications.map(notification => {
            const classes = classNames(style.notificationPreview, 'm-0', 'p-0', {
              [style.selected]: selectedNotification?.id === notification.id,
            })
            return (
              <Button
                key={notification.id}
                onClick={handleClick}
                value={notification.id}
                className={classes}
              >
                {(!('status' in notification) ||
                  notification.status === 'NOTIFICATION_STATUS_UNSEEN') && (
                  <Status
                    pulse={false}
                    status="good"
                    className={classNames('mr-cs-xs', style.unseenMark)}
                  />
                )}
                <Message content={notification.notification_type} />
              </Button>
            )
          })}
        </Col>
      </Col>
      <Col className={classNames(style.notificationContent, 'mt-cs-l', 'mb-cs-l', 'p-0')}>
        {selectedNotification && !archiving && (
          <>
            <Row justify="between" className={classNames(style.notificationHeader, 'm-0')}>
              <Message component="strong" content={'Title'} />
              <div>
                <DateTime value={selectedNotification.created_at} />
                <Button
                  onClick={handleArchive}
                  message="Archive"
                  icon="archive"
                  value={selectedNotification.id}
                />
              </div>
            </Row>
            <div style={{ padding: '20px' }}>
              Dear {userId}, <br />A new API key has been created for your{' '}
              {Object.keys(selectedNotification.entity_ids)[0].replace('_ids', '')}"
              {selectedNotification.entity_ids.application_ids.application_id}". <br />
              API Key ID: {selectedNotification.data.id}
              <br />
              API Key Name: {selectedNotification.data.name}
              <br />
              Rights:
              <br />
              {selectedNotification.data.rights.map((right, index) => (
                <div key={index}>
                  - {right}
                  <br />
                </div>
              ))}
              You can go to ... to view and edit this API key in the Console.
            </div>
          </>
        )}
      </Col>
    </Row>
  )
}

export default NotificationsContainer
