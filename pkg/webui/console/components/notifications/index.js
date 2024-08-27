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

import React, { useMemo } from 'react'

import PropTypes from '@ttn-lw/lib/prop-types'

import { getNotification } from './utils'

const NotificationContent = ({ receiver, notificationType, data }) => {
  const NotificationContent = useMemo(() => getNotification(notificationType), [notificationType])

  return <NotificationContent notificationData={data} receiver={receiver} />
}

NotificationContent.propTypes = {
  data: PropTypes.shape({}).isRequired,
  notificationType: PropTypes.string.isRequired,
  receiver: PropTypes.string.isRequired,
}

const NotificationTitle = ({ notificationType, data }) => {
  const Notification = useMemo(() => getNotification(notificationType), [notificationType])

  return <Notification.Title notificationData={data} />
}

NotificationTitle.propTypes = {
  data: PropTypes.shape({}).isRequired,
  notificationType: PropTypes.string.isRequired,
}

const NotificationPreview = ({ notificationType, data }) => {
  const Notification = useMemo(() => getNotification(notificationType), [notificationType])

  return <Notification.Preview notificationData={data} />
}

NotificationPreview.propTypes = {
  data: PropTypes.shape({}).isRequired,
  notificationType: PropTypes.string.isRequired,
}

const NotificationIcon = ({ notificationType, data }) => {
  const Notification = useMemo(() => getNotification(notificationType), [notificationType])

  return <Notification.Icon notificationData={data} />
}

NotificationIcon.propTypes = {
  data: PropTypes.shape({}).isRequired,
  notificationType: PropTypes.string.isRequired,
}

const ttiNotification = {
  Content: NotificationContent,
  Title: NotificationTitle,
  Preview: NotificationPreview,
  Icon: NotificationIcon,
}

export default ttiNotification
