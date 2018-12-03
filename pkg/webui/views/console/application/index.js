// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
import Message from '../../../lib/components/message'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import { withSideNavigation } from '../../../components/navigation/side/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import Spinner from '../../../components/spinner'

import ApplicationOverview from '../application-overview'
import ApplicationGeneralSettings from '../application-general-settings'

import { getApplication } from '../../../actions/application'

@connect(function ({ application }, props) {
  return {
    appId: props.match.params.appId,
    fetching: application.fetching,
    error: application.error,
  }
})
@withSideNavigation(function (props) {
  const matchedUrl = props.match.url

  return {
    header: { title: props.appId, icon: 'application' },
    entries: [
      {
        title: sharedMessages.overview,
        path: matchedUrl,
        icon: 'overview',
      },
      {
        title: sharedMessages.devices,
        path: `${matchedUrl}/devices`,
        icon: 'devices',
      },
      {
        title: sharedMessages.data,
        path: `${matchedUrl}/data`,
        icon: 'data',
      },
      {
        title: sharedMessages.payload_formats,
        path: `${matchedUrl}/payload-formats`,
        icon: 'payload_formats',
      },
      {
        title: sharedMessages.integrations,
        path: `${matchedUrl}/integrations`,
        icon: 'integration',
      },
      {
        title: sharedMessages.collaborators,
        path: `${matchedUrl}/collaborators`,
        icon: 'organization',
      },
      {
        title: sharedMessages.access,
        path: `${matchedUrl}/access`,
        icon: 'access',
      },
      {
        title: sharedMessages.generalSettings,
        path: `${matchedUrl}/general-settings`,
        icon: 'general_settings',
      },
    ],
  }
})
@withBreadcrumb('apps.single', function (props) {
  const { appId } = props
  return (
    <Breadcrumb
      path={`/console/applications/${appId}`}
      icon="application"
      content={appId}
    />
  )
})
export default class Application extends React.Component {

  componentDidMount () {
    const { dispatch, appId } = this.props

    dispatch(getApplication(appId))
  }

  render () {
    const { fetching, error, match } = this.props

    if (fetching) {
      return (
        <Spinner center>
          <Message content={sharedMessages.loading} />
        </Spinner>
      )
    }

    // show any application fetching error, e.g. not found, not rights, etc
    if (error) {
      return 'ERROR'
    }

    return (
      <Switch>
        <Route exact path={`${match.path}`} component={ApplicationOverview} />
        <Route path={`${match.path}/general-settings`} component={ApplicationGeneralSettings} />
      </Switch>
    )
  }
}
