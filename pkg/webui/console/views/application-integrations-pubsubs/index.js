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

import ApplicationPubsubEdit from '@console/views/application-integrations-pubsub-edit'
import ApplicationPubsubAdd from '@console/views/application-integrations-pubsub-add'
import ApplicationPubsubsList from '@console/views/application-integrations-pubsubs-list'
import SubViewError from '@console/views/error/sub-view'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayViewApplicationEvents } from '@console/lib/feature-checks'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'

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
