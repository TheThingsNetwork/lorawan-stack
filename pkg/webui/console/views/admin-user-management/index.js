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

import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import BreadcrumbView from '../../../lib/components/breadcrumb-view'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import PropTypes from '../../../lib/prop-types'
import sharedMessages from '../../../lib/shared-messages'
import IntlHelmet from '../../../lib/components/intl-helmet'
import NotFoundRoute from '../../../lib/components/not-found-route'
import withFeatureRequirement from '../../lib/components/with-feature-requirement'

import { mayManageUsers } from '../../lib/feature-checks'
import UserEdit from '../admin-user-management-edit'
import UserManagement from './admin-user-management'

@withFeatureRequirement(mayManageUsers, { redirect: '/' })
@withBreadcrumb('admin.user-management', function() {
  return (
    <Breadcrumb
      path={'/admin/user-management'}
      icon="user_management"
      content={sharedMessages.userManagement}
    />
  )
})
export default class UserManagementRouter extends Component {
  static propTypes = {
    match: PropTypes.match.isRequired,
  }
  render() {
    const { match } = this.props
    return (
      <BreadcrumbView>
        <IntlHelmet title={sharedMessages.userManagement} />
        <Switch>
          <Route exact path={`${match.path}`} component={UserManagement} />
          <Route path={`${match.path}/:userId`} component={UserEdit} />
          <NotFoundRoute />
        </Switch>
      </BreadcrumbView>
    )
  }
}
