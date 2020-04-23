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
import { connect } from 'react-redux'
import { Switch, Route } from 'react-router'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'

import ErrorView from '@ttn-lw/lib/components/error-view'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import GatewayApiKeyEdit from '@console/views/gateway-api-key-edit'
import GatewayApiKeyAdd from '@console/views/gateway-api-key-add'
import GatewayApiKeysList from '@console/views/gateway-api-keys-list'
import SubViewError from '@console/views/error/sub-view'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayViewOrEditGatewayApiKeys } from '@console/lib/feature-checks'

import { selectSelectedGatewayId } from '@console/store/selectors/gateways'

@connect(state => ({ gtwId: selectSelectedGatewayId(state) }))
@withFeatureRequirement(mayViewOrEditGatewayApiKeys, {
  redirect: ({ gtwId }) => `/gateways/${gtwId}`,
})
@withBreadcrumb('gateways.single.api-keys', ({ gtwId }) => (
  <Breadcrumb path={`/gateways/${gtwId}/api-keys`} content={sharedMessages.apiKeys} />
))
export default class GatewayApiKeys extends React.Component {
  static propTypes = {
    match: PropTypes.match.isRequired,
  }

  render() {
    const { match } = this.props

    return (
      <ErrorView ErrorComponent={SubViewError}>
        <Switch>
          <Route exact path={`${match.path}`} component={GatewayApiKeysList} />
          <Route exact path={`${match.path}/add`} component={GatewayApiKeyAdd} />
          <Route path={`${match.path}/:apiKeyId`} component={GatewayApiKeyEdit} />
        </Switch>
      </ErrorView>
    )
  }
}
