// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import { USER } from '@console/constants/entities'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import PageTitle from '@ttn-lw/components/page-title'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import { ApiKeyCreateForm } from '@console/containers/api-key-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getUsersRightsList } from '@console/store/actions/users'

import { selectUserId } from '@console/store/selectors/user'

const UserApiKeyAddInner = () => {
  const userId = useSelector(selectUserId)

  useBreadcrumbs(
    'user-settings.api-keys.add',
    <Breadcrumb path={`/user-settings/api-keys/add`} content={sharedMessages.add} />,
  )

  return (
    <div className="container container--xl grid">
      <PageTitle title={sharedMessages.addApiKey} />
      <div className="item-12">
        <ApiKeyCreateForm entity={USER} entityId={userId} />
      </div>
    </div>
  )
}

const UserApiKeyAdd = () => {
  const userId = useSelector(selectUserId)
  return (
    <RequireRequest requestAction={getUsersRightsList(userId)}>
      <UserApiKeyAddInner />
    </RequireRequest>
  )
}

export default UserApiKeyAdd
