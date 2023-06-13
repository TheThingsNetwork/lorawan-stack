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
import { Routes, Route } from 'react-router-dom'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'

import Require from '@console/lib/components/require'

import UserManagement from '@console/views/admin-user-management'
import PacketBroker from '@console/views/admin-packet-broker'

import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'

import { mayPerformAdminActions } from '@console/lib/feature-checks'

const AdminView = () => (
  <Require featureCheck={mayPerformAdminActions} otherwise={{ redirect: '/' }}>
    <IntlHelmet titleTemplate={`%s - Admin Configurations - ${selectApplicationSiteName()}`} />
    <Routes>
      <Route path="user-management/*" Component={UserManagement} />
      <Route path="packet-broker/*" Component={PacketBroker} />
      <Route path="*" Component={GenericNotFound} />
    </Routes>
  </Require>
)

export default AdminView
