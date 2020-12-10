// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import Breadcrumbs from '@ttn-lw/components/breadcrumbs'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import UserApiKeys from '@console/views/user-api-keys'

import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'
import PropTypes from '@ttn-lw/lib/prop-types'

const UserView = ({ match }) => (
  <>
    <Breadcrumbs />
    <IntlHelmet titleTemplate={`%s - User - ${selectApplicationSiteName()}`} />
    <Switch>
      <Route path={`${match.path}/api-keys`} component={UserApiKeys} />
      <NotFoundRoute />
    </Switch>
  </>
)

UserView.propTypes = {
  match: PropTypes.match.isRequired,
}

export default UserView
