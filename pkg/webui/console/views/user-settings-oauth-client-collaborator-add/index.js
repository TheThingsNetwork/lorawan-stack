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
import { useParams } from 'react-router-dom'

import { CLIENT } from '@console/constants/entities'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import AccountCollaboratorsForm from '@console/containers/collaborators-form'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { mayViewOrEditClientCollaborators } from '@console/lib/feature-checks'

const OAuthClientCollaboratorAddInner = () => {
  const { clientId } = useParams()

  useBreadcrumbs(
    'user-settings.oauth-clients.single.collaborators.add',
    <Breadcrumb
      path={`/user-settings/oauth-clients/${clientId}/collaborators/add`}
      content={sharedMessages.add}
    />,
  )

  return (
    <div className="container container--xl grid">
      <div className="item-12">
        <AccountCollaboratorsForm entity={CLIENT} entityId={clientId} />
      </div>
    </div>
  )
}

const OAuthClientCollaboratorAdd = () => (
  <Require
    featureCheck={mayViewOrEditClientCollaborators}
    otherwise={{ redirect: '/user-settings/oauth-clients' }}
  >
    <OAuthClientCollaboratorAddInner />
  </Require>
)

export default OAuthClientCollaboratorAdd
