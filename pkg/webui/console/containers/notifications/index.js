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

import React, { useCallback, useEffect, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { Col, Row } from 'react-grid-system'
import classNames from 'classnames'

import Button from '@ttn-lw/components/button'
import Icon from '@ttn-lw/components/icon'
import Status from '@ttn-lw/components/status'
import Pagination from '@ttn-lw/components/pagination'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'

import { getNotifications, updateNotificationStatus } from '@console/store/actions/notifications'

import { selectUserId } from '@console/store/selectors/logout'
import {
  selectNotifications,
  selectTotalNotificationsCount,
  selectTotalUnseenCount,
  selectUnseenNotifications,
} from '@console/store/selectors/notifications'

import style from './notifications.styl'

const pageSize = 6

const DEFAULT_PAGE = 1

const pageValidator = page => (!Boolean(page) || page < 0 ? DEFAULT_PAGE : page)

const NotificationsContainer = ({ setPage, page }) => {
  const userId = useSelector(selectUserId)
  const notifications = useSelector(selectNotifications)
  const totalNotifications = useSelector(selectTotalNotificationsCount)
  const totalUnseenCount = useSelector(selectTotalUnseenCount)
  const unseenNotifications = useSelector(selectUnseenNotifications)
  const unseenIds = Object.keys(unseenNotifications)
  const dispatch = useDispatch()
  const [archiving, setArchiving] = useState(false)
  const [selectedNotification, setSelectedNotification] = useState(undefined)
  const [unseenCount, setUnseenCount] = useState(totalUnseenCount)

  const fetchItems = useCallback(async () => {
    await dispatch(
      attachPromise(
        getNotifications(userId, ['NOTIFICATION_STATUS_UNSEEN', 'NOTIFICATION_STATUS_SEEN'], {
          limit: pageSize,
          page,
        }),
      ),
    )
  }, [dispatch, userId, page])

  useEffect(() => {
    fetchItems()
  }, [dispatch, userId, page, fetchItems])

  const onPageChange = useCallback(
    page => {
      setPage(pageValidator(page))
    },
    [setPage],
  )

  const handleClick = useCallback(
    async (e, id) => {
      setArchiving(false)
      setSelectedNotification(notifications.find(notification => notification.id === id))
      await dispatch(
        attachPromise(updateNotificationStatus(userId, [id], 'NOTIFICATION_STATUS_SEEN')),
      )
      setTimeout(async () => await fetchItems(), 300)
      if (unseenCount > 0) {
        setUnseenCount(unseenCount => (unseenCount === 1 ? 0 : unseenCount - 1))
      }
    },
    [notifications, dispatch, userId, fetchItems, setUnseenCount, unseenCount],
  )

  const handleArchive = useCallback(
    async (e, id) => {
      setArchiving(true)
      await dispatch(
        attachPromise(updateNotificationStatus(userId, [id], 'NOTIFICATION_STATUS_ARCHIVED')),
      )
      setTimeout(async () => await fetchItems(), 300)
    },
    [dispatch, userId, fetchItems],
  )

  const handleMarkAllAsSeen = useCallback(async () => {
    await dispatch(
      attachPromise(updateNotificationStatus(userId, unseenIds, 'NOTIFICATION_STATUS_SEEN')),
    )
    setTimeout(async () => await fetchItems(), 300)
    setUnseenCount(0)
  }, [dispatch, userId, unseenIds, fetchItems, setUnseenCount])

  return (
    <Row sm={12} lg={20} className={classNames(style.notificationsContainer, 'm-0')}>
      <Col lg={6} xl={5} className={classNames(style.notificationList, 'mt-cs-l', 'mb-cs-l')}>
        <Row justify="between" className={classNames(style.notificationHeader, 'm-0')}>
          <div>
            <Icon icon="notifications" />
            <Message component="strong" content={'Notifications'} />
            {Boolean(unseenCount) && (
              <span className={style.totalNotifications}>{unseenCount}</span>
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
        <Row className={style.notificationFooter}>
          <Pagination
            className={style.pagination}
            pageCount={Math.ceil(totalNotifications / pageSize) || 1}
            onPageChange={onPageChange}
            disableInitialCallback
            pageRangeDisplayed={2}
            forcePage={page}
          />
          <Message content="See archived messages" />
        </Row>
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

NotificationsContainer.propTypes = {
  page: PropTypes.number.isRequired,
  setPage: PropTypes.func.isRequired,
}

export default NotificationsContainer
