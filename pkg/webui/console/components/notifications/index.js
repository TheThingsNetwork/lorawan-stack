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

import React, { useMemo } from 'react'

import PropTypes from '@ttn-lw/lib/prop-types'

import { getNotificationContent, getNotificationPreview, getNotificationTitle } from './utils'

const NotificationContent = ({ reciever, notificationType, data }) => {
  const NotificationContent = useMemo(
    () => getNotificationContent(notificationType),
    [notificationType],
  )

  return <NotificationContent notificationData={data} reciever={reciever} />
}

NotificationContent.propTypes = {
  data: PropTypes.object.isRequired,
  notificationType: PropTypes.string.isRequired,
  reciever: PropTypes.object.isRequired,
}

const NotificationTitle = ({ notificationType, data }) => {
  const NotificationTitle = useMemo(
    () => getNotificationTitle(notificationType),
    [notificationType],
  )

  return <NotificationTitle notificationData={data} />
}

NotificationTitle.propTypes = {
  data: PropTypes.object.isRequired,
  notificationType: PropTypes.string.isRequired,
}

const NotificationPreview = ({ notificationType, data }) => {
  const NotificationPreview = useMemo(
    () => getNotificationPreview(notificationType),
    [notificationType],
  )

  return <NotificationPreview notificationData={data} />
}

NotificationPreview.propTypes = {
  data: PropTypes.object.isRequired,
  notificationType: PropTypes.string.isRequired,
}

Notification.Content = NotificationContent
Notification.Title = NotificationTitle
Notification.Preview = NotificationPreview

export default Notification
