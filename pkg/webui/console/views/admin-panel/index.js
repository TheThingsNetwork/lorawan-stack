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
import { Navigate, Route, Routes } from 'react-router-dom'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import Require from '@console/lib/components/require'

import UserManagement from '@console/views/admin-user-management'
import PacketBrokerRouter from '@console/views/admin-packet-broker'
import NetworkInformation from '@console/views/admin-panel-network-information'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  checkFromState,
  mayConfigurePacketBroker,
  mayManageUsers,
  mayPerformAdminActions,
} from '@console/lib/feature-checks'

const AdminPanel = () => {
  useBreadcrumbs(
    'overview.admin-panel',
    <Breadcrumb path="/admin-panel" content={sharedMessages.adminPanel} />,
  )
  const showUserManagement = useSelector(state => checkFromState(mayManageUsers, state))
  const showPacketBroker = useSelector(state => checkFromState(mayConfigurePacketBroker, state))

  return (
    <Require featureCheck={mayPerformAdminActions} otherwise={{ redirect: '/' }}>
      <IntlHelmet title={sharedMessages.adminPanel} />
      <Routes>
        <Route index element={<Navigate to="network-information" />} />
        <Route path="network-information" element={<NetworkInformation />} />
        {showUserManagement && <Route path="user-management/*" element={<UserManagement />} />}
        {showPacketBroker && <Route path="packet-broker/*" element={<PacketBrokerRouter />} />}
      </Routes>
    </Require>
  )
}

export default AdminPanel
