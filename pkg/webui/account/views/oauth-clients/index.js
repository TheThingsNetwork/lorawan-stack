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
import { Switch, Route } from 'react-router-dom'

import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import OAuthClient from '@account/views/oauth-client'
import ClientsList from '@account/views/oauth-clients-list'
import OAuthClientAdd from '@account/views/oauth-client-add'

import PropTypes from '@ttn-lw/lib/prop-types'
import { pathId as pathIdRegexp } from '@ttn-lw/lib/regexp'

const OAuthClients = props => {
  const { path } = props.match

  return (
    <Switch>
      <Route exact path={`${path}`} component={ClientsList} />
      <Route exact path={`${path}/add`} component={OAuthClientAdd} />
      <Route path={`${path}/:clientId${pathIdRegexp}`} component={OAuthClient} />
      <NotFoundRoute />
    </Switch>
  )
}
OAuthClients.propTypes = {
  match: PropTypes.match.isRequired,
}
export default OAuthClients
