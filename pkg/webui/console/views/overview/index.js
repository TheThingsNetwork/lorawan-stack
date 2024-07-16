// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
import { Routes, Route } from 'react-router-dom'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import ValidateRouteParam from '@ttn-lw/lib/components/validate-route-param'

import Organizations from '@console/views/organizations'
import AdminPanel from '@console/views/admin-panel'
import User from '@console/views/user'
import Notifications from '@console/views/notifications'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { uuid as uuidRegexp } from '@ttn-lw/lib/regexp'

import Overview from './overview'

const OverviewRoutes = () => {
  useBreadcrumbs('overview', <Breadcrumb path="/" content={sharedMessages.overview} />)

  return (
    <Routes>
      <Route index Component={Overview} />
      <Route path="organizations/*" Component={Organizations} />
      <Route path="admin-panel/*" Component={AdminPanel} />
      <Route path="user/*" Component={User} />
      <Route path="notifications" Component={Notifications} />
      <Route
        path="notifications/:category?/:id?"
        Component={Notifications}
        element={
          <ValidateRouteParam
            check={{ category: /^inbox|archived$/, id: uuidRegexp }}
            Component={Notifications}
          />
        }
      />
    </Routes>
  )
}
export default OverviewRoutes
