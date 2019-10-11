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

import React, { Component } from 'react'
import { connect } from 'react-redux'
import { Switch, Route } from 'react-router'

import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import sharedMessages from '../../../lib/shared-messages'
import NotFoundRoute from '../../../lib/components/not-found-route'
import { selectSelectedApplicationId } from '../../store/selectors/applications'
import PropTypes from '../../../lib/prop-types'
import DeviceAddSingle from '../device-add-single'
import DeviceAddBulk from '../device-add-bulk'

@withBreadcrumb('devices.add', function(props) {
  const { appId } = props.match.params
  return (
    <Breadcrumb
      path={`/applications/${appId}/devices/add`}
      icon="add"
      content={sharedMessages.add}
    />
  )
})
export default class DeviceAdd extends Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
  }

  render() {
    const { appId } = this.props
    const basePath = `/applications/${appId}/devices/add`

    return (
      <Switch>
        <Route exact path={basePath} component={DeviceAddSingle} />
        <Route exact path={`${basePath}/bulk`} component={DeviceAddBulk} />
        <NotFoundRoute />
      </Switch>
    )
  }
}
