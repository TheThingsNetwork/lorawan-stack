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
import { useSelector } from 'react-redux'
import { defineMessages } from '@formatjs/intl'

import Breadcrumbs from '@ttn-lw/components/breadcrumbs'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import PanelView from '@console/components/panel-view'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import UserManagement from '@console/views/admin-user-management'
import PacketBrokerRouter from '@console/views/admin-packet-broker'
import NetworkInformation from '@console/views/admin-panel-network-information'

import {
  checkFromState,
  mayConfigurePacketBroker,
  mayManageUsers,
  mayPerformAdminActions,
} from '@console/lib/feature-checks'

const m = defineMessages({
  adminPanel: 'Admin panel',
  networkInformation: 'Network information',
  userManagement: 'User management',
  globalNetworkSettings: 'Global network settings',
  peeringSettings: 'Peering settings',
})

const AdminPanel = () => {
  useBreadcrumbs('admin-panel', <Breadcrumb path="/admin-panel" content={m.adminPanel} />)
  const showUserManagement = useSelector(state => checkFromState(mayManageUsers, state))
  const showPacketBroker = useSelector(state => checkFromState(mayConfigurePacketBroker, state))

  return (
    <>
      <Breadcrumbs />
      <IntlHelmet title={m.adminPanel} />
      <PanelView>
        <PanelView.Item
          title={m.networkInformation}
          icon="view_compact"
          path="/network-information"
          component={NetworkInformation}
          exact
        />
        {showUserManagement && (
          <PanelView.Item
            title={m.userManagement}
            icon="user_management"
            path="/user-management"
            component={UserManagement}
            condition={showUserManagement}
          />
        )}
        {showPacketBroker && (
          <PanelView.Item
            title={m.peeringSettings}
            icon="packet_broker"
            path="/packet-broker"
            component={PacketBrokerRouter}
            condition={showPacketBroker}
          />
        )}
      </PanelView>
    </>
  )
}

export default withFeatureRequirement(mayPerformAdminActions, { redirect: '/' })(AdminPanel)
