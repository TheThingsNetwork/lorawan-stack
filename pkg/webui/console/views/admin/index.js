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
import { Switch, Route } from 'react-router'

import PropTypes from '../../../lib/prop-types'
import { selectApplicationSiteName } from '../../../lib/selectors/env'
import IntlHelmet from '../../../lib/components/intl-helmet'
import NotFoundRoute from '../../../lib/components/not-found-route'
import withFeatureRequirement from '../../lib/components/with-feature-requirement'

import UserManagement from '../admin-user-management'
import { mayPerformAdminActions } from '../../lib/feature-checks'

const AdminView = ({ match }) => (
  <React.Fragment>
    <IntlHelmet titleTemplate={`%s - Admin Configurations - ${selectApplicationSiteName()}`} />
    <Switch>
      <Route path={`${match.path}/user-management`} component={UserManagement} />
      <NotFoundRoute />
    </Switch>
  </React.Fragment>
)

AdminView.propTypes = {
  match: PropTypes.match.isRequired,
}

export default withFeatureRequirement(mayPerformAdminActions, { redirect: '/' })(AdminView)
