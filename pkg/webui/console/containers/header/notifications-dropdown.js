// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import React from 'react'
import { useSelector } from 'react-redux'
import classnames from 'classnames'
import { defineMessages } from 'react-intl'

import Link from '@ttn-lw/components/link'
import Status from '@ttn-lw/components/status'

import DateTime from '@ttn-lw/lib/components/date-time'
import Message from '@ttn-lw/lib/components/message'

import ttiNotification from '@console/components/notifications'

import notificationStyle from '@console/containers/notifications/notifications.styl'
import NOTIFICATION_STATUS from '@console/containers/notifications/notification-status'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  selectInboxNotifications,
  selectInboxNotificationsTotalCount,
} from '@console/store/selectors/notifications'

import style from './header.styl'

const m = defineMessages({
  description:
    'Showing last {numNotifications} of {totalNotifications} notifications • <Link>View all</Link>',
  noNotifications: 'All caught up!',
  noNotificationsDescription: 'You don’t have any notifications currently',
})

const NotificationsDropdown = () => {
  const dropdownItems = useSelector(selectInboxNotifications)
  const totalNotifications = useSelector(selectInboxNotificationsTotalCount)

  return (
    <>
      <div className={style.notificationsDropdownHeader}>
        <Message content={sharedMessages.notifications} />{' '}
        <Message
          className="c-text-neutral-semilight fw-normal fs-m"
          content={`(${totalNotifications})`}
        />
      </div>
      {dropdownItems && dropdownItems.length === 0 ? (
        <div className={style.emptyState}>
          <Message
            content={m.noNotifications}
            className="d-block text-center fw-bold c-text-neutral-semilight"
          />
          <Message
            content={m.noNotificationsDescription}
            className="d-block text-center fs-s c-text-neutral-light"
          />
        </div>
      ) : (
        <>
          {dropdownItems.slice(0, 3).map(notification => (
            <Link
              to={{
                pathname: `/notifications/inbox/${notification.id}`,
              }}
              key={notification.id}
              className={classnames(style.notificationsDropdownLink, 'd-flex')}
            >
              <div className={style.notificationsDropdownLinkIcon}>
                <ttiNotification.Icon
                  data={notification}
                  notificationType={notification.notification_type}
                />
              </div>
              <div className={style.notificationContainer}>
                <div className={classnames(style.title, 'fw-bold')}>
                  <ttiNotification.Title
                    data={notification}
                    notificationType={notification.notification_type}
                  />
                </div>
                <div
                  className={classnames(
                    notificationStyle.notificationPreviewContent,
                    style.previewContent,
                  )}
                >
                  <ttiNotification.Preview
                    data={notification}
                    notificationType={notification.notification_type}
                  />
                </div>

                <Status
                  pulse={false}
                  status="good"
                  className={classnames('d-flex al-center', {
                    [style.hideStatus]: [
                      NOTIFICATION_STATUS.SEEN,
                      NOTIFICATION_STATUS.ARCHIVED,
                    ].includes(notification.status),
                  })}
                  flipped
                >
                  <DateTime.Relative
                    relativeTimeStyle="short"
                    showAbsoluteAfter={2}
                    dateTimeProps={{
                      time: false,
                      dateFormatOptions: { month: '2-digit', day: '2-digit', year: 'numeric' },
                    }}
                    value={notification.created_at}
                    className="fs-s c-text-neutral-light"
                  />
                </Status>
              </div>
            </Link>
          ))}
          <div
            className={classnames(
              'p-cs-m c-text-neutral-semilight fs-s text-center c-bg-neutral-light',
              style.totalMessage,
            )}
          >
            <Message
              content={m.description}
              values={{
                numNotifications: dropdownItems.slice(0, 3).length,
                totalNotifications,
                Link: txt => (
                  <Link to={'/notifications/inbox'} primary className="td-none">
                    {txt}
                  </Link>
                ),
              }}
            />
          </div>
        </>
      )}
    </>
  )
}

export default NotificationsDropdown
