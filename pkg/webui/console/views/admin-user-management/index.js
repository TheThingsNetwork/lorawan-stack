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
import { Routes, Route } from 'react-router-dom'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'
import ValidateRouteParam from '@ttn-lw/lib/components/validate-route-param'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import Require from '@console/lib/components/require'

import UserAdd from '@console/views/admin-user-management-add'
import UserEdit from '@console/views/admin-user-management-edit'
import InvitationSend from '@console/views/admin-user-management-invitation-send'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { userPathId as userPathIdRegexp } from '@ttn-lw/lib/regexp'

import { mayManageUsers } from '@console/lib/feature-checks'

import UserManagement from './admin-user-management'

const UserManagementRouter = () => {
  useBreadcrumbs(
    'admin-panel.user-management',
    <Breadcrumb path="/admin-panel/user-management" content={sharedMessages.userManagement} />,
  )
  return (
    <Require featureCheck={mayManageUsers} otherwise={{ redirect: '/' }}>
      <IntlHelmet title={sharedMessages.userManagement} />
      <Routes>
        <Route index Component={UserManagement} />
        <Route path="add" Component={UserAdd} />
        <Route path="invitations/add" Component={InvitationSend} />
        <Route
          path=":userId/*"
          element={<ValidateRouteParam check={{ userId: userPathIdRegexp }} Component={UserEdit} />}
        />
        <Route path="*" Component={GenericNotFound} />
      </Routes>
    </Require>
  )
}

export default UserManagementRouter
