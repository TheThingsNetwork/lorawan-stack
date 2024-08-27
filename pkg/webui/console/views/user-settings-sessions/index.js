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

import { PAGE_SIZES } from '@ttn-lw/constants/page-sizes'

import PageTitle from '@ttn-lw/components/page-title'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import UserSessionsTable from '@console/containers/sessions-table'

import sharedMessages from '@ttn-lw/lib/shared-messages'

const SessionManagement = () => {
  useBreadcrumbs(
    'user-settings.session-management',
    <Breadcrumb path={`/user-settings/sessions`} content={sharedMessages.sessionManagement} />,
  )

  return (
    <div className="container container--xxl p-vert-cs-xs p-sides-0">
      <PageTitle title={sharedMessages.sessionManagement} hideHeading />
      <UserSessionsTable pageSize={PAGE_SIZES.REGULAR} />
    </div>
  )
}

export default SessionManagement
