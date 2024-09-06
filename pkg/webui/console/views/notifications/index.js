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

import React, { useEffect, useMemo } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import NotificationsContainer from '@console/containers/notifications'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  getArchivedNotifications,
  getInboxNotifications,
} from '@console/store/actions/notifications'

const NotificationsView = () => {
  const { category } = useParams()
  const navigate = useNavigate()

  useBreadcrumbs(
    'overview.notifications',
    <Breadcrumb path="/notifications" content={sharedMessages.notifications} />,
  )

  useEffect(() => {
    if (category !== 'archived' && category !== 'inbox') {
      navigate(`/notifications/inbox`, { replace: true })
    }
  }, [category, navigate])

  const action = useMemo(
    () =>
      (category === 'archived' ? getArchivedNotifications : getInboxNotifications)({
        page: 1,
        limit: 25,
      }),
    [category],
  )

  return (
    <RequireRequest requestAction={action} requestOnChange>
      <NotificationsContainer />
    </RequireRequest>
  )
}

export default NotificationsView
