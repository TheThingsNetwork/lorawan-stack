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
import { createSelector } from 'reselect'

import { PAGE_SIZES } from '@ttn-lw/constants/page-sizes'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import CollaboratorsTable from '@console/containers/collaborators-table'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import {
  selectCollaborators,
  selectCollaboratorsTotalCount,
} from '@ttn-lw/lib/store/selectors/collaborators'
import { getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'

import { mayViewOrEditClientCollaborators } from '@console/lib/feature-checks'

const OAuthClientCollaboratorsList = () => {
  const { clientId } = useParams()

  const baseDataSelectors = createSelector(
    [selectCollaborators, selectCollaboratorsTotalCount],
    (collaborators, totalCount) => ({
      collaborators,
      totalCount,
    }),
  )

  const getCollaborators = React.useCallback(
    filters => getCollaboratorsList('client', clientId, filters),
    [clientId],
  )

  useBreadcrumbs(
    'user-settings.oauth-clients.single.collaborators',
    <Breadcrumb
      path={`/user-settings/oauth-clients/${clientId}/collaborators`}
      content={sharedMessages.collaborators}
    />,
  )

  return (
    <Require
      featureCheck={mayViewOrEditClientCollaborators}
      otherwise={{ redirect: '/user-settings/oauth-clients' }}
    >
      <div className="container container--xxl">
        <IntlHelmet title={sharedMessages.collaborators} />
        <CollaboratorsTable
          pageSize={PAGE_SIZES.REGULAR}
          baseDataSelector={baseDataSelectors}
          getItemsAction={getCollaborators}
        />
      </div>
    </Require>
  )
}

export default OAuthClientCollaboratorsList
