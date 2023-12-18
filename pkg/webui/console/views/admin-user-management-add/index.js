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

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import UserDataFormAdd from '@console/containers/user-data-form/add'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getIsConfiguration } from '@console/store/actions/identity-server'

const UserManagementAdd = () => {
  useBreadcrumbs('admin-panel.user-management.add', [
    {
      path: `/admin-panel/user-management/add`,
      content: sharedMessages.add,
    },
  ])

  return (
    <RequireRequest requestAction={getIsConfiguration()}>
      <UserDataFormAdd />
    </RequireRequest>
  )
}

export default UserManagementAdd
