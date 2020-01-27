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
import ApplicationPubsubsList from '../application-integrations-pubsubs-list'
import ApplicationPubsubAdd from '../application-integrations-pubsub-add'
import ApplicationPubsubEdit from '../application-integrations-pubsub-edit'
import withFeatureRequirement from '../../lib/components/with-feature-requirement'

import { mayViewApplicationEvents } from '../../lib/feature-checks'
import { selectSelectedApplicationId } from '../../store/selectors/applications'
import PropTypes from '../../../lib/prop-types'

@connect(state => ({ appId: selectSelectedApplicationId(state) }))
@withFeatureRequirement(mayViewApplicationEvents, {
  redirect: ({ appId }) => `/applications/${appId}`,
})
@withBreadcrumb('apps.single.integrations.pubsubs', ({ appId }) => (
  <Breadcrumb
    path={`/applications/${appId}/integrations/pubsubs`}
    content={sharedMessages.pubsubs}
  />
))
export default class ApplicationPubsubs extends React.Component {
  static propTypes = {
    match: PropTypes.match.isRequired,
  }

  render() {
    const { match } = this.props

    return (
      <ErrorView ErrorComponent={SubViewError}>
        <Switch>
          <Route exact path={`${match.path}`} component={ApplicationPubsubsList} />
          <Route exact path={`${match.path}/add`} component={ApplicationPubsubAdd} />
          <Route path={`${match.path}/:pubsubId`} component={ApplicationPubsubEdit} />
        </Switch>
      </ErrorView>
    )
  }
}
