// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import PageTitle from '@ttn-lw/components/page-title'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import NetworkInformationContainer from '@console/containers/network-information-container'
import DeploymentComponentStatus from '@console/containers/deployment-component-status'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getApplicationsList } from '@console/store/actions/applications'
import { getGatewaysList } from '@console/store/actions/gateways'
import { getUsersList } from '@console/store/actions/users'
import { getOrganizationsList } from '@console/store/actions/organizations'

const NetworkInformation = () => {
  useBreadcrumbs(
    'overview.admin-panel.network-information',
    <Breadcrumb
      path={`/admin-panel/network-information`}
      content={sharedMessages.networkInformation}
    />,
  )

  const requestActions = [
    getApplicationsList(),
    getGatewaysList(),
    getUsersList(),
    getOrganizationsList(),
  ]

  return (
    <>
      <RequireRequest requestAction={requestActions}>
        <PageTitle title={sharedMessages.networkInformation} className="panel-title mb-0" />
        <NetworkInformationContainer />
        <DeploymentComponentStatus />
      </RequireRequest>
    </>
  )
}

export default NetworkInformation
