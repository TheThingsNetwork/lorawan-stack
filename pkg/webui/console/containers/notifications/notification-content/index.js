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

import React, { useEffect } from 'react'
import { useSelector } from 'react-redux'
import classNames from 'classnames'
import { defineMessages } from 'react-intl'
import { useParams } from 'react-router-dom'

import LAYOUT from '@ttn-lw/constants/layout'

import { IconChevronLeft, IconArchive } from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'

import DateTime from '@ttn-lw/lib/components/date-time'

import Notification from '@console/components/notifications'

import PropTypes from '@ttn-lw/lib/prop-types'

import { selectUserId } from '@console/store/selectors/logout'

import style from '../notifications.styl'

const m = defineMessages({
  archive: 'Archive',
  unarchive: 'Unarchive',
})

const NotificationContent = ({ onArchive, selectedNotification }) => {
  const [isMediumScreen, setIsMediumScreen] = React.useState(
    window.innerWidth < LAYOUT.BREAKPOINTS.L,
  )
  const userId = useSelector(selectUserId)
  const { category } = useParams()
  const isArchive = category === 'archived'

  useEffect(() => {
    const handleResize = () => {
      setIsMediumScreen(window.innerWidth < LAYOUT.BREAKPOINTS.L)
    }
    window.addEventListener('resize', handleResize)

    return () => window.removeEventListener('resize', handleResize)
  }, [])

  const dateFormat = isMediumScreen
    ? { day: '2-digit', month: '2-digit', year: 'numeric' }
    : { day: '2-digit', month: 'long', year: 'numeric' }

  const archiveMessage = isMediumScreen ? undefined : isArchive ? m.unarchive : m.archive

  return (
    <>
      <div className={style.notificationHeader}>
        <div className={style.notificationHeaderTitle}>
          <Button.Link
            to="/notifications/inbox"
            icon={IconChevronLeft}
            className="md-lg:d-flex d-none"
            naked
          />
          <div>
            <p className="m-0">
              <Notification.Title
                data={selectedNotification}
                notificationType={selectedNotification.notification_type}
              />
            </p>
            <DateTime
              value={selectedNotification.created_at}
              dateFormatOptions={dateFormat}
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
            dateFormatOptions={dateFormat}
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
            onClick={onArchive}
            message={archiveMessage}
            icon={IconArchive}
            value={selectedNotification.id}
            secondary
          />
        </div>
      </div>
      <div className="p-cs-xl md-lg:p-cs-l">
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
  onArchive: PropTypes.func.isRequired,
  selectedNotification: PropTypes.shape({
    id: PropTypes.string.isRequired,
    created_at: PropTypes.string.isRequired,
    notification_type: PropTypes.string.isRequired,
    status: PropTypes.string,
  }).isRequired,
}

export default NotificationContent
