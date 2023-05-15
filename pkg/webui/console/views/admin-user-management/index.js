// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import Require from '@console/lib/components/require'

import UserAdd from '@console/views/admin-user-management-add'
import UserEdit from '@console/views/admin-user-management-edit'
import InvitationSend from '@console/views/invitation-send'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import { userPathId as userPathIdRegexp } from '@ttn-lw/lib/regexp'

import { mayManageUsers } from '@console/lib/feature-checks'

import UserManagement from './admin-user-management'

const UserManagementRouter = ({ match }) => {
  useBreadcrumbs(
    'admin-panel.user-management',
    <Breadcrumb path={'/admin-panel/user-management'} content={sharedMessages.userManagement} />,
  )
  return (
    <Require featureCheck={mayManageUsers} otherwise={{ redirect: '/' }}>
      <IntlHelmet title={sharedMessages.userManagement} />
      <Routes>
        <Route exact path={`${match.path}`} component={UserManagement} />
        <Route path={`${match.path}/add`} component={UserAdd} />
        <Route path={`${match.path}/invitations/add`} component={InvitationSend} />
        <Route path={`${match.path}/:userId${userPathIdRegexp}`} component={UserEdit} sensitive />
        <NotFoundRoute />
      </Routes>
    </Require>
  )
}

UserManagementRouter.propTypes = {
  match: PropTypes.match.isRequired,
}

export default UserManagementRouter
