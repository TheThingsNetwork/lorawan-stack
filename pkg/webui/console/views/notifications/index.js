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
import { useParams } from 'react-router-dom'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Breadcrumbs from '@ttn-lw/components/breadcrumbs'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import NotificationsContainer from '@console/containers/notifications'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  getArchivedNotifications,
  getInboxNotifications,
} from '@console/store/actions/notifications'

const NotificationsView = () => {
  const { category } = useParams()
  useBreadcrumbs(
    'notifications',
    <Breadcrumb path="/notifications" content={sharedMessages.notifications} />,
  )

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
      <Breadcrumbs />
      <NotificationsContainer />
    </RequireRequest>
  )
}

export default NotificationsView
