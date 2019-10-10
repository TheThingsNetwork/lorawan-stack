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

import ApplicationAdd from '../application-add'
import ApplicationsList from '../applications-list'
import Application from '../application'

import sharedMessages from '../../../lib/shared-messages'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'

@withBreadcrumb('apps', function(props) {
  return <Breadcrumb path="/applications" content={sharedMessages.applications} />
})
export default class Applications extends React.Component {
  render() {
    const { path } = this.props.match
    return (
      <Switch>
        <Route exact path={`${path}`} component={ApplicationsList} />
        <Route exact path={`${path}/add`} component={ApplicationAdd} />
        <Route path={`${path}/:appId`} component={Application} />
      </Switch>
    )
  }
}
