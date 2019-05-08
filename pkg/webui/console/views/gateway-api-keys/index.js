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

import sharedMessages from '../../../lib/shared-messages'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'

import GatewayApiKeysList from '../gateway-api-keys-list'

@withBreadcrumb('gateways.single.api-keys', function (props) {
  const gtwId = props.match.params.gtwId

  return (
    <Breadcrumb
      path={`/console/gateways/${gtwId}/api-keys`}
      icon="api_keys"
      content={sharedMessages.apiKeys}
    />
  )
})
export default class GatewayApiKeys extends React.Component {

  render () {
    const { match } = this.props

    return (
      <Switch>
        <Route exact path={`${match.path}`} component={GatewayApiKeysList} />
      </Switch>
    )
  }
}
