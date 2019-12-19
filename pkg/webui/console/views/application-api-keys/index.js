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
import withFeatureRequirement from '../../lib/components/with-feature-requirement'
import ApplicationApiKeysList from '../application-api-keys-list'
import ApplicationApiKeyEdit from '../application-api-key-edit'
import ApplicationApiKeyAdd from '../application-api-key-add'

import { mayViewOrEditApplicationApiKeys } from '../../lib/feature-checks'
import { selectSelectedApplicationId } from '../../store/selectors/applications'
import PropTypes from '../../../lib/prop-types'

@connect(state => ({ appId: selectSelectedApplicationId(state) }))
@withFeatureRequirement(mayViewOrEditApplicationApiKeys, {
  redirect: ({ appId }) => `/applications/${appId}`,
})
@withBreadcrumb('apps.single.api-keys', ({ appId }) => (
  <Breadcrumb
    path={`/applications/${appId}/api-keys`}
    icon="api_keys"
    content={sharedMessages.apiKeys}
  />
))
export default class ApplicationAccess extends React.Component {
  static propTypes = {
    match: PropTypes.match.isRequired,
  }

  render() {
    const { match } = this.props
    return (
      <ErrorView ErrorComponent={SubViewError}>
        <Switch>
          <Route exact path={`${match.path}`} component={ApplicationApiKeysList} />
          <Route exact path={`${match.path}/add`} component={ApplicationApiKeyAdd} />
          <Route path={`${match.path}/:apiKeyId`} component={ApplicationApiKeyEdit} />
        </Switch>
      </ErrorView>
    )
  }
}
