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
import { useParams } from 'react-router-dom'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import UserDataFormEdit from '@console/containers/user-data-form/edit'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getUser } from '@console/store/actions/users'

const UserManagementEdit = () => {
  const { userId } = useParams()

  useBreadcrumbs('admin-panel.user-management.edit', [
    {
      path: `./${userId}`,
      content: sharedMessages.edit,
    },
  ])

  return (
    <RequireRequest
      requestAction={getUser(userId, [
        'name',
        'primary_email_address',
        'state',
        'admin',
        'description',
      ])}
    >
      <UserDataFormEdit />
    </RequireRequest>
  )
}

export default UserManagementEdit
