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
import { Switch, Route } from 'react-router-dom'

import GatewaysList from '../gateways-list'
import GatewayAdd from '../gateway-add'
import Gateway from '../gateway'

import sharedMessages from '../../../lib/shared-messages'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'

@withBreadcrumb('gateways', function (props) {
  return (
    <Breadcrumb
      path="/gateways"
      content={sharedMessages.gateways}
    />
  )
})
export default class Gateways extends React.Component {

  render () {
    const { path } = this.props.match

    return (
      <Switch>
        <Route exact path={`${path}`} component={GatewaysList} />
        <Route path={`${path}/add`} component={GatewayAdd} />
        <Route path={`${path}/:gtwId`} component={Gateway} />
      </Switch>
    )
  }
}
