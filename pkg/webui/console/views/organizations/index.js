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

import OrganizationsList from '../organizations-list'
import OrganizationAdd from '../organization-add'
import Organization from '../organization'

import sharedMessages from '../../../lib/shared-messages'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import withFeatureRequirement from '../../lib/components/with-feature-requirement'

import { mayViewOrganizationsOfUser } from '../../lib/feature-checks'

@withFeatureRequirement(mayViewOrganizationsOfUser, { redirect: '/' })
@withBreadcrumb('orgs', function(props) {
  return <Breadcrumb path="/organizations" content={sharedMessages.organizations} />
})
class Organizations extends React.Component {
  render() {
    const { path } = this.props.match
    return (
      <Switch>
        <Route exact path={`${path}`} component={OrganizationsList} />
        <Route exact path={`${path}/add`} component={OrganizationAdd} />
        <Route path={`${path}/:orgId`} component={Organization} />
      </Switch>
    )
  }
}

export default Organizations
