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

import sharedMessages from '../../../lib/shared-messages'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import ErrorView from '../../../lib/components/error-view'
import SubViewError from '../error/sub-view'
import GatewayApiKeysList from '../gateway-api-keys-list'
import GatewayApiKeyAdd from '../gateway-api-key-add'
import GatewayApiKeyEdit from '../gateway-api-key-edit'
import withFeatureRequirement from '../../lib/components/with-feature-requirement'

import { mayViewOrEditGatewayApiKeys } from '../../lib/feature-checks'
import { selectSelectedGatewayId } from '../../store/selectors/gateways'
import PropTypes from '../../../lib/prop-types'

@connect(state => ({ gtwId: selectSelectedGatewayId(state) }))
@withFeatureRequirement(mayViewOrEditGatewayApiKeys, {
  redirect: ({ gtwId }) => `/gateways/${gtwId}`,
})
@withBreadcrumb('gateways.single.api-keys', ({ gtwId }) => (
  <Breadcrumb
    path={`/gateways/${gtwId}/api-keys`}
    icon="api_keys"
    content={sharedMessages.apiKeys}
  />
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
