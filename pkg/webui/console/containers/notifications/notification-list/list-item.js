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
import classNames from 'classnames'
import { useParams } from 'react-router-dom'

import Button from '@ttn-lw/components/button'
import Status from '@ttn-lw/components/status'
import Spinner from '@ttn-lw/components/spinner'

import DateTime from '@ttn-lw/lib/components/date-time'

import ttiNotification from '@console/components/notifications'

import NOTIFICATION_STATUS from '@console/containers/notifications/notification-status'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from '../notifications.styl'

export const NotificationListItem = ({
  notification,
  isSelected,
  isNextSelected,
  isUpdatePending,
}) => {
  const { category } = useParams()
  const showUnseenStatus =
    ![NOTIFICATION_STATUS.SEEN, NOTIFICATION_STATUS.ARCHIVED].includes(notification.status) &&
    !isUpdatePending
  const classes = classNames(style.notificationPreview, {
    [style.notificationSelected]: isSelected,
    [style.notificationNextSelected]: isNextSelected,
    [style.unseen]: showUnseenStatus,
  })

  return (
    <Button.Link
      key={notification.id}
      to={`/notifications/${category}/${notification.id}`}
      className={classes}
      data-test-id="notification-list-item"
      value={notification.id}
    >
      <div className="w-full pos-relative">
        {showUnseenStatus && <Status pulse={false} status="good" className={style.unseenMark} />}
        <div className={style.notificationPreviewTitle}>
          <div className={style.notificationPreviewTitleText}>
            <ttiNotification.Title
              data={notification}
              notificationType={notification.notification_type}
            />
          </div>
          <div>
            <DateTime.Relative
              relativeTimeStyle="short"
              showAbsoluteAfter={2}
              dateTimeProps={{
                time: false,
                dateFormatOptions: { month: '2-digit', day: '2-digit', year: 'numeric' },
              }}
              value={notification.created_at}
              className={style.notificationTime}
            />
          </div>
        </div>
        <div className={style.notificationPreviewContent}>
          <ttiNotification.Preview
            data={notification}
            notificationType={notification.notification_type}
          />
        </div>
      </div>
    </Button.Link>
  )
}

NotificationListItem.propTypes = {
  isNextSelected: PropTypes.bool,
  isSelected: PropTypes.bool,
  isUpdatePending: PropTypes.bool.isRequired,
  notification: PropTypes.shape({
    id: PropTypes.string,
    created_at: PropTypes.string,
    notification_type: PropTypes.string,
    status: PropTypes.string,
  }).isRequired,
}

NotificationListItem.defaultProps = {
  isSelected: false,
  isNextSelected: false,
}

export const NotificationListSpinner = () => {
  const classes = classNames(style.notificationPreview, 'm-0', 'p-0')

  return (
    <div className={classes}>
      <Spinner faded micro center />
    </div>
  )
}
