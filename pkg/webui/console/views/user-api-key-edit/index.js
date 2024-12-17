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
import { useParams } from 'react-router-dom'

import { USER } from '@console/constants/entities'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import PageTitle from '@ttn-lw/components/page-title'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import { ApiKeyEditForm } from '@console/containers/api-key-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getUsersRightsList } from '@console/store/actions/users'
import { getApiKey } from '@console/store/actions/api-keys'

import { selectUserId } from '@console/store/selectors/user'
import { selectApiKeyById } from '@console/store/selectors/api-keys'

const UserApiKeyEditInner = () => {
  const userId = useSelector(selectUserId)
  const { apiKeyId } = useParams()

  useBreadcrumbs(
    'user-settings.api-keys.edit',
    <Breadcrumb path={`/user-settings/api-keys/edit/${apiKeyId}`} content={sharedMessages.edit} />,
  )

  return (
    <div className="container container--xl grid">
      <PageTitle title={sharedMessages.keyEdit} />
      <div className="item-12">
        <ApiKeyEditForm entity={USER} entityId={userId} />
      </div>
    </div>
  )
}

const UserApiKeyEdit = () => {
  const userId = useSelector(selectUserId)
  const { apiKeyId } = useParams()

  // Check if API key still exists after possibly being deleted.
  const apiKey = useSelector(state => selectApiKeyById(state, apiKeyId))
  const hasApiKey = Boolean(apiKey)

  return (
    <RequireRequest
      requestAction={[getUsersRightsList(userId), getApiKey('users', userId, apiKeyId)]}
    >
      {hasApiKey && <UserApiKeyEditInner />}
    </RequireRequest>
  )
}

export default UserApiKeyEdit
