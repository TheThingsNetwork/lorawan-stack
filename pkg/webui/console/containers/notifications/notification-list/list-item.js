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

import React from 'react'
import classNames from 'classnames'
import { useSelector } from 'react-redux'

import Button from '@ttn-lw/components/button'
import Status from '@ttn-lw/components/status'
import Spinner from '@ttn-lw/components/spinner'

import DateTime from '@ttn-lw/lib/components/date-time'

import Notification from '@console/components/notifications'

import PropTypes from '@ttn-lw/lib/prop-types'

import { selectUnseenIds } from '@console/store/selectors/notifications'

import style from '../notifications.styl'

export const NotificationListItem = ({ notification, isSelected, isNextSelected, handleClick }) => {
  const unseenIds = useSelector(selectUnseenIds)
  const showUnseenStatus = unseenIds.includes(notification.id)
  const classes = classNames(style.notificationPreview, {
    [style.notificationSelected]: isSelected,
    [style.notificationNextSelected]: isNextSelected,
  })
  const titleClasses = classNames(style.notificationPreviewTitle, {
    [style.notificationRead]: !showUnseenStatus,
  })
  const previewClasses = classNames(style.notificationPreviewContent, {
    [style.notificationRead]: !showUnseenStatus,
  })

  return (
    <Button
      key={notification.id}
      onClick={handleClick}
      value={notification.id}
      className={classes}
      data-test-id="notification-list-item"
    >
      {showUnseenStatus && <Status pulse={false} status="good" className={style.unseenMark} />}
      <div className="w-full">
        <div className={titleClasses}>
          <div className={style.notificationPreviewTitleText}>
            <Notification.Title
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
        <div className={previewClasses}>
          <Notification.Preview
            data={notification}
            notificationType={notification.notification_type}
          />
        </div>
      </div>
    </Button>
  )
}

NotificationListItem.propTypes = {
  handleClick: PropTypes.func.isRequired,
  isNextSelected: PropTypes.bool,
  isSelected: PropTypes.bool,
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
