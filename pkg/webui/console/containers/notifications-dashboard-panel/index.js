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
import { createSelector } from 'reselect'
import classnames from 'classnames'

import Panel from '@ttn-lw/components/panel'
import Status from '@ttn-lw/components/status'

import FetchTable from '@ttn-lw/containers/fetch-table'

import DateTime from '@ttn-lw/lib/components/date-time'

import Notification from '@console/components/notifications'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getNotifications } from '@console/store/actions/notifications'

import {
  selectNotifications,
  selectTotalNotificationsCount,
  selectTotalUnseenCount,
} from '@console/store/selectors/notifications'
import { selectUserId } from '@console/store/selectors/logout'

import style from './notifications-dashboard-panel.styl'

const NotificationsDashboardPanel = () => {
  const totalUnseenNotifications = useSelector(selectTotalUnseenCount)
  const userId = useSelector(selectUserId)

  const MessageDecorator = () => (
    <span className={style.notificationPanelTotal}>{totalUnseenNotifications}</span>
  )

  const headers = [
    {
      name: 'notification',
      displayName: sharedMessages.message,
      width: 30,
      render: notification => (
        <div className="d-flex gap-cs-xs">
          {!notification.status && <Status pulse status="good" />}
          <div
            className={classnames('w-full', {
              [style.notificationPanelRead]: notification.status === 'NOTIFICATION_STATUS_SEEN',
            })}
          >
            <div className={style.notificationPanelTitle}>
              <Notification.Title
                data={notification}
                notificationType={notification.notification_type}
              />
            </div>
            <div className={style.notificationPanelPreview}>
              <Notification.Preview
                data={notification}
                notificationType={notification.notification_type}
              />
            </div>
          </div>
        </div>
      ),
    },
    {
      name: 'notification.created_at',
      displayName: sharedMessages.time,
      width: 10,
      render: date => <DateTime.Relative value={date} />,
    },
  ]

  const getItems = React.useCallback(
    () =>
      getNotifications(userId, ['NOTIFICATION_STATUS_UNSEEN', 'NOTIFICATION_STATUS_SEEN'], {
        limit: 5,
        page: 1,
      }),
    [userId],
  )

  const baseDataSelectors = createSelector(
    [selectNotifications, selectTotalNotificationsCount],
    (notifications, totalCount) => {
      const decoratedNotifications = []
      for (const notification of notifications) {
        decoratedNotifications.push({
          notification,
          id: notification.id,
        })
      }
      return {
        notifications: decoratedNotifications,
        totalCount,
        mayAdd: false,
      }
    },
  )

  return (
    <Panel
      title={sharedMessages.notifications}
      path="/notifications"
      icon="inbox"
      buttonTitle="View all"
      messageDecorators={totalUnseenNotifications > 0 ? <MessageDecorator /> : undefined}
    >
      <FetchTable
        entity="notifications"
        headers={headers}
        pageSize={5}
        baseDataSelector={baseDataSelectors}
        getItemsAction={getItems}
        getItemPathPrefix={() => '/notifications'}
        rowLinkState={notification => ({
          notification: notification.notification,
        })}
        small
      />
    </Panel>
  )
}

export default NotificationsDashboardPanel
