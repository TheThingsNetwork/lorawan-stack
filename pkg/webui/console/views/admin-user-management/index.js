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

import React, { Component } from 'react'
import { Switch, Route } from 'react-router'

import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Breadcrumbs from '@ttn-lw/components/breadcrumbs'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import UserAdd from '@console/views/admin-user-management-add'
import UserEdit from '@console/views/admin-user-management-edit'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { mayManageUsers } from '@console/lib/feature-checks'

import UserManagement from './admin-user-management'

@withFeatureRequirement(mayManageUsers, { redirect: '/' })
@withBreadcrumb('admin.user-management', function () {
  return <Breadcrumb path={'/admin/user-management'} content={sharedMessages.userManagement} />
})
export default class UserManagementRouter extends Component {
  static propTypes = {
    match: PropTypes.match.isRequired,
  }
  render() {
    const { match } = this.props
    return (
      <React.Fragment>
        <Breadcrumbs />
        <IntlHelmet title={sharedMessages.userManagement} />
        <Switch>
          <Route exact path={`${match.path}`} component={UserManagement} />
          <Route path={`${match.path}/add`} component={UserAdd} />
          <Route path={`${match.path}/:userId`} component={UserEdit} />
          <NotFoundRoute />
        </Switch>
      </React.Fragment>
    )
  }
}
