// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
import { Routes, Route, Navigate } from 'react-router-dom'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'

import ProfileSettings from '@console/views/user-settings-profile'
import UserApiKeys from '@console/views/user-api-keys'

import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import ChangePassword from '../user-settings-password'

const UserSettings = () => {
  useBreadcrumbs(
    'user-settings',
    <Breadcrumb path={`/user-settings`} content={sharedMessages.userSettings} />,
  )
  return (
    <>
      <IntlHelmet titleTemplate={`%s - User settings - ${selectApplicationSiteName()}`} />
      <Routes>
        <Route path="profile" Component={ProfileSettings} />
        <Route path="password" Component={ChangePassword} />
        <Route path="api-keys/*" Component={UserApiKeys} />
        <Route index element={<Navigate to="profile" />} />
        <Route path="*" Component={GenericNotFound} />
      </Routes>
    </>
  )
}

export default UserSettings
