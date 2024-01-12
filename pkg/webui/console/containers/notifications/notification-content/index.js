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

import Button from '@ttn-lw/components/button'

import DateTime from '@ttn-lw/lib/components/date-time'

import Notification from '@console/components/notifications'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'

import { updateNotificationStatus } from '@console/store/actions/notifications'

import { selectUserId } from '@console/store/selectors/logout'

import style from '../notifications.styl'

const m = defineMessages({
  archive: 'Archive',
  unarchive: 'Unarchive',
})

const NotificationContent = ({
  isArchive,
  setSelectedNotification,
  selectedNotification,
  setIsArchiving,
}) => {
  const userId = useSelector(selectUserId)
  const dispatch = useDispatch()

  const handleArchive = useCallback(
    async (e, id) => {
      setIsArchiving(true)
      const updateFilter = isArchive ? 'NOTIFICATION_STATUS_SEEN' : 'NOTIFICATION_STATUS_ARCHIVED'
      await dispatch(attachPromise(updateNotificationStatus(userId, [id], updateFilter)))
    },
    [dispatch, userId, isArchive, setIsArchiving],
  )

  const handleBack = useCallback(() => {
    setSelectedNotification(undefined)
  }, [setSelectedNotification])

  return (
    <>
      <div className={style.notificationHeader}>
        <div className={style.notificationHeaderTitle}>
          <Button icon="arrow_back_ios" naked onClick={handleBack} className={style.backButton} />
          <div>
            <p className="m-0">
              <Notification.Title
                data={selectedNotification}
                notificationType={selectedNotification.notification_type}
              />
            </p>
            <DateTime
              value={selectedNotification.created_at}
              dateFormatOptions={{ day: 'numeric', month: 'long', year: 'numeric' }}
              timeFormatOptions={{
                hour: 'numeric',
                minute: 'numeric',
                hourCycle: 'h23',
              }}
              className={classNames(style.notificationHeaderDate, {
                [style.notificationSelectedMobile]: Boolean(selectedNotification),
              })}
            />
          </div>
        </div>
        <div className={style.notificationHeaderAction}>
          <DateTime
            value={selectedNotification.created_at}
            dateFormatOptions={{ day: 'numeric', month: 'long', year: 'numeric' }}
            timeFormatOptions={{
              hour: 'numeric',
              minute: 'numeric',
              hourCycle: 'h23',
            }}
            className={classNames(style.notificationHeaderDate, {
              [style.notificationSelected]: Boolean(selectedNotification),
            })}
          />
          <Button
            onClick={handleArchive}
            message={isArchive ? m.unarchive : m.archive}
            icon="archive"
            value={selectedNotification.id}
            secondary
          />
        </div>
      </div>
      <div className="p-cs-xl">
        <Notification.Content
          receiver={userId}
          data={selectedNotification}
          notificationType={selectedNotification.notification_type}
        />
      </div>
    </>
  )
}

NotificationContent.propTypes = {
  isArchive: PropTypes.bool.isRequired,
  selectedNotification: PropTypes.shape({
    id: PropTypes.string.isRequired,
    created_at: PropTypes.string.isRequired,
    notification_type: PropTypes.string.isRequired,
    status: PropTypes.string,
  }).isRequired,
  setIsArchiving: PropTypes.func.isRequired,
  setSelectedNotification: PropTypes.func.isRequired,
}

export default NotificationContent
