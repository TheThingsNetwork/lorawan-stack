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

import { APPLICATION } from '@console/constants/entities'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import { ApiKeyCreateForm } from '@console/containers/api-key-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'

const ApplicationApiKeyAdd = () => {
  const { appId } = useParams()

  useBreadcrumbs(
    'apps.single.api-keys.add',
    <Breadcrumb path={`/applications/${appId}/api-keys/add`} content={sharedMessages.add} />,
  )

  return (
    <div className="container container--xl grid">
      <PageTitle title={sharedMessages.addApiKey} />
      <div className="item-12">
        <ApiKeyCreateForm entityId={appId} entity={APPLICATION} />
      </div>
    </div>
  )
}

export default ApplicationApiKeyAdd
