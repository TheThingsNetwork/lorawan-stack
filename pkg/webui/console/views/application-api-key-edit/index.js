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
import { useSelector } from 'react-redux'

import { APPLICATION } from '@console/constants/entities'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import { ApiKeyEditForm } from '@console/containers/api-key-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getApiKey } from '@console/store/actions/api-keys'

import { selectApiKeyById } from '@console/store/selectors/api-keys'

const ApplicationApiKeyEditInner = () => {
  const { apiKeyId, appId } = useParams()

  useBreadcrumbs(
    'apps.single.api-keys.edit',
    <Breadcrumb
      path={`/applications/${appId}/api-keys/${apiKeyId}`}
      content={sharedMessages.edit}
    />,
  )

  return (
    <div className="container container--xl grid">
      <PageTitle title={sharedMessages.keyEdit} />
      <div className="item-12">
        <ApiKeyEditForm entity={APPLICATION} entityId={appId} />
      </div>
    </div>
  )
}

const ApplicationApiKeyEdit = () => {
  const { apiKeyId, appId } = useParams()

  // Check if API key still exists after possibly being deleted.
  const apiKey = useSelector(state => selectApiKeyById(state, apiKeyId))
  const hasApiKey = Boolean(apiKey)

  return (
    <RequireRequest requestAction={getApiKey('application', appId, apiKeyId)}>
      {hasApiKey && <ApplicationApiKeyEditInner />}
    </RequireRequest>
  )
}

export default ApplicationApiKeyEdit
